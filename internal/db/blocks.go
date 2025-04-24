package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/elgris/stom"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const BlocksTable = "blocks"

const (
	BlocksID        = "id"
	BlocksBlockerID = "blocker_id"
	BlocksBlockedID = "blocked_id"
	BlocksCreatedAt = "created_at"
)

type Block struct {
	ID        int64     `db:"id"`
	BlockerID int64     `db:"blocker_id" insert:"blocker_id" delete:"blocker_id"`
	BlockedID int64     `db:"blocked_id" insert:"blocked_id" delete:"blocked_id"`
	CreatedAt time.Time `db:"created_at"`
}

var (
	stomBlockSelect = stom.MustNewStom(Block{}).SetTag(selectTag)
	stomBlockInsert = stom.MustNewStom(Block{}).SetTag(insertTag)
	stomBlockDelete = stom.MustNewStom(Block{}).SetTag(deleteTag)
)

func (b *Block) columns(pref string) []string {
	return colNamesWithPref(stomBlockSelect.TagValues(), pref)
}

type BlockQuery interface {
	GetByID(ctx context.Context, id int64) (*Block, error)
	GetAllByBlockerID(ctx context.Context, blockerID int64) ([]*Block, error)
	Insert(ctx context.Context, block *Block) (*Block, error)
	Delete(ctx context.Context, block *Block) error
}

type blockQuery struct {
	runner *pgxpool.Pool
	sq     squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewBlockQuery(runner *pgxpool.Pool, sq squirrel.StatementBuilderType, logger *zap.Logger) BlockQuery {
	return &blockQuery{
		runner: runner,
		sq:     sq,
		logger: logger,
	}
}

func (b blockQuery) GetByID(ctx context.Context, id int64) (*Block, error) {
	b.logger.Debug("Fetching block by ID", zap.Int64("block_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	block := &Block{}
	qb, args, err := b.sq.Select(block.columns("")...).
		From(BlocksTable).
		Where(squirrel.Eq{BlocksID: id}).
		ToSql()
	if err != nil {
		b.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, b.runner, block, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			b.logger.Warn("Database error",
				zap.Int64("block_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			b.logger.Warn("Failed to fetch block", zap.Int64("block_id", id), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	b.logger.Info("Block fetched successfully", zap.Int64("block_id", id))
	return block, nil
}

func (b blockQuery) GetAllByBlockerID(ctx context.Context, blockerID int64) ([]*Block, error) {
	b.logger.Debug("Fetching blocks by blocker ID", zap.Int64("blocker_id", blockerID))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var blocks []*Block
	qb, args, err := b.sq.Select((&Block{}).columns("")...).
		From(BlocksTable).
		Where(squirrel.Eq{BlocksBlockerID: blockerID}).
		ToSql()
	if err != nil {
		b.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Select(ctx, b.runner, &blocks, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			b.logger.Warn("Database error",
				zap.Int64("blocker_id", blockerID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			b.logger.Error("Failed to fetch blocks", zap.Int64("blocker_id", blockerID), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	b.logger.Info("Blocks fetched successfully",
		zap.Int64("blocker_id", blockerID),
		zap.Int("count", len(blocks)),
	)
	return blocks, nil
}

func (b blockQuery) Insert(ctx context.Context, block *Block) (*Block, error) {
	b.logger.Debug("Inserting block",
		zap.Int64("blocker_id", block.BlockerID),
		zap.Int64("blocked_id", block.BlockedID),
	)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	insertMap, err := stomBlockInsert.ToMap(block)
	if err != nil {
		b.logger.Error("Failed to map struct", zap.Error(err))
		return nil, fmt.Errorf("failed to map struct: %w", err)
	}
	qb, args, err := b.sq.Insert(BlocksTable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		b.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, b.runner, block, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			b.logger.Warn("Database error",
				zap.Int64("blocker_id", block.BlockerID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			b.logger.Error("Failed to insert block",
				zap.Int64("blocker_id", block.BlockerID),
				zap.Error(err),
			)
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	b.logger.Info("Block inserted successfully", zap.Int64("block_id", block.ID))
	return block, nil
}

func (b blockQuery) Delete(ctx context.Context, block *Block) error {
	b.logger.Debug("Deleting block",
		zap.Int64("blocker_id", block.BlockerID),
		zap.Int64("blocked_id", block.BlockedID),
	)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	deleteMap, err := stomBlockDelete.ToMap(block)
	if err != nil {
		b.logger.Error("Failed to map struct", zap.Error(err))
		return fmt.Errorf("failed to map struct: %w", err)
	}
	qb, args, err := b.sq.Delete(BlocksTable).
		Where(squirrel.Eq{
			BlocksBlockerID: deleteMap[BlocksBlockerID],
			BlocksBlockedID: deleteMap[BlocksBlockedID],
		}).
		ToSql()
	if err != nil {
		b.logger.Error("Failed to build query", zap.Error(err))
		return fmt.Errorf("failed to build query: %w", err)
	}
	_, err = b.runner.Exec(ctx, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			b.logger.Warn("Database error",
				zap.Int64("blocker_id", block.BlockerID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			b.logger.Error("Failed to delete block",
				zap.Int64("blocker_id", block.BlockerID),
				zap.Error(err),
			)
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}
	b.logger.Info("Block deleted successfully",
		zap.Int64("blocker_id", block.BlockerID),
		zap.Int64("blocked_id", block.BlockedID),
	)
	return nil
}
