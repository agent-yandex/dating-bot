package db

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/elgris/stom"
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
	UserID      int64  `db:"user_id" insert:"user_id"`
	MinAge      int    `db:"min_age" insert:"min_age" update:"min_age"`
	MaxAge      int    `db:"max_age" insert:"max_age" update:"max_age"`
	GenderPref  string `db:"gender_preference" insert:"gender_preference" update:"gender_preference"`
	MaxDistance int    `db:"max_distance_km" insert:"max_distance_km" update:"max_distance_km"`
	UpdatedAt   string `db:"updated_at"`
}

var (
	stomPrefSelect = stom.MustNewStom(UserPreference{}).SetTag(selectTag)
	stomPrefUpdate = stom.MustNewStom(UserPreference{}).SetTag(updateTag)
	stomPrefInsert = stom.MustNewStom(UserPreference{}).SetTag(insertTag)
)

func (up UserPreference) getTableName() string {
	return UserPreferencesTable
}

func (up UserPreference) columns(pref string) []string {
	return colNamesWithPref(
		stomPrefSelect.TagValues(),
		pref,
	)
}

type UserPreferencesQuery interface {
	GetByUserID(ctx context.Context, userID int64) (*UserPreference, error)
	Insert(ctx context.Context, pref *UserPreference) (int64, error)
	Update(ctx context.Context, pref *UserPreference) error
}

type userPreferencesQuery struct {
	runner *sql.DB
	sq     squirrel.StatementBuilderType
}

func NewUserPreferencesQuery(runner *sql.DB, sq squirrel.StatementBuilderType) UserPreferencesQuery {
	return &userPreferencesQuery{
		runner: runner,
		sq:     sq,
	}
}

func (up *userPreferencesQuery) GetByUserID(ctx context.Context, userID int64) (*UserPreference, error) {
	return getByID[UserPreference](ctx, up.runner, up.sq, UserPreferencesUserID, userID)
}

func (up *userPreferencesQuery) Insert(ctx context.Context, pref *UserPreference) (int64, error) {
	insertMap, err := stomPrefInsert.ToMap(pref)
	if err != nil {
		return 0, err
	}
	qb, args, err := up.sq.Insert(pref.getTableName()).
		SetMap(insertMap).
		ToSql()
	if err != nil {
		return 0, err
	}
	res, err := up.runner.ExecContext(ctx, qb, args...)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (up *userPreferencesQuery) Update(ctx context.Context, pref *UserPreference) error {
	updateMap, err := stomPrefUpdate.ToMap(pref)
	if err != nil {
		return err
	}
	qb, args, err := up.sq.Update(UserPreferencesTable).
		SetMap(updateMap).
		Where(squirrel.Eq{UserPreferencesUserID: pref.UserID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = up.runner.ExecContext(ctx, qb, args...)

	return err
}
