package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/agent-yandex/dating-bot/internal/db"
	"github.com/agent-yandex/dating-bot/internal/deps"
	"github.com/agent-yandex/dating-bot/internal/tg/states"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

const (
	searchResultsTTL = 30 * time.Minute
)

type CallbackHandler struct {
	stateMgr *states.Manager
	db       *deps.DB
	redis    *redis.Client
	logger   *zap.Logger
}

func NewCallbackHandler(stateMgr *states.Manager, db *deps.DB, redis *redis.Client, logger *zap.Logger) *CallbackHandler {
	return &CallbackHandler{
		stateMgr: stateMgr,
		db:       db,
		redis:    redis,
		logger:   logger,
	}
}

func (h *CallbackHandler) RegisterCallbacks(d *ext.Dispatcher) {
	d.AddHandler(handlers.NewCallback(nil, h.callbackHandler))
}

func (h *CallbackHandler) callbackHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.CallbackQuery == nil {
		return nil
	}

	userID := ctx.CallbackQuery.From.Id

	defer func() {
		_, _ = b.AnswerCallbackQuery(ctx.CallbackQuery.Id, nil)
	}()

	switch {
	case strings.HasPrefix(ctx.CallbackQuery.Data, "like:"):
		var profileID int64
		fmt.Sscanf(ctx.CallbackQuery.Data, "like:%d", &profileID)
		return h.handleLike(b, ctx, userID, profileID)

	case strings.HasPrefix(ctx.CallbackQuery.Data, "dislike:"):
		var profileID int64
		fmt.Sscanf(ctx.CallbackQuery.Data, "dislike:%d", &profileID)
		return h.handleDislike(b, ctx, userID, profileID)

	case strings.HasPrefix(ctx.CallbackQuery.Data, "like_like:"):
		var profileID int64
		fmt.Sscanf(ctx.CallbackQuery.Data, "like_like:%d", &profileID)
		return h.handleLikeFromLikes(b, ctx, userID, profileID)

	case strings.HasPrefix(ctx.CallbackQuery.Data, "dislike_like:"):
		var profileID int64
		fmt.Sscanf(ctx.CallbackQuery.Data, "dislike_like:%d", &profileID)
		return h.handleDislikeFromLikes(b, ctx, userID, profileID)

	default:
		return nil
	}
}

func (h *CallbackHandler) handleLike(b *gotgbot.Bot, ctx *ext.Context, userID, profileID int64) error {
	chatID := ctx.CallbackQuery.Message.GetChat().Id
	messageID := ctx.CallbackQuery.Message.GetMessageId()

	like := &db.Like{
		FromUserID: userID,
		ToUserID:   profileID,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
	}

	_, err := h.db.Likes.Insert(context.Background(), like)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			h.logger.Info("User already liked profile",
				zap.Int64("from_user_id", userID),
				zap.Int64("to_user_id", profileID))
			_, _ = b.SendMessage(chatID, "Вы уже лайкнули этого пользователя.", nil)
		} else {
			h.logger.Error("Failed to insert like",
				zap.Int64("from_user_id", userID),
				zap.Int64("to_user_id", profileID),
				zap.Error(err))
			_, _ = b.SendMessage(chatID, "Произошла ошибка при сохранении лайка.", nil)
			return err
		}
	} else {
		// Обновляем рейтинг пользователя, получившего лайк
		if err := h.db.Users.UpdateRating(context.Background(), profileID); err != nil {
			h.logger.Error("Failed to update rating",
				zap.Int64("user_id", profileID),
				zap.Error(err))
			// Продолжаем, так как ошибка не критична для пользовательского опыта
		}
	}

	mutualLikes, err := h.db.Likes.GetAllByToUserID(context.Background(), userID)
	if err != nil {
		h.logger.Error("Failed to check mutual likes",
			zap.Int64("user_id", userID),
			zap.Error(err))
	} else {
		for _, l := range mutualLikes {
			if l.FromUserID == profileID {
				// Удаляем лайк от profileID к userID
				err = h.db.Likes.DeleteByIDs(context.Background(), profileID, userID)
				if err != nil {
					h.logger.Error("Failed to delete like",
						zap.Int64("from_user_id", profileID),
						zap.Int64("to_user_id", userID),
						zap.Error(err))
					// Продолжаем, так как лайк уже записан
				}

				// Удаляем лайк от userID к profileID
				err = h.db.Likes.DeleteByIDs(context.Background(), userID, profileID)
				if err != nil {
					h.logger.Error("Failed to delete mutual like",
						zap.Int64("from_user_id", userID),
						zap.Int64("to_user_id", profileID),
						zap.Error(err))
					// Продолжаем, так как первый лайк уже удалён
				}

				// Обновляем рейтинг обоих пользователей
				for _, uid := range []int64{userID, profileID} {
					if err := h.db.Users.UpdateRating(context.Background(), uid); err != nil {
						h.logger.Error("Failed to update rating",
							zap.Int64("user_id", uid),
							zap.Error(err))
						// Продолжаем, так как ошибка не критична
					}
				}

				// Очищаем кэш Redis для списка лайков обоих пользователей
				ctx := context.Background()
				for _, uid := range []int64{userID, profileID} {
					keys, err := h.redis.Keys(ctx, fmt.Sprintf("likes:%d:*", uid)).Result()
					if err != nil {
						h.logger.Error("Failed to get Redis keys for likes cache",
							zap.Int64("user_id", uid),
							zap.Error(err))
					} else if len(keys) > 0 {
						if err := h.redis.Del(ctx, keys...).Err(); err != nil {
							h.logger.Error("Failed to delete Redis keys for likes cache",
								zap.Int64("user_id", uid),
								zap.Error(err))
						} else {
							h.logger.Info("Cleared likes cache",
								zap.Int64("user_id", uid),
								zap.Strings("keys", keys))
						}
					}
				}

				// Отправляем уведомления о взаимном лайке
				err = h.notifyMutualLikeWithLinks(b, userID, profileID)
				if err != nil {
					h.logger.Error("Failed to notify mutual like",
						zap.Int64("user1_id", userID),
						zap.Int64("user2_id", profileID),
						zap.Error(err))
				}
				break
			}
		}
	}

	_, err = b.DeleteMessage(chatID, messageID, nil)
	if err != nil {
		h.logger.Warn("Failed to delete message",
			zap.Int64("chat_id", chatID),
			zap.Int64("message_id", messageID),
			zap.Error(err))
	}

	currentIndex := h.stateMgr.GetCurrentIndex(userID) + 1
	offset := uint64(currentIndex/50) * 50
	profiles, err := h.getSearchResults(userID, offset)
	if err != nil {
		h.logger.Error("Failed to get next profiles",
			zap.Int64("user_id", userID),
			zap.Error(err))
		_, _ = b.SendMessage(chatID, "Произошла ошибка при загрузке следующей анкеты.", nil)
		return err
	}

	h.stateMgr.SetCurrentIndex(userID, currentIndex)

	return h.sendProfile(b, chatID, userID, profiles, currentIndex%50)
}

func (h *CallbackHandler) handleDislike(b *gotgbot.Bot, ctx *ext.Context, userID, profileID int64) error {
	chatID := ctx.CallbackQuery.Message.GetChat().Id
	messageID := ctx.CallbackQuery.Message.GetMessageId()

	h.logger.Info("User disliked profile",
		zap.Int64("user_id", userID),
		zap.Int64("profile_id", profileID))

	_, err := b.DeleteMessage(chatID, messageID, nil)
	if err != nil {
		h.logger.Warn("Failed to delete message",
			zap.Int64("chat_id", chatID),
			zap.Int64("message_id", messageID),
			zap.Error(err))
	}

	currentIndex := h.stateMgr.GetCurrentIndex(userID) + 1
	offset := uint64(currentIndex/50) * 50
	profiles, err := h.getSearchResults(userID, offset)
	if err != nil {
		h.logger.Error("Failed to get next profiles",
			zap.Int64("user_id", userID),
			zap.Error(err))
		_, _ = b.SendMessage(chatID, "Произошла ошибка при загрузке следующей анкеты.", nil)
		return err
	}

	h.stateMgr.SetCurrentIndex(userID, currentIndex)

	return h.sendProfile(b, chatID, userID, profiles, currentIndex%50)
}

func (h *CallbackHandler) handleLikeFromLikes(b *gotgbot.Bot, ctx *ext.Context, userID, profileID int64) error {
	chatID := ctx.CallbackQuery.Message.GetChat().Id
	messageID := ctx.CallbackQuery.Message.GetMessageId()

	// Записываем лайк текущего пользователя
	like := &db.Like{
		FromUserID: userID,
		ToUserID:   profileID,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
	}

	_, err := h.db.Likes.Insert(context.Background(), like)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			h.logger.Info("User already liked profile",
				zap.Int64("from_user_id", userID),
				zap.Int64("to_user_id", profileID))
			_, _ = b.SendMessage(chatID, "Вы уже лайкнули этого пользователя.", nil)
		} else {
			h.logger.Error("Failed to insert like",
				zap.Int64("from_user_id", userID),
				zap.Int64("to_user_id", profileID),
				zap.Error(err))
			_, _ = b.SendMessage(chatID, "Произошла ошибка при сохранении лайка.", nil)
			return err
		}
	} else {
		// Обновляем рейтинг пользователя, получившего лайк
		if err := h.db.Users.UpdateRating(context.Background(), profileID); err != nil {
			h.logger.Error("Failed to update rating",
				zap.Int64("user_id", profileID),
				zap.Error(err))
			// Продолжаем, так как ошибка не критична
		}

		// Удаляем лайк от profileID к userID
		err = h.db.Likes.DeleteByIDs(context.Background(), profileID, userID)
		if err != nil {
			h.logger.Error("Failed to delete like",
				zap.Int64("from_user_id", profileID),
				zap.Int64("to_user_id", userID),
				zap.Error(err))
			// Продолжаем, так как лайк уже записан
		}

		// Удаляем лайк от userID к profileID
		err = h.db.Likes.DeleteByIDs(context.Background(), userID, profileID)
		if err != nil {
			h.logger.Error("Failed to delete mutual like",
				zap.Int64("from_user_id", userID),
				zap.Int64("to_user_id", profileID),
				zap.Error(err))
			// Продолжаем, так как первый лайк уже удалён
		}

		// Обновляем рейтинг обоих пользователей
		for _, uid := range []int64{userID, profileID} {
			if err := h.db.Users.UpdateRating(context.Background(), uid); err != nil {
				h.logger.Error("Failed to update rating",
					zap.Int64("user_id", uid),
					zap.Error(err))
				// Продолжаем, так как ошибка не критична
			}
		}

		// Очищаем кэш Redis для списка лайков обоих пользователей
		ctx := context.Background()
		for _, uid := range []int64{userID, profileID} {
			keys, err := h.redis.Keys(ctx, fmt.Sprintf("likes:%d:*", uid)).Result()
			if err != nil {
				h.logger.Error("Failed to get Redis keys for likes cache",
					zap.Int64("user_id", uid),
					zap.Error(err))
			} else if len(keys) > 0 {
				if err := h.redis.Del(ctx, keys...).Err(); err != nil {
					h.logger.Error("Failed to delete Redis keys for likes cache",
						zap.Int64("user_id", uid),
						zap.Error(err))
				} else {
					h.logger.Info("Cleared likes cache",
						zap.Int64("user_id", uid),
						zap.Strings("keys", keys))
				}
			}
		}

		// Отправляем уведомления о взаимном лайке
		err = h.notifyMutualLikeWithLinks(b, userID, profileID)
		if err != nil {
			h.logger.Error("Failed to notify mutual like",
				zap.Int64("user1_id", userID),
				zap.Int64("user2_id", profileID),
				zap.Error(err))
		}
	}

	_, err = b.DeleteMessage(chatID, messageID, nil)
	if err != nil {
		h.logger.Warn("Failed to delete message",
			zap.Int64("chat_id", chatID),
			zap.Int64("message_id", messageID),
			zap.Error(err))
	}

	currentIndex := h.stateMgr.GetLikesCurrentIndex(userID) + 1
	offset := uint64(currentIndex/10) * 10
	profiles, err := h.getLikeResults(userID, offset)
	if err != nil {
		h.logger.Error("Failed to get next like profiles",
			zap.Int64("user_id", userID),
			zap.Error(err))
		_, _ = b.SendMessage(chatID, "Произошла ошибка при загрузке следующего профиля.", nil)
		return err
	}

	h.stateMgr.SetLikesCurrentIndex(userID, currentIndex)

	return h.sendLikeProfile(b, chatID, userID, profiles, currentIndex%10)
}

func (h *CallbackHandler) handleDislikeFromLikes(b *gotgbot.Bot, ctx *ext.Context, userID, profileID int64) error {
	chatID := ctx.CallbackQuery.Message.GetChat().Id
	messageID := ctx.CallbackQuery.Message.GetMessageId()

	err := h.db.Likes.DeleteByIDs(context.Background(), profileID, userID)
	if err != nil {
		h.logger.Error("Failed to delete like",
			zap.Int64("from_user_id", profileID),
			zap.Int64("to_user_id", userID),
			zap.Error(err))
		_, _ = b.SendMessage(chatID, "Произошла ошибка при удалении лайка.", nil)
		return err
	}

	// Обновляем рейтинг пользователя, чей лайк удалён
	if err := h.db.Users.UpdateRating(context.Background(), userID); err != nil {
		h.logger.Error("Failed to update rating",
			zap.Int64("user_id", userID),
			zap.Error(err))
		// Продолжаем, так как ошибка не критична
	}

	_, err = b.DeleteMessage(chatID, messageID, nil)
	if err != nil {
		h.logger.Warn("Failed to delete message",
			zap.Int64("chat_id", chatID),
			zap.Int64("message_id", messageID),
			zap.Error(err))
	}

	currentIndex := h.stateMgr.GetLikesCurrentIndex(userID) + 1
	offset := uint64(currentIndex/10) * 10
	profiles, err := h.getLikeResults(userID, offset)
	if err != nil {
		h.logger.Error("Failed to get next like profiles",
			zap.Int64("user_id", userID),
			zap.Error(err))
		_, _ = b.SendMessage(chatID, "Произошла ошибка при загрузке следующего профиля.", nil)
		return err
	}

	h.stateMgr.SetLikesCurrentIndex(userID, currentIndex)

	return h.sendLikeProfile(b, chatID, userID, profiles, currentIndex%10)
}

func (h *CallbackHandler) notifyMutualLikeWithLinks(b *gotgbot.Bot, userID1, userID2 int64) error {
	user1, err := h.db.Users.GetByID(context.Background(), userID1)
	if err != nil {
		return err
	}
	user2, err := h.db.Users.GetByID(context.Background(), userID2)
	if err != nil {
		return err
	}

	user1Name := "Пользователь"
	user1Link := fmt.Sprintf("tg://user?id=%d", userID1)
	if user1.Username != nil {
		user1Name = *user1.Username
	}

	user2Name := "Пользователь"
	user2Link := fmt.Sprintf("tg://user?id=%d", userID2)
	if user2.Username != nil {
		user2Name = *user2.Username
	}

	user1ChatID := userID1
	var cityName string
	if user2.CityID != nil {
		city, err := h.db.Cities.GetByID(context.Background(), *user2.CityID)
		if err != nil {
			h.logger.Error("Failed to get city for user2",
				zap.Int64("user_id", userID2),
				zap.Error(err))
		} else {
			cityName = city.Name
		}
	}
	user2Profile := FormatProfile(user2, cityName)
	_, err = b.SendMessage(user1ChatID,
		fmt.Sprintf("Взаимный лайк! 💕 Вы понравились %s!\n\n%s\nСвязаться: %s", user2Name, user2Profile, user2Link),
		&gotgbot.SendMessageOpts{ParseMode: "HTML"})
	if err != nil {
		h.logger.Error("Failed to notify user1",
			zap.Int64("user_id", userID1),
			zap.Error(err))
	}

	user2ChatID := userID2
	cityName = ""
	if user1.CityID != nil {
		city, err := h.db.Cities.GetByID(context.Background(), *user1.CityID)
		if err != nil {
			h.logger.Error("Failed to get city for user1",
				zap.Int64("user_id", userID1),
				zap.Error(err))
		} else {
			cityName = city.Name
		}
	}
	user1Profile := FormatProfile(user1, cityName)
	_, err = b.SendMessage(user2ChatID,
		fmt.Sprintf("Взаимный лайк! 💕 Вы понравились %s!\n\n%s\nСвязаться: %s", user1Name, user1Profile, user1Link),
		&gotgbot.SendMessageOpts{ParseMode: "HTML"})
	if err != nil {
		h.logger.Error("Failed to notify user2",
			zap.Int64("user_id", userID2),
			zap.Error(err))
	}

	return nil
}

func (h *CallbackHandler) getSearchResults(userID int64, offset uint64) ([]*db.User, error) {
	ctx := context.Background()

	cacheKey := fmt.Sprintf("search:%d:%d", userID, offset)
	cached, err := h.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var profiles []*db.User
		if err := json.Unmarshal([]byte(cached), &profiles); err == nil {
			return profiles, nil
		}
	}

	profiles, err := h.db.Users.SelectUsers(ctx, userID, offset)
	if err != nil {
		h.logger.Error("Failed to fetch profiles",
			zap.Int64("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	if len(profiles) > 0 {
		profilesJson, _ := json.Marshal(profiles)
		h.redis.Set(ctx, cacheKey, profilesJson, searchResultsTTL)
	}
	return profiles, nil
}

func (h *CallbackHandler) getLikeResults(userID int64, offset uint64) ([]*db.User, error) {
	ctx := context.Background()

	cacheKey := fmt.Sprintf("likes:%d:%d", userID, offset)
	cached, err := h.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var profiles []*db.User
		if err := json.Unmarshal([]byte(cached), &profiles); err == nil {
			return profiles, nil
		}
	}

	profiles, err := h.db.Likes.GetAllByToUserIDWithUsers(ctx, userID, offset, 10)
	if err != nil {
		h.logger.Error("Failed to fetch like profiles",
			zap.Int64("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	if len(profiles) > 0 {
		profilesJson, _ := json.Marshal(profiles)
		h.redis.Set(ctx, cacheKey, profilesJson, searchResultsTTL)
	}
	return profiles, nil
}

func (h *CallbackHandler) sendProfile(b *gotgbot.Bot, chatID, userID int64, profiles []*db.User, currentIndex int) error {
	if currentIndex >= len(profiles) {
		nextOffset := uint64((h.stateMgr.GetCurrentIndex(userID)/50)+1) * 50
		nextProfiles, err := h.getSearchResults(userID, nextOffset)
		if err != nil {
			h.logger.Error("Failed to get next profiles",
				zap.Int64("user_id", userID),
				zap.Error(err))
			_, _ = b.SendMessage(chatID, "Произошла ошибка при загрузке анкет.", nil)
			return err
		}

		if len(nextProfiles) == 0 {
			h.stateMgr.ResetCurrentIndex(userID)
			_, err := b.SendMessage(chatID, "Больше анкет не найдено.", nil)
			return err
		}

		currentIndex = 0
		h.stateMgr.SetCurrentIndex(userID, int(nextOffset))
		profiles = nextProfiles
	}

	profile := profiles[currentIndex]
	var cityName string
	if profile.CityID != nil {
		city, err := h.db.Cities.GetByID(context.Background(), *profile.CityID)
		if err != nil {
			h.logger.Error("Failed to get city for profile",
				zap.Int64("user_id", profile.ID),
				zap.Error(err))
		} else {
			cityName = city.Name
		}
	}
	profileText := FormatProfile(profile, cityName)

	keyboard := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "❤️ Лайк", CallbackData: fmt.Sprintf("like:%d", profile.ID)},
			{Text: "👎 Дизлайк", CallbackData: fmt.Sprintf("dislike:%d", profile.ID)},
		},
	}

	_, err := b.SendMessage(chatID, profileText, &gotgbot.SendMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyboard},
	})
	if err != nil {
		h.logger.Error("Failed to send profile",
			zap.Int64("chat_id", chatID),
			zap.Int64("user_id", userID),
			zap.Error(err))
	}
	return err
}

func (h *CallbackHandler) sendLikeProfile(b *gotgbot.Bot, chatID, userID int64, profiles []*db.User, currentIndex int) error {
	if currentIndex >= len(profiles) {
		nextOffset := uint64((h.stateMgr.GetLikesCurrentIndex(userID)/10)+1) * 10
		nextProfiles, err := h.getLikeResults(userID, nextOffset)
		if err != nil {
			h.logger.Error("Failed to get next like profiles",
				zap.Int64("user_id", userID),
				zap.Error(err))
			_, _ = b.SendMessage(chatID, "Произошла ошибка при загрузке лайков.", nil)
			return err
		}

		if len(nextProfiles) == 0 {
			h.stateMgr.ResetLikesCurrentIndex(userID)
			ctx := context.Background()
			keys, err := h.redis.Keys(ctx, fmt.Sprintf("likes:%d:*", userID)).Result()
			if err != nil {
				h.logger.Error("Failed to get Redis keys for likes cache",
					zap.Int64("user_id", userID),
					zap.Error(err))
			} else if len(keys) > 0 {
				if err := h.redis.Del(ctx, keys...).Err(); err != nil {
					h.logger.Error("Failed to delete Redis keys for likes cache",
						zap.Int64("user_id", userID),
						zap.Error(err))
				} else {
					h.logger.Info("Cleared likes cache",
						zap.Int64("user_id", userID),
						zap.Strings("keys", keys))
				}
			}
			_, err = b.SendMessage(chatID, "Больше лайков не найдено.", nil)
			return err
		}

		currentIndex = 0
		h.stateMgr.SetLikesCurrentIndex(userID, int(nextOffset))
		profiles = nextProfiles
	}

	profile := profiles[currentIndex]
	var cityName string
	if profile.CityID != nil {
		city, err := h.db.Cities.GetByID(context.Background(), *profile.CityID)
		if err != nil {
			h.logger.Error("Failed to get city for profile",
				zap.Int64("user_id", profile.ID),
				zap.Error(err))
		} else {
			cityName = city.Name
		}
	}
	profileText := FormatProfile(profile, cityName)

	keyboard := [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "❤️ Лайк", CallbackData: fmt.Sprintf("like_like:%d", profile.ID)},
			{Text: "👎 Дизлайк", CallbackData: fmt.Sprintf("dislike_like:%d", profile.ID)},
		},
	}

	_, err := b.SendMessage(chatID, profileText, &gotgbot.SendMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyboard},
	})
	if err != nil {
		h.logger.Error("Failed to send like profile",
			zap.Int64("chat_id", chatID),
			zap.Int64("user_id", userID),
			zap.Error(err))
	}
	return err
}
