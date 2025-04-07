package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"

	"github.com/agent-yandex/dating-bot/internal/config"
	"github.com/agent-yandex/dating-bot/internal/db"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()

	runner, err := initDB(cfg)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	defer runner.Close()
	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	profileQuery := db.NewProfilesQuery(runner, sq)
	fmt.Println(profileQuery.GetByID(ctx, 1))
}

func initDB(cfg config.AppConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
	var err error
	var runner *sql.DB
	for i := 0; i < 5; i++ {
		runner, err = sql.Open("postgres", connStr)
		if err == nil {
			err = runner.Ping()
			if err == nil {
				break
			}
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
		return nil, err
	}

	return runner, nil
}
