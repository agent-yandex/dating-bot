package handlers

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/agent-yandex/dating-bot/internal/deps"
	"github.com/agent-yandex/dating-bot/internal/tg/states"
	"go.uber.org/zap"
)

// CallbackHandler manages callback queries from inline keyboards.
type CallbackHandler struct {
	stateMgr *states.Manager
	db       *deps.DB
	logger   *zap.Logger
}

// NewCallbackHandler creates a new instance of CallbackHandler.
func NewCallbackHandler(stateMgr *states.Manager, db *deps.DB, logger *zap.Logger) *CallbackHandler {
	return &CallbackHandler{
		stateMgr: stateMgr,
		db:       db,
		logger:   logger,
	}
}

// RegisterCallbacks registers the callback query handler with the dispatcher.
func (h *CallbackHandler) RegisterCallbacks(d *ext.Dispatcher) {
	d.AddHandler(handlers.NewCallback(nil, h.callbackHandler))
}

// callbackHandler processes callback queries from inline keyboards.
func (h *CallbackHandler) callbackHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.CallbackQuery == nil {
		return nil
	}

	//userID := ctx.CallbackQuery.From.Id
	//chatID := ctx.CallbackQuery.Message.GetChat().Id
	//
	//switch ctx.CallbackQuery.Data {
	//case "some":
	//
	//}
	return nil
}
