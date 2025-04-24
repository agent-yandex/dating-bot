package handlers

import (
	"context"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/agent-yandex/dating-bot/internal/db"
	"github.com/agent-yandex/dating-bot/internal/tg/models"
	"github.com/agent-yandex/dating-bot/internal/tg/states"
	"go.uber.org/zap"
)

func (h *MessageHandler) handleViewProfile(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id

	user, err := h.db.Users.GetByID(context.Background(), userID)
	if err != nil {
		h.logger.Error("Failed to fetch user profile", zap.Int64("user_id", userID), zap.Error(err))
		_, err := b.SendMessage(chatID, "Произошла ошибка при загрузке профиля.", nil)
		return err
	}
	if user == nil {
		_, err := b.SendMessage(chatID, "У вас нет профиля. Создайте его сначала.", nil)
		return err
	}
	city, err := h.db.Cities.GetByID(context.Background(), *user.CityID)
	if err != nil {
		h.logger.Error("Failed to fetch city Name for view profile", zap.Int64("user_id", userID), zap.Error(err))
	}

	profileText := FormatProfile(user, city.Name)
	_, err = b.SendMessage(chatID, profileText, &gotgbot.SendMessageOpts{
		ReplyMarkup: GetMainReplyKeyboard(),
	})
	return err
}

func (h *MessageHandler) handleProfileCreation(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id

	user, err := h.db.Users.GetByID(context.Background(), userID)
	if err != nil {
		h.logger.Error("Failed to check user existence", zap.Int64("user_id", userID), zap.Error(err))
		_, err := b.SendMessage(chatID, "Произошла ошибка. Попробуйте снова.", nil)
		return err
	}
	if user != nil {
		h.isEditing[userID] = true
		h.stateMgr.Set(userID, states.StateEditName)
		_, err := b.SendMessage(chatID, "У вас уже есть профиль. Давайте отредактируем его. Введите ваше имя:", &gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.ReplyKeyboardRemove{RemoveKeyboard: true},
		})
		return err
	}

	h.stateMgr.Set(userID, states.StateEditName)
	h.isEditing[userID] = false
	_, err = b.SendMessage(chatID, "Отлично! Теперь введите ваше имя:", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardRemove{RemoveKeyboard: true},
	})
	return err
}

func (h *MessageHandler) handleName(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id
	input := ctx.Message.Text

	if _, exists := h.tempUserData[userID]; !exists {
		h.tempUserData[userID] = &models.TempUserData{}
	}

	if len(input) > 50 {
		_, err := b.SendMessage(chatID, "Имя слишком длинное (макс. 50 символов). Попробуйте снова:", nil)
		return err
	}

	h.tempUserData[userID].Username = &input
	h.stateMgr.Set(userID, states.StateEditGender)
	_, err := b.SendMessage(chatID, "Укажите ваш пол:", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
			Keyboard: [][]gotgbot.KeyboardButton{
				{{Text: "Мужской"}, {Text: "Женский"}},
			},
			ResizeKeyboard:  true,
			OneTimeKeyboard: true,
		},
	})
	return err
}

func (h *MessageHandler) handleGender(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id
	input := strings.ToLower(ctx.Message.Text)

	var gender string
	switch input {
	case "мужской":
		gender = "m"
	case "женский":
		gender = "f"
	default:
		_, err := b.SendMessage(chatID, "Пожалуйста, выберите пол из предложенных вариантов:", &gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
				Keyboard: [][]gotgbot.KeyboardButton{
					{{Text: "Мужской"}, {Text: "Женский"}},
				},
				ResizeKeyboard:  true,
				OneTimeKeyboard: true,
			},
		})
		return err
	}

	if _, exists := h.tempUserData[userID]; !exists {
		h.tempUserData[userID] = &models.TempUserData{}
	}
	h.tempUserData[userID].Gender = gender

	h.stateMgr.Set(userID, states.StateEditAge)
	_, err := b.SendMessage(chatID, "Введите ваш возраст (10-100):", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardRemove{RemoveKeyboard: true},
	})
	return err
}

func (h *MessageHandler) handleAge(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id
	input := ctx.Message.Text

	age, err := strconv.Atoi(input)
	if err != nil || age < 10 || age > 100 {
		_, err := b.SendMessage(chatID, "Пожалуйста, введите корректный возраст (10-100):", nil)
		return err
	}

	if _, exists := h.tempUserData[userID]; !exists {
		h.tempUserData[userID] = &models.TempUserData{}
	}
	h.tempUserData[userID].Age = age

	h.stateMgr.Set(userID, states.StateEditCity)
	_, err = b.SendMessage(chatID, "Введите ваш город:", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardRemove{RemoveKeyboard: true},
	})
	return err
}

func (h *MessageHandler) handleCity(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id
	input := ctx.Message.Text

	cityID, err := h.db.Cities.GetIDByName(context.Background(), input)
	if err != nil {
		h.logger.Error("Failed to find city", zap.String("city_name", input), zap.Error(err))
		_, err := b.SendMessage(chatID, "Произошла ошибка при поиске города. Попробуйте снова:", nil)
		return err
	}
	if cityID == 0 {
		_, err := b.SendMessage(chatID, "Город не найден. Уточните название (например, Москва, Санкт-Петербург, Нижний Новгород):", nil)
		return err
	}

	if _, exists := h.tempUserData[userID]; !exists {
		h.tempUserData[userID] = &models.TempUserData{}
	}
	h.tempUserData[userID].CityID = &cityID

	h.stateMgr.Set(userID, states.StateEditBio)
	_, err = b.SendMessage(chatID, "Расскажите о себе (макс. 500 символов):", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardRemove{RemoveKeyboard: true},
	})
	return err
}

func (h *MessageHandler) handleBio(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id
	input := ctx.Message.Text

	if len(input) > 500 {
		_, err := b.SendMessage(chatID, "Описание слишком длинное (макс. 500 символов). Попробуйте снова:", nil)
		return err
	}

	if _, exists := h.tempUserData[userID]; !exists {
		h.tempUserData[userID] = &models.TempUserData{}
	}
	h.tempUserData[userID].Bio = &input

	return h.finalizeProfile(b, ctx)
}

func (h *MessageHandler) finalizeProfile(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id
	tempData := h.tempUserData[userID]

	user := &db.User{
		Username: tempData.Username,
		Gender:   tempData.Gender,
		Age:      tempData.Age,
		CityID:   tempData.CityID,
		Bio:      tempData.Bio,
		IsActive: true,
	}

	var err error
	var successMessage string
	if h.isEditing[userID] {
		_, err = h.db.Users.Update(context.Background(), user, userID)
		successMessage = "Профиль обновлен! Что хотите сделать дальше?"
	} else {
		user.ID = userID
		_, err = h.db.Users.Insert(context.Background(), user)
		_, err = h.db.UserPreferences.Insert(context.Background(), userID)
		successMessage = "Профиль создан! Теперь вы можете искать другие анкеты."
	}

	if err != nil {
		h.logger.Error("Failed to save profile", zap.Int64("user_id", userID), zap.Error(err))
		_, err = b.SendMessage(chatID, "Произошла ошибка при сохранении профиля. Попробуйте позже.", nil)
		return err
	}

	delete(h.tempUserData, userID)
	delete(h.isEditing, userID)
	h.stateMgr.Reset(userID)

	_, err = b.SendMessage(chatID, successMessage, &gotgbot.SendMessageOpts{
		ReplyMarkup: GetMainReplyKeyboard(),
	})
	return err
}
