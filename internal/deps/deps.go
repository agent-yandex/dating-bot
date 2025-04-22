package deps

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/agent-yandex/dating-bot/internal/config"
	"github.com/agent-yandex/dating-bot/internal/db"
	"github.com/agent-yandex/dating-bot/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DB struct {
	Users           db.UserQuery
	UserPreferences db.UserPreferencesQuery
	Blocks          db.BlockQuery
	Cities          db.CityQuery
	Likes           db.LikeQuery
}

type Dependencies struct {
	DB     DB
	Pool   *pgxpool.Pool
	Logger *zap.Logger
}

func ProvideDependencies(ctx context.Context, cfg config.AppConfig) (*Dependencies, error) {
	logger := logger.NewLogger()

	pool, err := db.InitDB(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to init db", zap.Error(err))
		return nil, err
	}

	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	deps := &Dependencies{
		DB: DB{
			Users:           db.NewUserQuery(pool, sq, logger),
			UserPreferences: db.NewUserPreferencesQuery(pool, sq, logger),
			Blocks:          db.NewBlockQuery(pool, sq, logger),
			Cities:          db.NewCityQuery(pool, sq, logger),
			Likes:           db.NewLikeQuery(pool, sq, logger),
		},
		Pool:   pool,
		Logger: logger,
	}

	if err := pool.Ping(ctx); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
		pool.Close()
		return nil, err
	}

	logger.Info("Dependencies initialized successfully")
	return deps, nil
}

func (d *Dependencies) Cleanup() {
	d.Logger.Info("Cleaning up dependencies")
	d.Logger.Sync()
	d.Pool.Close()
}
