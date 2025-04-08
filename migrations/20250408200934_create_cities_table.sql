-- +goose Up
-- +goose StatementBegin
CREATE TABLE cities (
                        id serial PRIMARY KEY,
                        name varchar(100) NOT NULL,
                        location geography(POINT, 4326) NOT NULL,
                        country varchar(50) NOT NULL,
                        CONSTRAINT unique_city_country UNIQUE (name, country)
) WITH (fillfactor = 90);

CREATE INDEX idx_cities_location ON cities USING GIST(location);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cities CASCADE;
-- +goose StatementEnd
