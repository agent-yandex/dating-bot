package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/elgris/stom"
	"github.com/georgysavva/scany/v2/sqlscan"
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
	FromUserID int64     `db:"from_user_id" insert:"from_user_id"`
	ToUserID   int64     `db:"to_user_id" insert:"to_user_id"`
	Message    *string   `db:"message" insert:"message"`
	CreatedAt  time.Time `db:"created_at"`
	ExpiresAt  time.Time `db:"expires_at"`
}

var (
	stomLikeSelect = stom.MustNewStom(Like{}).SetTag(selectTag)
	stomLikeInsert = stom.MustNewStom(Like{}).SetTag(insertTag)
)

func (l Like) getTableName() string {
	return LikesTable
}

func (l Like) columns(pref string) []string {
	return colNamesWithPref(
		stomLikeSelect.TagValues(),
		pref,
	)
}

type LikeQuery interface {
	GetByID(ctx context.Context, id int64) (*Like, error)
	GetAllByToUserID(ctx context.Context, userID int64) ([]*Like, error)
	Insert(ctx context.Context, like *Like) (int64, error)
	Delete(ctx context.Context, like *Like) error
}

type likeQuery struct {
	runner *sql.DB
	sq     squirrel.StatementBuilderType
}

func NewLikeQuery(runner *sql.DB, sq squirrel.StatementBuilderType) LikeQuery {
	return &likeQuery{
		runner: runner,
		sq:     sq,
	}
}

func (l likeQuery) GetByID(ctx context.Context, id int64) (*Like, error) {
	return getByID[Like](ctx, l.runner, l.sq, UsersID, id)
}

func (l likeQuery) GetAllByToUserID(ctx context.Context, userID int64) ([]*Like, error) {
	var likes []*Like
	qb, args, err := l.sq.Select((&Like{}).columns("")...).
		From(LikesTable).
		Where(squirrel.Eq{LikesToUserID: userID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return likes, sqlscan.Select(ctx, l.runner, &likes, qb, args...)
}

func (l likeQuery) Insert(ctx context.Context, like *Like) (int64, error) {
	insertMap, err := stomLikeInsert.ToMap(like)
	if err != nil {
		return 0, err
	}
	qb, args, err := l.sq.Insert(like.getTableName()).
		SetMap(insertMap).
		ToSql()
	if err != nil {
		return 0, err
	}
	res, err := l.runner.ExecContext(ctx, qb, args...)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (l likeQuery) Delete(ctx context.Context, like *Like) error {
	qb, args, err := l.sq.Delete(LikesTable).
		Where(squirrel.Eq{
			LikesToUserID:   like.ToUserID,
			LikesFromUserID: like.FromUserID,
		}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = l.runner.ExecContext(ctx, qb, args...)
	return err
}
