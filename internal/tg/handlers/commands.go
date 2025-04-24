package handlers

import (
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/agent-yandex/dating-bot/internal/deps"
	"github.com/agent-yandex/dating-bot/internal/tg/states"
	"go.uber.org/zap"
)

// CommandHandler manages command handling for the Telegram bot.
type CommandHandler struct {
	db       *deps.DB
	logger   *zap.Logger
	stateMgr *states.Manager
}

// NewCommandHandler creates a new instance of CommandHandler.
func NewCommandHandler(db *deps.DB, logger *zap.Logger, stateMgr *states.Manager) *CommandHandler {
	return &CommandHandler{
		db:       db,
		logger:   logger,
		stateMgr: stateMgr,
	}
}

// RegisterCommands registers the command handlers with the dispatcher.
func (h *CommandHandler) RegisterCommands(d *ext.Dispatcher) {
	d.AddHandler(handlers.NewCommand("start", h.handleStart))
}

// handleStart processes the /start command.
func (h *CommandHandler) handleStart(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id

	user, err := h.db.Users.GetByID(context.Background(), userID)
	if err != nil {
		h.logger.Error("Failed to check user existence", zap.Int64("user_id", userID), zap.Error(err))
		_, err := b.SendMessage(chatID, "Произошла ошибка. Попробуйте позже.", nil)
		return err
	}

	welcomeMessage := "Добро пожаловать! 👋\nЯ помогу вам найти интересных людей. Что хотите сделать?"
	var keyboard [][]gotgbot.KeyboardButton

	if user == nil {
		keyboard = [][]gotgbot.KeyboardButton{
			{{Text: "Создать/редактировать профиль"}},
		}
	} else {
		// Profile exists
		keyboard = [][]gotgbot.KeyboardButton{
			{{Text: "Посмотреть профиль"}},
			{{Text: "Создать/редактировать профиль"}},
			{{Text: "Изменить настройки поиска"}},
			{{Text: "Посмотреть настройки поиска"}},
			{{Text: "Поиск анкет"}},
		}
	}

	// Send welcome message with reply keyboard
	_, err = b.SendMessage(chatID, welcomeMessage, &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
			Keyboard:       keyboard,
			ResizeKeyboard: true,
		},
	})
	return err
}
