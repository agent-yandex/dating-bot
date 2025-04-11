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
	BlocksID        = "id"
	BlocksBlockerID = "blocker_id"
	BlocksBlockedID = "blocked_id"
	BlocksCreatedAt = "created_at"
)

type Block struct {
	ID        int64     `db:"id"`
	BlockerID int64     `db:"blocker_id" insert:"blocker_id"`
	BlockedID int64     `db:"blocked_id" insert:"blocked_id"`
	CreatedAt time.Time `db:"created_at"`
}

var (
	stomBlockSelect = stom.MustNewStom(Block{}).SetTag(selectTag)
	stomBlockInsert = stom.MustNewStom(Block{}).SetTag(insertTag)
)

func (b Block) getTableName() string {
	return BlocksTable
}

func (b Block) columns(pref string) []string {
	return colNamesWithPref(
		stomBlockSelect.TagValues(),
		pref,
	)
}

type BlockQuery interface {
	GetByID(ctx context.Context, id int64) (*Block, error)
	GetAllByBlockerID(ctx context.Context, blockerID int64) ([]*Block, error)
	Insert(ctx context.Context, block *Block) (int64, error)
	Delete(ctx context.Context, block *Block) error
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
	return getByID[Block](ctx, b.runner, b.sq, BlocksID, id)
}

func (b blockQuery) GetAllByBlockerID(ctx context.Context, blockerID int64) ([]*Block, error) {
	var blocks []*Block
	qb, args, err := b.sq.Select((&Block{}).columns("")...).
		From(BlocksTable).
		Where(squirrel.Eq{BlocksBlockerID: blockerID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return blocks, sqlscan.Select(ctx, b.runner, &blocks, qb, args...)
}

func (b blockQuery) Insert(ctx context.Context, block *Block) (int64, error) {
	return insert(ctx, b.runner, b.sq, block)
}

func (b blockQuery) Delete(ctx context.Context, block *Block) error {
	qb, args, err := b.sq.Delete(BlocksTable).
		Where(squirrel.Eq{
			BlocksBlockerID: block.BlockerID,
			BlocksBlockedID: block.BlockedID,
		}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = b.runner.ExecContext(ctx, qb, args...)

	return err
}
