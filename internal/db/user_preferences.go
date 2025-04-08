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
	UserPreferencesHashKey     = "hash_key"
	UserPreferencesMinAge      = "min_age"
	UserPreferencesMaxAge      = "max_age"
	UserPreferencesGenderPref  = "gender_preference"
	UserPreferencesSameCountry = "same_country_only"
	UserPreferencesMaxDistance = "max_distance_km"
	UserPreferencesUpdatedAt   = "updated_at"
)

type UserPreference struct {
	UserID      int64  `db:"user_id" insert:"user_id"`
	HashKey     int64  `db:"hash_key" insert:"hash_key"`
	MinAge      int    `db:"min_age" insert:"min_age" update:"min_age"`
	MaxAge      int    `db:"max_age" insert:"max_age" update:"max_age"`
	GenderPref  string `db:"gender_preference" insert:"gender_preference" update:"gender_preference"`
	SameCountry bool   `db:"same_country_only" insert:"same_country_only" update:"same_country_only"`
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

func (up *UserPreference) GetPartition() int {
	return int(up.HashKey)
}

type UserPreferencesQuery interface {
	GetByUserID(ctx context.Context, userID int64, hashKey int64) (*UserPreference, error)
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

func (up *userPreferencesQuery) GetByUserID(ctx context.Context, userID int64, hashKey int64) (*UserPreference, error) {
	pref := &UserPreference{}
	qb, args, err := up.sq.Select(pref.columns("")...).
		From(UserPreferencesTable).
		Where(squirrel.Eq{UserPreferencesUserID: userID, UserPreferencesHashKey: hashKey}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return pref, sqlscan.Get(ctx, up.runner, pref, qb, args...)
}
