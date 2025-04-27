package handlers

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/agent-yandex/dating-bot/internal/tg/states"
	"go.uber.org/zap"
)

func (h *MessageHandler) handleViewLikes(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id

	h.stateMgr.Set(userID, states.StateViewLikes)

	currentIndex := h.stateMgr.GetLikesCurrentIndex(userID)
	offset := uint64(currentIndex/10) * 10
	profiles, err := h.callback.getLikeResults(userID, offset)
	if err != nil {
		h.logger.Error("Failed to get like results",
			zap.Int64("user_id", userID),
			zap.Error(err))
		_, err = b.SendMessage(chatID, "Произошла ошибка при загрузке лайков. Попробуйте позже.", nil)
		return err
	}

	if len(profiles) == 0 {
		h.stateMgr.ResetLikesCurrentIndex(userID)
		_, err = b.SendMessage(chatID, "Больше лайков не найдено.", nil)
		return err
	}

	return h.callback.sendLikeProfile(b, chatID, userID, profiles, currentIndex%10)
}
