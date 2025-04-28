-- +goose Up
-- +goose StatementBegin
CREATE TABLE cities (
                        id serial PRIMARY KEY,
                        name varchar(200) NOT NULL,
                        location geography(POINT, 4326) NOT NULL
) WITH (fillfactor = 90);

CREATE INDEX idx_cities_location ON cities USING GIST(location);

CREATE TABLE temp_geonames (
                               geonameid int,
                               name varchar(200),
                               asciiname varchar(200),
                               alternatenames text,
                               latitude float,
                               longitude float,
                               feature_class char(1),
                               feature_code varchar(10),
                               country_code char(2),
                               cc2 varchar(200),
                               admin1_code varchar(20),
                               admin2_code varchar(80),
                               admin3_code varchar(20),
                               admin4_code varchar(20),
                               population bigint,
                               elevation int,
                               dem int,
                               timezone varchar(40),
                               modification_date date
);

COPY temp_geonames (
    geonameid, name, asciiname, alternatenames, latitude, longitude,
    feature_class, feature_code, country_code, cc2, admin1_code,
    admin2_code, admin3_code, admin4_code, population, elevation,
    dem, timezone, modification_date
    ) FROM '/RU.txt' DELIMITER E'\t' NULL '';

INSERT INTO cities (name, location)
SELECT
    SPLIT_PART(lower(alternatenames), ',', -1) AS alt,
    ST_GeogFromText('POINT(' || longitude || ' ' || latitude || ')') AS location
FROM temp_geonames
WHERE feature_class = 'P'
  AND country_code = 'RU'
  AND alternatenames IS NOT NULL
  AND latitude IS NOT NULL
  AND longitude IS NOT NULL
  AND latitude BETWEEN -90 AND 90
  AND longitude BETWEEN -180 AND 180;

DROP TABLE temp_geonames;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cities CASCADE;
-- +goose StatementEnd