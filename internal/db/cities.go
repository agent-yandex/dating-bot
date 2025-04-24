package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/elgris/stom"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const Citiestable = "cities"

const (
	CitiesID       = "id"
	CitiesName     = "name"
	CitiesLocation = "location"
)

type City struct {
	ID       int64  `db:"id" insert:"id"`
	Name     string `db:"name" insert:"name" update:"name"`
	Location string `db:"location" insert:"location" update:"location"`
}

var (
	stomCitySelect = stom.MustNewStom(City{}).SetTag(selectTag)
	stomCityUpdate = stom.MustNewStom(City{}).SetTag(updateTag)
	stomCityInsert = stom.MustNewStom(City{}).SetTag(insertTag)
)

func (c *City) columns(pref string) []string {
	return colNamesWithPref(stomCitySelect.TagValues(), pref)
}

type CityQuery interface {
	GetByID(ctx context.Context, id int64) (*City, error)
	GetIDByName(ctx context.Context, name string) (int64, error)
	Insert(ctx context.Context, city *City) (*City, error)
}

type cityQuery struct {
	runner *pgxpool.Pool
	sq     squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewCityQuery(runner *pgxpool.Pool, sq squirrel.StatementBuilderType, logger *zap.Logger) CityQuery {
	return &cityQuery{
		runner: runner,
		sq:     sq,
		logger: logger,
	}
}

func (c cityQuery) GetByID(ctx context.Context, id int64) (*City, error) {
	c.logger.Debug("Fetching city by ID", zap.Int64("city_id", id))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	city := &City{}
	qb, args, err := c.sq.Select(city.columns("")...).
		From(Citiestable).
		Where(squirrel.Eq{CitiesID: id}).
		ToSql()
	if err != nil {
		c.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, c.runner, city, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			c.logger.Warn("Database error",
				zap.Int64("city_id", id),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			c.logger.Warn("Failed to fetch city", zap.Int64("city_id", id), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	c.logger.Info("City fetched successfully", zap.Int64("city_id", id))
	return city, nil
}

func (c cityQuery) GetIDByName(ctx context.Context, name string) (int64, error) {
	name = strings.ToLower(name)
	c.logger.Debug("Fetching city ID by name", zap.String("name", name))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var cityID int64
	qb, args, err := c.sq.Select("id").
		From(Citiestable).
		Where(squirrel.Eq{CitiesName: name}).
		ToSql()
	if err != nil {
		c.logger.Error("Failed to build query", zap.Error(err))
		return 0, fmt.Errorf("failed to build query: %w", err)
	}
	err = c.runner.QueryRow(ctx, qb, args...).Scan(&cityID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			c.logger.Warn("Database error",
				zap.String("name", name),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			c.logger.Warn("Failed to fetch city ID", zap.String("name", name), zap.Error(err))
		}
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}
	c.logger.Info("City ID fetched successfully", zap.String("name", name), zap.Int64("city_id", cityID))
	return cityID, nil
}

func (c cityQuery) Insert(ctx context.Context, city *City) (*City, error) {
	c.logger.Debug("Inserting city", zap.String("name", city.Name))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	insertMap, err := stomCityInsert.ToMap(city)
	if err != nil {
		c.logger.Error("Failed to map struct", zap.Error(err))
		return nil, fmt.Errorf("failed to map struct: %w", err)
	}
	qb, args, err := c.sq.Insert(Citiestable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		c.logger.Error("Failed to build query", zap.Error(err))
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	err = pgxscan.Get(ctx, c.runner, city, qb, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			c.logger.Warn("Database error",
				zap.String("name", city.Name),
				zap.String("pg_error_code", pgErr.Code),
				zap.Error(err),
			)
		} else {
			c.logger.Error("Failed to insert city", zap.String("name", city.Name), zap.Error(err))
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	c.logger.Info("City inserted successfully", zap.Int64("city_id", city.ID))
	return city, nil
}
