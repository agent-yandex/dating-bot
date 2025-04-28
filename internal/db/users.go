package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/elgris/stom"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
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
	UsersRating       = "rating"
	UsersCreatedAt    = "created_at"
	UsersUpdatedAt    = "updated_at"
	TgUsername        = "tg_username"
)

type User struct {
	ID           int64     `db:"id" insert:"id"`
	Username     *string   `db:"username" insert:"username" update:"username"`
	TgUsername   string    `db:"tg_username" insert:"tg_username"`
	Gender       string    `db:"gender" insert:"gender" update:"gender"`
	Age          int       `db:"age" insert:"age" update:"age"`
	ProfilePhoto *string   `db:"profile_photo_url" insert_photo:"profile_photo_url" update_photo:"profile_photo_url"`
	CityID       *int64    `db:"city_id" insert:"city_id" update:"city_id"`
	Bio          *string   `db:"bio" insert:"bio" update:"bio"`
	IsActive     bool      `db:"is_active" insert_active:"is_active" update_active:"is_active"`
	IsPremium    bool      `db:"is_premium"`
	Rating       int       `db:"rating"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

var (
	stomUserSelect = stom.MustNewStom(User{}).SetTag(selectTag)
	stomUserInsert = stom.MustNewStom(User{}).SetTag(insertTag)
	stomUserUpdate = stom.MustNewStom(User{}).SetTag(updateTag)
)

func (u *User) columns(pref string) []string {
	return colNamesWithPref(stomUserSelect.TagValues(), pref)
}

type UserQuery interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	Insert(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User, id int64) (*User, error)
	UpdateProfilePhoto(ctx context.Context, id int64, profilePhoto *string) error
	UpdateActive(ctx context.Context, id int64, isActive bool) error
	UpdateRating(ctx context.Context, id int64) error
	SelectUsers(ctx context.Context, id int64, offset uint64) ([]*User, error)
	Delete(ctx context.Context, id int64) error
}

type userQuery struct {
	runner *pgxpool.Pool
	sq     squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewUserQuery(runner *pgxpool.Pool, sq squirrel.StatementBuilderType, logger *zap.Logger) UserQuery {
	return &userQuery{
		runner: runner,
		sq:     sq,
		logger: logger,
	}
}

func (u userQuery) GetByID(ctx context.Context, id int64) (*User, error) {
	u.logger.Debug("Fetching user by ID", zap.Int64("user_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user := &User{}
	qb, args, err := u.sq.Select(user.columns("")...).
		From(UsersTable).
		Where(squirrel.Eq{UsersID: id}).
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	err = pgxscan.Get(ctx, u.runner, user, qb, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Int64("user_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Warn("Failed to fetch user", zap.Int64("user_id", id), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	u.logger.Info("User fetched successfully", zap.Int64("user_id", id))
	return user, nil
}

func (u userQuery) Insert(ctx context.Context, user *User) (*User, error) {
	u.logger.Debug("Inserting user", zap.Any("user", user))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	insertMap, err := stomUserInsert.ToMap(user)
	if err != nil {
		u.logger.Error("Failed to map struct", zap.Error(err))
		return nil, fmt.Errorf("failed to map struct: %w", err)
	}
	qb, args, err := u.sq.Insert(UsersTable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, u.runner, user, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Any("user", user),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to insert user", zap.Any("user", user), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	u.logger.Info("User inserted successfully", zap.Int64("user_id", user.ID))
	return user, nil
}

func (u userQuery) Update(ctx context.Context, user *User, id int64) (*User, error) {
	u.logger.Debug("Updating user", zap.Int64("user_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	updateMap, err := stomUserUpdate.ToMap(user)
	if err != nil {
		u.logger.Error("Failed to map struct", zap.Error(err))
		return nil, fmt.Errorf("failed to map struct: %w", err)
	}
	qb, args, err := u.sq.Update(UsersTable).
		SetMap(updateMap).
		Where(squirrel.Eq{UsersID: id}).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, u.runner, user, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Int64("user_id", user.ID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to update user", zap.Int64("user_id", user.ID), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	u.logger.Info("User updated successfully", zap.Int64("user_id", user.ID))
	return user, nil
}

func (u userQuery) UpdateProfilePhoto(ctx context.Context, id int64, profilePhoto *string) error {
	u.logger.Debug("Updating user profile photo", zap.Int64("user_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	updateMap := map[string]interface{}{
		UsersProfilePhoto: profilePhoto,
	}
	qb, args, err := u.sq.Update(UsersTable).
		SetMap(updateMap).
		Where(squirrel.Eq{UsersID: id}).
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return fmt.Errorf("failed to build query: %w", err)
	}
	_, err = u.runner.Exec(ctx, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Int64("user_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to update profile photo", zap.Int64("user_id", id), zap.Error(err))
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}
	u.logger.Info("Profile photo updated successfully", zap.Int64("user_id", id))
	return nil
}

func (u userQuery) UpdateActive(ctx context.Context, id int64, isActive bool) error {
	u.logger.Debug("Updating user active status",
		zap.Int64("user_id", id),
		zap.Bool("is_active", isActive),
	)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	updateMap := map[string]interface{}{
		UsersIsActive: isActive,
	}
	qb, args, err := u.sq.Update(UsersTable).
		SetMap(updateMap).
		Where(squirrel.Eq{UsersID: id}).
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return fmt.Errorf("failed to build query: %w", err)
	}
	_, err = u.runner.Exec(ctx, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Int64("user_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to update active status", zap.Int64("user_id", id), zap.Error(err))
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}
	u.logger.Info("Active status updated successfully", zap.Int64("user_id", id))
	return nil
}

func (u userQuery) SelectUsers(ctx context.Context, id int64, offset uint64) ([]*User, error) {
	u.logger.Debug("Selecting users",
		zap.Int64("user_id", id),
		zap.Uint64("offset", offset),
	)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var users []*User

	qb := u.sq.Select((&User{}).columns("u")...).
		From(UsersTable+" u").
		InnerJoin("cities c ON u.city_id = c.id").
		InnerJoin(UserPreferencesTable+" up_own ON up_own.user_id = ?", id).
		InnerJoin(UsersTable+" u_own ON u_own.id = ?", id).
		InnerJoin("cities c_own ON u_own.city_id = c_own.id").
		LeftJoin(BlocksTable+" b1 ON u.id = b1.blocked_id AND b1.blocker_id = ?", id).
		LeftJoin(BlocksTable+" b2 ON u.id = b2.blocker_id AND b2.blocked_id = ?", id).
		LeftJoin(LikesTable+" l ON l.from_user_id = ? AND l.to_user_id = u.id", id).
		Where(squirrel.And{
			squirrel.NotEq{"u.id": id},
			squirrel.Eq{"u.is_active": true},
			squirrel.Eq{"b1.id": nil},
			squirrel.Eq{"b2.id": nil},
			squirrel.Eq{"l.id": nil},
			squirrel.Expr("u.age >= up_own.min_age"),
			squirrel.Expr("u.age <= up_own.max_age"),
			squirrel.Or{
				squirrel.Eq{"up_own.gender_preference": "a"},
				squirrel.Expr("u.gender = up_own.gender_preference"),
			},
			squirrel.Expr("ST_Distance(c.location, c_own.location) <= up_own.max_distance_km * 1000"),
		}).
		OrderBy("u.rating DESC, u.id").
		Limit(50).
		Offset(offset)

	query, args, err := qb.ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	err = pgxscan.Select(ctx, u.runner, &users, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Int64("user_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to select users", zap.Int64("user_id", id), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	u.logger.Info("Users selected successfully",
		zap.Int64("user_id", id),
		zap.Int("count", len(users)),
	)
	return users, nil
}

func (u userQuery) Delete(ctx context.Context, id int64) error {
	u.logger.Debug("Deleting user", zap.Int64("user_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	qb, args, err := u.sq.Delete(UsersTable).
		Where(squirrel.Eq{UsersID: id}).
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := u.runner.Exec(ctx, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Int64("user_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to delete user", zap.Int64("user_id", id), zap.Error(err))
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		u.logger.Warn("No user found to delete", zap.Int64("user_id", id))
		return fmt.Errorf("no user found with id %d", id)
	}

	u.logger.Info("User deleted successfully", zap.Int64("user_id", id))
	return nil
}

func (u userQuery) UpdateRating(ctx context.Context, userID int64) error {
	u.logger.Debug("Updating user rating", zap.Int64("user_id", userID))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var rating int
	query := `
		SELECT COUNT(*)
		FROM likes
		WHERE to_user_id = $1
		AND expires_at > NOW()
	`
	err := u.runner.QueryRow(ctx, query, userID).Scan(&rating)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Int64("user_id", userID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to count likes for rating", zap.Int64("user_id", userID), zap.Error(err))
		}
		return fmt.Errorf("failed to count likes: %w", err)
	}

	// Обновляем рейтинг в таблице users
	updateMap := map[string]interface{}{
		UsersRating: rating,
	}
	qb, args, err := u.sq.Update(UsersTable).
		SetMap(updateMap).
		Where(squirrel.Eq{UsersID: userID}).
		ToSql()
	if err != nil {
		u.logger.Error("Failed to build query", zap.Error(err))
		return fmt.Errorf("failed to build query: %w", err)
	}
	_, err = u.runner.Exec(ctx, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			u.logger.Warn("Database error",
				zap.Int64("user_id", userID),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			u.logger.Error("Failed to update rating", zap.Int64("user_id", userID), zap.Error(err))
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}

	u.logger.Info("Rating updated successfully", zap.Int64("user_id", userID), zap.Int("rating", rating))
	return nil
}
