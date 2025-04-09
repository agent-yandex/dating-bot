package db

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/elgris/stom"
	"github.com/georgysavva/scany/v2/sqlscan"
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
	MaxDistance int    `db:"max_distance_km" insert:"MaxDistance" update:"MaxDistance"`
	UpdatedAt   string `db:"updated_at"`
}

var (
	stomPrefSelect = stom.MustNewStom(UserPreference{}).SetTag(selectTag)
	stomPrefUpdate = stom.MustNewStom(UserPreference{}).SetTag(updateTag)
	stomPrefInsert = stom.MustNewStom(UserPreference{}).SetTag(insertTag)
)

func (up *UserPreference) columns(pref string) []string {
	return colNamesWithPref(
		stomPrefSelect.TagValues(),
		pref,
	)
}

type UserPreferencesQuery interface {
	GetByUserID(ctx context.Context, userID int64) (*UserPreference, error)
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
	pref := &UserPreference{}
	qb, args, err := up.sq.Select(pref.columns("")...).
		From(UserPreferencesTable).
		Where(squirrel.Eq{UserPreferencesUserID: userID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return pref, sqlscan.Get(ctx, up.runner, pref, qb, args...)
}
