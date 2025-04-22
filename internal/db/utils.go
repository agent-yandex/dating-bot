package db

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
)

const (
	insertTag = "insert"
	selectTag = "db"
	updateTag = "update"
	deleteTag = "delete"
)

func colNamesWithPref(cols []string, pref string) []string {
	prefCols := make([]string, len(cols))
	copy(prefCols, cols)
	sort.Strings(prefCols)
	if pref == "" {
		return prefCols
	}

	for i := range prefCols {
		if !strings.Contains(prefCols[i], ".") {
			prefCols[i] = fmt.Sprintf("%s.%s", pref, prefCols[i])
		}
	}
	return prefCols
}

type dbElement interface {
	columns(pref string) []string
	getTableName() string
}

func getByID[T dbElement](ctx context.Context, runner *sql.DB, sq squirrel.StatementBuilderType, columnID string, id int64) (*T, error) {
	var result T
	qb, args, err := sq.Select(result.columns("")...).
		From(result.getTableName()).
		Where(squirrel.Eq{columnID: id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return &result, sqlscan.Get(ctx, runner, &result, qb, args...)
}
