package handlers

import (
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/agent-yandex/dating-bot/internal/deps"
	"github.com/agent-yandex/dating-bot/internal/tg/models"
	"github.com/agent-yandex/dating-bot/internal/tg/states"
	"go.uber.org/zap"
)

type MessageHandler struct {
	stateMgr            *states.Manager
	db                  *deps.DB
	logger              *zap.Logger
	tempUserData        map[int64]*models.TempUserData
	tempUserPreferences map[int64]*models.TempUserPreferencesData
	isEditing           map[int64]bool
}

func NewMessageHandler(stateMgr *states.Manager, db *deps.DB, logger *zap.Logger) *MessageHandler {
	return &MessageHandler{
		stateMgr:            stateMgr,
		db:                  db,
		logger:              logger,
		tempUserData:        make(map[int64]*models.TempUserData),
		tempUserPreferences: make(map[int64]*models.TempUserPreferencesData),
		isEditing:           make(map[int64]bool),
	}
}

func (h *MessageHandler) RegisterMessages(d *ext.Dispatcher) {
	d.AddHandler(handlers.NewMessage(
		func(msg *gotgbot.Message) bool {
			return msg.Text != "" && !strings.HasPrefix(msg.Text, "/")
		},
		h.handleMessage,
	))
}

func (h *MessageHandler) handleMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.Message == nil || ctx.Message.Text == "" {
		return nil
	}

	userID := ctx.Message.From.Id
	chatID := ctx.Message.Chat.Id
	currentState := h.stateMgr.Get(userID)

	switch ctx.Message.Text {
	case "Создать/редактировать профиль":
		return h.handleProfileCreation(b, ctx)
	case "Посмотреть профиль":
		return h.handleViewProfile(b, ctx)
	case "Изменить настройки поиска":
		return h.handleUserPreferencesEdit(b, ctx)
	case "Посмотреть настройки поиска":
		return h.handleViewUserPreferences(b, ctx)
	case "Поиск анкет":
		_, err := b.SendMessage(chatID, "Функция поиска анкет пока в разработке!", nil)
		return err
	}

	switch currentState {
	case states.StateEditName:
		return h.handleName(b, ctx)
	case states.StateEditGender:
		return h.handleGender(b, ctx)
	case states.StateEditAge:
		return h.handleAge(b, ctx)
	case states.StateEditCity:
		return h.handleCity(b, ctx)
	case states.StateEditBio:
		return h.handleBio(b, ctx)
	case states.StateEditPrefGender:
		return h.handlePrefGender(b, ctx)
	case states.StateEditPrefMinage:
		return h.handlePrefMinAge(b, ctx)
	case states.StateEditPrefMaxAge:
		return h.handlePrefMaxAge(b, ctx)
	case states.StateEditPrefMaxDistance:
		return h.handlePrefMaxDistance(b, ctx)
	default:
		return nil
	}
}
