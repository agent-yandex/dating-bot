-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
                       id bigserial PRIMARY KEY,
                       telegram_id bigint UNIQUE NOT NULL,
                       username varchar(32),
                       gender char(1) NOT NULL CHECK (gender IN ('m', 'f')),
                       age integer NOT NULL CHECK (age >= 18 AND age <= 100),
                       profile_photo_url varchar(255),
                       city_id integer REFERENCES cities(id) ON DELETE SET NULL,
                       bio text,
                       is_active boolean DEFAULT TRUE,
                       is_premium boolean DEFAULT FALSE,
                       created_at timestamp with time zone DEFAULT now(),
                       updated_at timestamp with time zone DEFAULT now(),
                       hash_key bigint GENERATED ALWAYS AS (telegram_id % 16) STORED
) PARTITION BY LIST (hash_key);

DO $$
BEGIN
FOR i IN 0..15 LOOP
        EXECUTE format('CREATE TABLE users_p%s PARTITION OF users FOR VALUES IN (%s)', i, i);
EXECUTE format('CREATE INDEX CONCURRENTLY idx_users_gender_age_p%s ON users_p%s(gender, age) WHERE is_active = TRUE', i, i);
EXECUTE format('CREATE INDEX CONCURRENTLY idx_users_city_id_p%s ON users_p%s(city_id)', i, i);
END LOOP;
END;
$$;

CREATE UNIQUE INDEX idx_users_telegram_id ON users(telegram_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users CASCADE;
-- +goose StatementEnd
