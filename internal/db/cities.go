package db

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/elgris/stom"
	"github.com/georgysavva/scany/v2/sqlscan"
)

const Citiestable = "cities"

const (
	CitiesID       = "id"
	CitiesName     = "name"
	CitiesLocation = "location"
)

type City struct {
	ID       int64  `db:"id" insert:"id"`
	Name     string `db:"name" insert:"name"`
	Location string `db:"location" insert:"location" update:"location"`
}

var (
	stomCitySelect = stom.MustNewStom(City{}).SetTag(selectTag)
	stomCityInsert = stom.MustNewStom(City{}).SetTag(insertTag)
)

func (c City) getTableName() string {
	return Citiestable
}

func (c City) columns(pref string) []string {
	return colNamesWithPref(
		stomCitySelect.TagValues(),
		pref,
	)
}

type CityQuery interface {
	GetByID(ctx context.Context, id int64) (*City, error)
}

type cityQuery struct {
	runner *sql.DB
	sq     squirrel.StatementBuilderType
}

func NewCityQuery(runner *sql.DB, sq squirrel.StatementBuilderType) UserQuery {
	return &userQuery{
		runner: runner,
		sq:     sq,
	}
}

func (c cityQuery) GetByID(ctx context.Context, id int64) (*City, error) {
	return getByID[City](ctx, c.runner, c.sq, UsersID, id)
}

func (c cityQuery) GetIDByName(ctx context.Context, name string) (*int64, error) {
	var cityID *int64
	qb, args, err := c.sq.Select(CitiesID).
		From(Citiestable).
		Where(squirrel.Eq{CitiesName: name}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return cityID, sqlscan.Get(ctx, c.runner, cityID, qb, args...)
}

func (c cityQuery) Insert(ctx context.Context, city *City) (int64, error) {
	insertMap, err := stomCityInsert.ToMap(city)
	if err != nil {
		return 0, err
	}
	qb, args, err := c.sq.Insert(city.getTableName()).
		SetMap(insertMap).
		ToSql()
	if err != nil {
		return 0, err
	}
	res, err := c.runner.ExecContext(ctx, qb, args...)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}
