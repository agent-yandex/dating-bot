package handlers

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/agent-yandex/dating-bot/internal/tg/states"
	"go.uber.org/zap"
)

func (h *MessageHandler) handleSearching(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id

	h.stateMgr.Set(userID, states.StateSearching)

	currentIndex := h.stateMgr.GetCurrentIndex(userID)
	offset := uint64(currentIndex/50) * 50
	profiles, err := h.callback.getSearchResults(userID, offset)
	if err != nil {
		h.logger.Error("Failed to get search results",
			zap.Int64("user_id", userID),
			zap.Error(err))
		_, err = b.SendMessage(chatID, "Произошла ошибка при поиске анкет. Попробуйте позже.", nil)
		return err
	}

	if len(profiles) == 0 {
		h.stateMgr.ResetCurrentIndex(userID)
		_, err = b.SendMessage(chatID, "Анкет не найдено. Попробуйте изменить настройки поиска", nil)
		return err
	}

	return h.callback.sendProfile(b, chatID, userID, profiles, currentIndex%50)
}
