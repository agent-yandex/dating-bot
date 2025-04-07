package db

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/elgris/stom"
	"github.com/georgysavva/scany/v2/sqlscan"
)

const ProfilesTable = "profiles"

const (
	ProfileID          = "id"
	ProfileName        = "name"
	ProfileAge         = "age"
	ProfileDescription = "description"
	ProfileImageURL    = "image_url"
)

type Profile struct {
	ID       string `db:"id" insert:"id"`
	Name     string `db:"name" insert:"name" update:"name"`
	Age      int64  `db:"age" insert:"age" update:"age"`
	ImageURL string `db:"image_url" insert:"image_url" update:"image_url"`
}

var (
	stomProfileSelect = stom.MustNewStom(Profile{}).SetTag(selectTag)
	stomProfileUpdate = stom.MustNewStom(Profile{}).SetTag(updateTag)
	stomProfileInsert = stom.MustNewStom(Profile{}).SetTag(insertTag)
)

func (p *Profile) columns(pref string) []string {
	return colNamesWithPref(
		stomProfileSelect.TagValues(),
		pref,
	)
}

type ProfilesQuery interface {
	GetByID(ctx context.Context, id int64) (*Profile, error)
}

type profilesQuery struct {
	runner *sql.DB
	sq     squirrel.StatementBuilderType
}

func NewProfilesQuery(runner *sql.DB, sq squirrel.StatementBuilderType) ProfilesQuery {
	return &profilesQuery{
		runner: runner,
		sq:     sq,
	}
}

func (p profilesQuery) GetByID(ctx context.Context, id int64) (*Profile, error) {
	profile := &Profile{}
	qb, args, err := p.sq.Select(profile.columns("")...).
		From(ProfilesTable).
		Where(squirrel.Eq{ProfileID: id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return profile, sqlscan.Select(ctx, p.runner, profile, qb, args...)
}
