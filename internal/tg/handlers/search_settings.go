package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/agent-yandex/dating-bot/internal/db"
	"github.com/agent-yandex/dating-bot/internal/tg/models"
	"github.com/agent-yandex/dating-bot/internal/tg/states"
	"go.uber.org/zap"
)

func (h *MessageHandler) handleViewUserPreferences(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id

	userPref, err := h.db.UserPreferences.GetByUserID(context.Background(), userID)
	if err != nil {
		h.logger.Error("Failed to fetch user preferences", zap.Int64("user_id", userID), zap.Error(err))
		_, err := b.SendMessage(chatID, "Произошла ошибка при загрузке настроек поиска.", nil)
		return err
	}

	userPrefText := FormatUserPref(userPref)
	_, err = b.SendMessage(chatID, userPrefText, &gotgbot.SendMessageOpts{
		ReplyMarkup: GetMainReplyKeyboard(),
	})
	return err
}

func (h *MessageHandler) handleUserPreferencesEdit(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id

	h.stateMgr.Set(userID, states.StateEditPrefGender)
	_, err := b.SendMessage(chatID, "Укажите интересующий Вас пол:", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
			Keyboard: [][]gotgbot.KeyboardButton{
				{{Text: "Мужской"}, {Text: "Женский"}, {Text: "Любой"}},
			},
			ResizeKeyboard:  true,
			OneTimeKeyboard: true,
		},
	})
	return err
}

func (h *MessageHandler) handlePrefGender(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id
	input := strings.ToLower(ctx.Message.Text)

	var gender string
	switch input {
	case "мужской":
		gender = "m"
	case "женский":
		gender = "f"
	case "любой":
		gender = "a"
	default:
		_, err := b.SendMessage(chatID, "Пожалуйста, выберите пол из предложенных вариантов:", &gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
				Keyboard: [][]gotgbot.KeyboardButton{
					{{Text: "Мужской"}, {Text: "Женский"}, {Text: "Любой"}},
				},
				ResizeKeyboard:  true,
				OneTimeKeyboard: true,
			},
		})
		return err
	}

	if _, exists := h.tempUserPreferences[userID]; !exists {
		h.tempUserPreferences[userID] = &models.TempUserPreferencesData{}
	}
	h.tempUserPreferences[userID].Gender = gender

	h.stateMgr.Set(userID, states.StateEditPrefMinage)
	_, err := b.SendMessage(chatID, "Введите минимальный возраст поиска (10-100):", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardRemove{RemoveKeyboard: true},
	})
	return err
}

func (h *MessageHandler) handlePrefMinAge(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id
	input := ctx.Message.Text

	minAge, err := strconv.Atoi(input)
	if err != nil || minAge < 10 || minAge > 100 {
		_, err := b.SendMessage(chatID, "Пожалуйста, введите корректный возраст (10-100):", nil)
		return err
	}

	if _, exists := h.tempUserPreferences[userID]; !exists {
		h.tempUserPreferences[userID] = &models.TempUserPreferencesData{}
	}
	h.tempUserPreferences[userID].MinAge = minAge

	h.stateMgr.Set(userID, states.StateEditPrefMaxAge)
	_, err = b.SendMessage(chatID, "Введите максимальный возраст поиска (10-100):", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardRemove{RemoveKeyboard: true},
	})
	return err
}

func (h *MessageHandler) handlePrefMaxAge(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id
	input := ctx.Message.Text

	maxAge, err := strconv.Atoi(input)
	if err != nil || maxAge < 10 || maxAge > 100 || maxAge < h.tempUserPreferences[userID].MinAge {
		_, err := b.SendMessage(chatID, "Пожалуйста, введите корректный возраст (10-100)\nОн должен быть выше минимального:", nil)
		return err
	}

	if _, exists := h.tempUserPreferences[userID]; !exists {
		h.tempUserPreferences[userID] = &models.TempUserPreferencesData{}
	}
	h.tempUserPreferences[userID].MaxAge = maxAge

	h.stateMgr.Set(userID, states.StateEditPrefMaxDistance)
	_, err = b.SendMessage(chatID, "Введите область поиска (в км):", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardRemove{RemoveKeyboard: true},
	})
	return err
}

func (h *MessageHandler) handlePrefMaxDistance(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id
	input := ctx.Message.Text

	maxDistance, err := strconv.Atoi(input)
	if err != nil || maxDistance <= 0 {
		_, err = b.SendMessage(chatID, "Введите корректную область поиска, > 0:", nil)
		return err
	}
	if _, exists := h.tempUserPreferences[userID]; !exists {
		h.tempUserPreferences[userID] = &models.TempUserPreferencesData{}
	}
	h.tempUserPreferences[userID].MaxDistance = maxDistance
	return h.finalizeUserPreferences(b, ctx)
}

func (h *MessageHandler) finalizeUserPreferences(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id
	tempData := h.tempUserPreferences[userID]

	userPref := &db.UserPreference{
		GenderPref:  tempData.Gender,
		MinAge:      tempData.MinAge,
		MaxAge:      tempData.MaxAge,
		MaxDistance: tempData.MaxDistance,
	}

	err := h.db.UserPreferences.Update(context.Background(), userPref, userID)
	if err != nil {
		h.logger.Error("Failed to save profile", zap.Int64("user_id", userID), zap.Error(err))
		_, err = b.SendMessage(chatID, "Произошла ошибка при сохранении профиля. Попробуйте позже.", nil)
		return err
	}

	h.stateMgr.ResetCurrentIndex(userID)
	ctxRedis := context.Background()
	keys, err := h.redis.Keys(ctxRedis, fmt.Sprintf("search:%d:*", userID)).Result()
	if err != nil {
		h.logger.Error("Failed to fetch search cache keys",
			zap.Int64("user_id", userID),
			zap.Error(err))
	} else if len(keys) > 0 {
		_, err = h.redis.Del(ctxRedis, keys...).Result()
		if err != nil {
			h.logger.Error("Failed to delete search cache",
				zap.Int64("user_id", userID),
				zap.Error(err))
		}
	}

	successMessage := "Настройки поиска обновлены! Что хотите сделать дальше?"

	delete(h.tempUserPreferences, userID)
	delete(h.isEditing, userID)
	h.stateMgr.Reset(userID)

	_, err = b.SendMessage(chatID, successMessage, &gotgbot.SendMessageOpts{
		ReplyMarkup: GetMainReplyKeyboard(),
	})
	return err
}
