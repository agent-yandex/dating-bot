package db

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/elgris/stom"
	"github.com/georgysavva/scany/v2/sqlscan"
)

const UsersTable = "users"

const (
	UsersID           = "id"
	UsersTelegramID   = "telegram_id"
	UsersUsername     = "username"
	UsersGender       = "gender"
	UsersAge          = "age"
	UsersProfilePhoto = "profile_photo_url"
	UsersCityID       = "city_id"
	UsersBio          = "bio"
	UsersIsActive     = "is_active"
	UsersIsPremium    = "is_premium"
	UsersCreatedAt    = "created_at"
	UsersUpdatedAt    = "updated_at"
	UsersHashKey      = "hash_key"
)

type User struct {
	ID           int64   `db:"id" insert:"id"`
	TelegramID   int64   `db:"telegram_id" insert:"telegram_id"`
	Username     *string `db:"username" insert:"username" update:"username"`
	Gender       string  `db:"gender" insert:"gender" update:"gender"`
	Age          int     `db:"age" insert:"age" update:"age"`
	ProfilePhoto *string `db:"profile_photo_url" insert:"profile_photo_url" update:"profile_photo_url"`
	CityID       *int    `db:"city_id" insert:"city_id" update:"city_id"`
	Bio          *string `db:"bio" insert:"bio" update:"bio"`
	IsActive     bool    `db:"is_active" insert:"is_active" update:"is_active"`
	IsPremium    bool    `db:"is_premium" insert:"is_premium" update:"is_premium"`
	CreatedAt    string  `db:"created_at"`
	UpdatedAt    string  `db:"updated_at"`
	HashKey      int64   `db:"hash_key" insert:"hash_key"`
}

var (
	stomUserSelect = stom.MustNewStom(User{}).SetTag(selectTag)
	stomUserUpdate = stom.MustNewStom(User{}).SetTag(updateTag)
	stomUserInsert = stom.MustNewStom(User{}).SetTag(insertTag)
)

func (u *User) columns(pref string) []string {
	return colNamesWithPref(
		stomUserSelect.TagValues(),
		pref,
	)
}

type UserQuery interface {
	GetByID(ctx context.Context, id int64) (*User, error)
}

type userQuery struct {
	runner *sql.DB
	sq     squirrel.StatementBuilderType
}

func NewUserQuery(runner *sql.DB, sq squirrel.StatementBuilderType) UserQuery {
	return &userQuery{
		runner: runner,
		sq:     sq,
	}
}

func (u userQuery) GetByID(ctx context.Context, id int64) (*User, error) {
	user := &User{}
	qb, args, err := u.sq.Select(user.columns("")...).
		From(UsersTable).
		Where(squirrel.Eq{UsersID: id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return user, sqlscan.Select(ctx, u.runner, user, qb, args...)
}
