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
	"github.com/go-redis/redis/v8"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6378",
		Password: "",
		DB:       0,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("Successfully connected to Redis")

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

	stateMgr := states.NewManager(redisClient)

	callbackHandler := handlers.NewCallbackHandler(stateMgr, &depends.DB, redisClient, depends.Logger)
	messageHandler := handlers.NewMessageHandler(stateMgr, &depends.DB, redisClient, callbackHandler, depends.Logger)
	commandHandler := handlers.NewCommandHandler(stateMgr, &depends.DB, depends.Logger)

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
