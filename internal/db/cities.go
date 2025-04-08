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
	CitiesCountry  = "country"
)

type City struct {
	ID       int64  `db:"id" insert:"id"`
	Name     string `db:"name" insert:"name" update:"name"`
	Location string `db:"location" insert:"location" update:"location"`
	Country  string `db:"country" insert:"country" update:"country"`
}

var (
	stomCitySelect = stom.MustNewStom(City{}).SetTag(selectTag)
	stomCityUpdate = stom.MustNewStom(City{}).SetTag(updateTag)
	stomCityInsert = stom.MustNewStom(City{}).SetTag(insertTag)
)

func (c *City) columns(pref string) []string {
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
	city := &City{}
	qb, args, err := c.sq.Select(city.columns("")...).
		From(Citiestable).
		Where(squirrel.Eq{CitiesID: id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return city, sqlscan.Select(ctx, c.runner, city, qb, args...)
}
