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
	LikesID           = "id"
	LikesFromUserID   = "from_user_id"
	LikesToUserID     = "to_user_id"
	LikesFromUserHash = "from_user_hash"
	LikesMessage      = "message"
	LikesCreatedAt    = "created_at"
	LikesExpiresAt    = "expires_at"
)

type Like struct {
	ID           int64     `db:"id" insert:"id"`
	FromUserID   int64     `db:"from_user_id" insert:"from_user_id"`
	ToUserID     int64     `db:"to_user_id" insert:"to_user_id"`
	FromUserHash int64     `db:"from_user_hash" insert:"from_user_hash"`
	Message      *string   `db:"message" insert:"message" update:"message"`
	CreatedAt    time.Time `db:"created_at"`
	ExpiresAt    time.Time `db:"expires_at"`
}

var (
	stomLikeSelect = stom.MustNewStom(Like{}).SetTag(selectTag)
	stomLikeUpdate = stom.MustNewStom(Like{}).SetTag(updateTag)
	stomLikeInsert = stom.MustNewStom(Like{}).SetTag(insertTag)
)

func (l *Like) columns(pref string) []string {
	return colNamesWithPref(
		stomLikeSelect.TagValues(),
		pref,
	)
}

type LikeQuery interface {
	GetByID(ctx context.Context, id int64, fromUserHash int64) (*Like, error)
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

func (l likeQuery) GetByID(ctx context.Context, id int64, fromUserHash int64) (*Like, error) {
	like := &Like{}
	qb, args, err := l.sq.Select(like.columns("")...).
		From(LikesTable).
		Where(squirrel.Eq{LikesID: id, LikesFromUserHash: fromUserHash}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return like, sqlscan.Get(ctx, l.runner, like, qb, args...)
}
