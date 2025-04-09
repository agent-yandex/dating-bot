-- +goose Up
-- +goose StatementBegin
-- Создание основной таблицы cities
CREATE TABLE cities (
    id serial PRIMARY KEY,
    name varchar(10000) NOT NULL,  -- Увеличен размер до 300 символов
    location geography(POINT, 4326) NOT NULL
) WITH (fillfactor = 90);

-- Создание GIST-индекса для поля location
CREATE INDEX idx_cities_location ON cities USING GIST(location);

-- Создание временной таблицы для импорта данных из GeoNames RU
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

-- Импорт данных из файла RU.txt
COPY temp_geonames (
    geonameid, name, asciiname, alternatenames, latitude, longitude,
    feature_class, feature_code, country_code, cc2, admin1_code,
    admin2_code, admin3_code, admin4_code, population, elevation,
    dem, timezone, modification_date
) FROM '/RU.txt' DELIMITER E'\t' NULL '';

-- Перенос только городов (feature_class = 'P') в таблицу cities
INSERT INTO cities (name, location)
SELECT 
    RTRIM(REGEXP_REPLACE(SPLIT_PART(alternatenames, ',', -1), '[()]', '', 'g')) AS alt,
    ST_GeogFromText('POINT(' || longitude || ' ' || latitude || ')') AS location
FROM temp_geonames
WHERE feature_class = 'P'
AND country_code = 'RU'
AND alternatenames IS NOT NULL
AND latitude IS NOT NULL
AND longitude IS NOT NULL;

-- Удаление временной таблицы
DROP TABLE temp_geonames;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cities CASCADE;
-- +goose StatementEnd