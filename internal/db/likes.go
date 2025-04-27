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

const LikesTable = "likes"

const (
	LikesID         = "id"
	LikesFromUserID = "from_user_id"
	LikesToUserID   = "to_user_id"
	LikesMessage    = "message"
	LikesCreatedAt  = "created_at"
	LikesExpiresAt  = "expires_at"
)

type Like struct {
	ID         int64     `db:"id"`
	FromUserID int64     `db:"from_user_id" insert:"from_user_id" delete:"from_user_id"`
	ToUserID   int64     `db:"to_user_id" insert:"to_user_id" delete:"to_user_id"`
	Message    *string   `db:"message" insert:"message"`
	CreatedAt  time.Time `db:"created_at"`
	ExpiresAt  time.Time `db:"expires_at"`
}

var (
	stomLikeSelect = stom.MustNewStom(Like{}).SetTag(selectTag)
	stomLikeInsert = stom.MustNewStom(Like{}).SetTag(insertTag)
	stomLikeDelete = stom.MustNewStom(Like{}).SetTag("delete")
)

func (l *Like) columns(pref string) []string {
	return colNamesWithPref(stomLikeSelect.TagValues(), pref)
}

type LikeQuery interface {
	GetByID(ctx context.Context, id int64) (*Like, error)
	GetAllByToUserID(ctx context.Context, userID int64) ([]*Like, error)
	GetAllByToUserIDWithUsers(ctx context.Context, toUserID int64, offset, limit uint64) ([]*User, error)
	Insert(ctx context.Context, like *Like) (*Like, error)
	Delete(ctx context.Context, like *Like) error
	DeleteByIDs(ctx context.Context, fromUserID, toUserID int64) error
}

type likeQuery struct {
	runner *pgxpool.Pool
	sq     squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewLikeQuery(runner *pgxpool.Pool, sq squirrel.StatementBuilderType, logger *zap.Logger) LikeQuery {
	return &likeQuery{
		runner: runner,
		sq:     sq,
		logger: logger,
	}
}

func (l likeQuery) GetByID(ctx context.Context, id int64) (*Like, error) {
	l.logger.Debug("Fetching like by ID", zap.Int64("like_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	like := &Like{}
	qb, args, err := l.sq.Select(like.columns("")...).
		From(LikesTable).
		Where(squirrel.Eq{LikesID: id}).
		ToSql()
	if err != nil {
		l.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, l.runner, like, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			l.logger.Warn("Database error",
				zap.Int64("like_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			l.logger.Warn("Failed to fetch like", zap.Int64("like_id", id), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	l.logger.Info("Like fetched successfully", zap.Int64("like_id", id))
	return like, nil
}

func (l likeQuery) GetAllByToUserID(ctx context.Context, userID int64) ([]*Like, error) {
	l.logger.Debug("Fetching likes by to_user_id", zap.Int64("to_user_id", userID))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var likes []*Like
	qb, args, err := l.sq.Select((&Like{}).columns("")...).
		From(LikesTable).
		Where(squirrel.Eq{LikesToUserID: userID}).
		ToSql()
	if err != nil {
		l.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Select(ctx, l.runner, &likes, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			l.logger.Warn("Database error",
				zap.Int64("to_user_id", userID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			l.logger.Warn("Failed to fetch likes", zap.Int64("to_user_id", userID), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	l.logger.Info("Likes fetched successfully",
		zap.Int64("to_user_id", userID),
		zap.Int("count", len(likes)),
	)
	return likes, nil
}

func (l likeQuery) GetAllByToUserIDWithUsers(ctx context.Context, toUserID int64, offset, limit uint64) ([]*User, error) {
	l.logger.Debug("Fetching users who liked",
		zap.Int64("to_user_id", toUserID),
		zap.Uint64("offset", offset),
		zap.Uint64("limit", limit))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var users []*User
	qb := l.sq.Select((&User{}).columns("u")...).
		From(LikesTable + " l").
		InnerJoin(UsersTable + " u ON l.from_user_id = u.id").
		Where(squirrel.Eq{"l.to_user_id": toUserID}).
		Where(squirrel.Expr("l.expires_at > NOW()")).
		OrderBy("l.created_at DESC").
		Limit(limit).
		Offset(offset)

	query, args, err := qb.ToSql()
	if err != nil {
		l.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	err = pgxscan.Select(ctx, l.runner, &users, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			l.logger.Warn("Database error",
				zap.Int64("to_user_id", toUserID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err))
		} else {
			l.logger.Error("Failed to fetch users who liked",
				zap.Int64("to_user_id", toUserID),
				zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	l.logger.Info("Users who liked fetched successfully",
		zap.Int64("to_user_id", toUserID),
		zap.Int("count", len(users)))
	return users, nil
}

func (l likeQuery) Insert(ctx context.Context, like *Like) (*Like, error) {
	l.logger.Debug("Inserting like",
		zap.Int64("from_user_id", like.FromUserID),
		zap.Int64("to_user_id", like.ToUserID),
	)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	insertMap, err := stomLikeInsert.ToMap(like)
	if err != nil {
		l.logger.Error("Failed to map struct", zap.Error(err))
		return nil, fmt.Errorf("failed to map struct: %w", err)
	}
	qb, args, err := l.sq.Insert(LikesTable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		l.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, l.runner, like, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			l.logger.Warn("Database error",
				zap.Int64("from_user_id", like.FromUserID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			l.logger.Error("Failed to insert like",
				zap.Int64("from_user_id", like.FromUserID),
				zap.Error(err),
			)
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	l.logger.Info("Like inserted successfully", zap.Int64("like_id", like.ID))
	return like, nil
}

func (l likeQuery) Delete(ctx context.Context, like *Like) error {
	l.logger.Debug("Deleting like",
		zap.Int64("from_user_id", like.FromUserID),
		zap.Int64("to_user_id", like.ToUserID),
	)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	deleteMap, err := stomLikeDelete.ToMap(like)
	if err != nil {
		l.logger.Error("Failed to map struct", zap.Error(err))
		return fmt.Errorf("failed to map struct: %w", err)
	}
	qb, args, err := l.sq.Delete(LikesTable).
		Where(squirrel.Eq{
			LikesFromUserID: deleteMap[LikesFromUserID],
			LikesToUserID:   deleteMap[LikesToUserID],
		}).
		ToSql()
	if err != nil {
		l.logger.Error("Failed to build query", zap.Error(err))
		return fmt.Errorf("failed to build query: %w", err)
	}
	_, err = l.runner.Exec(ctx, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			l.logger.Warn("Database error",
				zap.Int64("from_user_id", like.FromUserID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			l.logger.Error("Failed to delete like",
				zap.Int64("from_user_id", like.FromUserID),
				zap.Error(err),
			)
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}
	l.logger.Info("Like deleted successfully",
		zap.Int64("from_user_id", like.FromUserID),
		zap.Int64("to_user_id", like.ToUserID),
	)
	return nil
}

func (l likeQuery) DeleteByIDs(ctx context.Context, fromUserID, toUserID int64) error {
	l.logger.Debug("Deleting like by IDs",
		zap.Int64("from_user_id", fromUserID),
		zap.Int64("to_user_id", toUserID),
	)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	qb, args, err := l.sq.Delete(LikesTable).
		Where(squirrel.Eq{
			LikesFromUserID: fromUserID,
			LikesToUserID:   toUserID,
		}).
		ToSql()
	if err != nil {
		l.logger.Error("Failed to build query", zap.Error(err))
		return fmt.Errorf("failed to build query: %w", err)
	}
	result, err := l.runner.Exec(ctx, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			l.logger.Warn("Database error",
				zap.Int64("from_user_id", fromUserID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			l.logger.Error("Failed to delete like",
				zap.Int64("from_user_id", fromUserID),
				zap.Error(err),
			)
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}
	if result.RowsAffected() == 0 {
		l.logger.Warn("No like found to delete",
			zap.Int64("from_user_id", fromUserID),
			zap.Int64("to_user_id", toUserID))
	}
	l.logger.Info("Like deleted successfully",
		zap.Int64("from_user_id", fromUserID),
		zap.Int64("to_user_id", toUserID),
	)
	return nil
}
