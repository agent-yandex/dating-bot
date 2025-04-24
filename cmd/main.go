package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/agent-yandex/dating-bot/internal/config"
	"github.com/agent-yandex/dating-bot/internal/deps"
	"github.com/agent-yandex/dating-bot/internal/tg/handlers"
	"github.com/agent-yandex/dating-bot/internal/tg/states"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()

	depends, err := deps.ProvideDependencies(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to provide dependencies: %v", err)
	}
	defer depends.Cleanup()

	b, err := gotgbot.NewBot(cfg.TELEGRAM_BOT_TOKEN, nil)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	dp := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("Error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})

	updater := ext.NewUpdater(dp, nil)

	stateMgr := states.NewManager()

	messageHandler := handlers.NewMessageHandler(stateMgr, &depends.DB, depends.Logger)
	callbackHandler := handlers.NewCallbackHandler(stateMgr, &depends.DB, depends.Logger)
	commandHandler := handlers.NewCommandHandler(&depends.DB, depends.Logger, stateMgr)

	messageHandler.RegisterMessages(dp)
	callbackHandler.RegisterCallbacks(dp)
	commandHandler.RegisterCommands(dp)

	err = updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to start polling: %v", err)
	}
	log.Printf("%s is up and running...\n", b.User.Username)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	updater.Stop()
	log.Println("Bot stopped")
}
