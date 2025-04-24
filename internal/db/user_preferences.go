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

const UserPreferencesTable = "user_preferences"

const (
	UserPreferencesUserID      = "user_id"
	UserPreferencesMinAge      = "min_age"
	UserPreferencesMaxAge      = "max_age"
	UserPreferencesGenderPref  = "gender_preference"
	UserPreferencesMaxDistance = "max_distance_km"
	UserPreferencesUpdatedAt   = "updated_at"
)

type UserPreference struct {
	UserID      int64     `db:"user_id" insert:"user_id"`
	MinAge      int       `db:"min_age" insert:"min_age" update:"min_age"`
	MaxAge      int       `db:"max_age" insert:"max_age" update:"max_age"`
	GenderPref  string    `db:"gender_preference" insert:"gender_preference" update:"gender_preference"`
	MaxDistance int       `db:"max_distance_km" insert:"max_distance_km" update:"max_distance_km"`
	UpdatedAt   time.Time `db:"updated_at"`
}

var (
	stomPrefSelect = stom.MustNewStom(UserPreference{}).SetTag(selectTag)
	stomPrefUpdate = stom.MustNewStom(UserPreference{}).SetTag(updateTag)
	stomPrefInsert = stom.MustNewStom(UserPreference{}).SetTag(insertTag)
)

func (up *UserPreference) columns(pref string) []string {
	return colNamesWithPref(stomPrefSelect.TagValues(), pref)
}

type UserPreferencesQuery interface {
	GetByUserID(ctx context.Context, userID int64) (*UserPreference, error)
	Insert(ctx context.Context, id int64) (*UserPreference, error)
	Update(ctx context.Context, pref *UserPreference, id int64) error
}

type userPreferencesQuery struct {
	runner *pgxpool.Pool
	sq     squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewUserPreferencesQuery(runner *pgxpool.Pool, sq squirrel.StatementBuilderType, logger *zap.Logger) UserPreferencesQuery {
	return &userPreferencesQuery{
		runner: runner,
		sq:     sq,
		logger: logger,
	}
}

func (up userPreferencesQuery) GetByUserID(ctx context.Context, userID int64) (*UserPreference, error) {
	up.logger.Debug("Fetching user preference", zap.Int64("user_id", userID))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pref := &UserPreference{}
	qb, args, err := up.sq.Select(pref.columns("")...).
		From(UserPreferencesTable).
		Where(squirrel.Eq{UserPreferencesUserID: userID}).
		ToSql()
	if err != nil {
		up.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, up.runner, pref, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			up.logger.Warn("Database error",
				zap.Int64("user_id", userID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			up.logger.Warn("Failed to fetch preference", zap.Int64("user_id", userID), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	up.logger.Info("Preference fetched successfully", zap.Int64("user_id", userID))
	return pref, nil
}

func (up userPreferencesQuery) Insert(ctx context.Context, id int64) (*UserPreference, error) {
	up.logger.Debug("Inserting user preference", zap.Int64("user_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	qb, args, err := up.sq.Insert(UserPreferencesTable).
		Columns(UserPreferencesUserID).
		Values(id).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		up.logger.Error("Failed to build query",
			zap.Int64("user_id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	result := &UserPreference{}
	err = pgxscan.Get(ctx, up.runner, result, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			up.logger.Warn("Database error",
				zap.Int64("user_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			up.logger.Error("Failed to insert preference",
				zap.Int64("user_id", id),
				zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	up.logger.Info("Preference inserted successfully",
		zap.Int64("user_id", id))
	return result, nil
}

func (up userPreferencesQuery) Update(ctx context.Context, pref *UserPreference, id int64) error {
	up.logger.Debug("Updating user preference", zap.Int64("user_id", pref.UserID))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	updateMap, err := stomPrefUpdate.ToMap(pref)
	if err != nil {
		up.logger.Error("Failed to map struct", zap.Error(err))
		return fmt.Errorf("failed to map struct: %w", err)
	}
	qb, args, err := up.sq.Update(UserPreferencesTable).
		SetMap(updateMap).
		Where(squirrel.Eq{UserPreferencesUserID: id}).
		ToSql()
	if err != nil {
		up.logger.Error("Failed to build query", zap.Error(err))
		return fmt.Errorf("failed to build query: %w", err)
	}
	_, err = up.runner.Exec(ctx, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			up.logger.Warn("Database error",
				zap.Int64("user_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			up.logger.Error("Failed to update preference", zap.Int64("user_id", id), zap.Error(err))
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}
	up.logger.Info("Preference updated successfully", zap.Int64("user_id", id))
	return nil
}
