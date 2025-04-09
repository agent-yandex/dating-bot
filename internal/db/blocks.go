package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/elgris/stom"
	"github.com/georgysavva/scany/v2/sqlscan"
)

const BlocksTable = "blocks"

const (
	BlocksID          = "id"
	BlocksBlockerID   = "blocker_id"
	BlocksBlockedID   = "blocked_id"
	BlocksCreatedAt   = "created_at"
)

type Block struct {
	ID          int64     `db:"id" insert:"id"`
	BlockerID   int64     `db:"blocker_id" insert:"blocker_id"`
	BlockedID   int64     `db:"blocked_id" insert:"blocked_id"`
	CreatedAt   time.Time `db:"created_at"`
}

var (
	stomBlockSelect = stom.MustNewStom(Block{}).SetTag(selectTag)
	stomBlockUpdate = stom.MustNewStom(Block{}).SetTag(updateTag)
	stomBlockInsert = stom.MustNewStom(Block{}).SetTag(insertTag)
)

func (b *Block) columns(pref string) []string {
	return colNamesWithPref(
		stomBlockSelect.TagValues(),
		pref,
	)
}

type BlockQuery interface {
	GetByID(ctx context.Context, id int64) (*Block, error)
}

type blockQuery struct {
	runner *sql.DB
	sq     squirrel.StatementBuilderType
}

func NewBlockQuery(runner *sql.DB, sq squirrel.StatementBuilderType) BlockQuery {
	return &blockQuery{
		runner: runner,
		sq:     sq,
	}
}

func (b blockQuery) GetByID(ctx context.Context, id int64) (*Block, error) {
	block := &Block{}
	qb, args, err := b.sq.Select(block.columns("")...).
		From(BlocksTable).
		Where(squirrel.Eq{BlocksID: id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return block, sqlscan.Get(ctx, b.runner, block, qb, args...)
}
