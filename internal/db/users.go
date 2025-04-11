package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/elgris/stom"
	"github.com/georgysavva/scany/v2/sqlscan"
)

const UsersTable = "users"

const (
	UsersID           = "id"
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
)

type User struct {
	ID           int64   `db:"id" insert:"id"`
	Username     *string `db:"username" insert:"username" update:"username"`
	Gender       string  `db:"gender" insert:"gender" update:"gender"`
	Age          int     `db:"age" insert:"age" update:"age"`
	ProfilePhoto *string `db:"profile_photo_url" insert_photo:"profile_photo_url" update_photo:"profile_photo_url"`
	CityID       *int    `db:"city_id" insert:"city_id" update:"city_id"`
	Bio          *string `db:"bio" insert:"bio" update:"bio"`
	IsActive     bool    `db:"is_active" insert_active:"is_active" update_active:"is_active"`
	IsPremium    bool    `db:"is_premium"`
	CreatedAt    string  `db:"created_at"`
	UpdatedAt    string  `db:"updated_at"`
}

var (
	stomUserSelect = stom.MustNewStom(User{}).SetTag(selectTag)
	stomUserInsert = stom.MustNewStom(User{}).SetTag(insertTag)
	stomUserUpdate = stom.MustNewStom(User{}).SetTag(updateTag)
)

func (u User) getTableName() string {
	return UsersTable
}

func (u User) columns(pref string) []string {
	return colNamesWithPref(
		stomUserSelect.TagValues(),
		pref,
	)
}

type UserQuery interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	Insert(ctx context.Context, user *User) (int64, error)
	Update(ctx context.Context, user *User) error
	UpdateProfilePhoto(ctx context.Context, id int64, profilePhoto *string) error
	UpdateActive(ctx context.Context, id int64, isActive bool) error
	SelectUsers(ctx context.Context, id int64, offset uint64) ([]*User, error)
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

func (u *userQuery) GetByID(ctx context.Context, id int64) (*User, error) {
	return getByID[User](ctx, u.runner, u.sq, UsersID, id)
}

func (u *userQuery) Insert(ctx context.Context, user *User) (int64, error) {
	return insert(ctx, u.runner, u.sq, user)
}

func (u *userQuery) Update(ctx context.Context, user *User) error {
	updateMap, err := stomUserUpdate.ToMap(user)
	if err != nil {
		return err
	}
	qb, args, err := u.sq.Update(UsersTable).
		SetMap(updateMap).
		Where(squirrel.Eq{UsersID: user.ID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = u.runner.ExecContext(ctx, qb, args...)

	return err
}

func (u *userQuery) UpdateProfilePhoto(ctx context.Context, id int64, profilePhoto *string) error {
	updateMap := map[string]interface{}{
		UsersProfilePhoto: profilePhoto,
	}
	qb, args, err := u.sq.Update(UsersTable).
		SetMap(updateMap).
		Where(squirrel.Eq{UsersID: id}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = u.runner.ExecContext(ctx, qb, args...)

	return err
}

func (u *userQuery) UpdateActive(ctx context.Context, id int64, isActive bool) error {
	updateMap := map[string]interface{}{
		UsersIsActive: isActive,
	}
	qb, args, err := u.sq.Update(UsersTable).
		SetMap(updateMap).
		Where(squirrel.Eq{UsersID: id}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = u.runner.ExecContext(ctx, qb, args...)

	return err
}

// SelectUsers retrieves users matching the requesting user's preferences, excluding blocked users
// and respecting age, gender, and distance constraints. Returns up to 50 users per page.
// Assumes user_preferences and city_id exist for the requesting user.
func (u *userQuery) SelectUsers(ctx context.Context, id int64, offset uint64) ([]*User, error) {
	/*
		SELECT u.id, u.username, u.gender, u.age, u.profile_photo_url,
		       u.city_id, u.bio, u.is_active, u.is_premium, u.created_at, u.updated_at
		FROM users u
		INNER JOIN cities c ON u.city_id = c.id
		INNER JOIN user_preferences up_own ON up_own.user_id = $1
		INNER JOIN users u_own ON u_own.id = $1
		INNER JOIN cities c_own ON u_own.city_id = c_own.id
		LEFT JOIN blocks b1 ON u.id = b1.blocked_id AND b1.blocker_id = $1
		LEFT JOIN blocks b2 ON u.id = b2.blocker_id AND b2.blocked_id = $1
		WHERE u.id != $1
		  AND u.is_active = true
		  AND b1.id IS NULL
		  AND b2.id IS NULL
		  AND u.age >= up_own.min_age
		  AND u.age <= up_own.max_age
		  AND (up_own.gender_preference = 'a' OR u.gender = up_own.gender_preference)
		  AND ST_Distance(c.location, c_own.location) <= up_own.max_distance_km * 1000
		ORDER BY u.id
		LIMIT 50 OFFSET $2
	*/
	var users []*User

	qb := u.sq.Select((&User{}).columns("u")...).
		From(UsersTable+" u").
		InnerJoin("cities c ON u.city_id = c.id").
		InnerJoin(UserPreferencesTable+" up_own ON up_own.user_id = ?", id).
		InnerJoin(UsersTable+" u_own ON u_own.id = ?", id).
		InnerJoin("cities c_own ON u_own.city_id = c_own.id").
		LeftJoin(BlocksTable+" b1 ON u.id = b1.blocked_id AND b1.blocker_id = ?", id).
		LeftJoin(BlocksTable+" b2 ON u.id = b2.blocker_id AND b2.blocked_id = ?", id).
		Where(squirrel.And{
			squirrel.NotEq{"u.id": id},
			squirrel.Eq{"u.is_active": true},
			squirrel.Eq{"b1.id": nil},
			squirrel.Eq{"b2.id": nil},
			squirrel.GtOrEq{"u.age": squirrel.Expr("up_own.min_age")},
			squirrel.LtOrEq{"u.age": squirrel.Expr("up_own.max_age")},
			squirrel.Or{
				squirrel.Eq{"up_own.gender_preference": "a"},
				squirrel.Eq{"u.gender": squirrel.Expr("up_own.gender_preference")},
			},
			squirrel.LtOrEq{
				"ST_Distance(c.location, c_own.location)": squirrel.Expr("up_own.max_distance_km * 1000"),
			},
		}).
		OrderBy("u.id").
		Limit(50).
		Offset(offset)

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	return users, sqlscan.Select(ctx, u.runner, &users, query, args...)
}
