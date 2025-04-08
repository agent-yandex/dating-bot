-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
                       id bigint not null,
                       telegram_id bigint NOT NULL,
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
                       hash_key bigint NOT NULL,
                       PRIMARY KEY (id, hash_key),
                       UNIQUE (telegram_id, hash_key)  -- Добавляем hash_key в UNIQUE ограничение
) PARTITION BY LIST (hash_key);

-- Триггер для автоматического вычисления hash_key
CREATE OR REPLACE FUNCTION set_user_hash()
RETURNS TRIGGER AS $$
BEGIN
    NEW.hash_key := NEW.telegram_id % 16;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_user_hash
    BEFORE INSERT OR UPDATE ON users
                         FOR EACH ROW EXECUTE FUNCTION set_user_hash();

DO $$
BEGIN
FOR i IN 0..15 LOOP
        EXECUTE format('CREATE TABLE users_p%s PARTITION OF users FOR VALUES IN (%s)', i, i);
EXECUTE format('CREATE INDEX CONCURRENTLY idx_users_gender_age_p%s ON users_p%s(gender, age) WHERE is_active = TRUE', i, i);
EXECUTE format('CREATE INDEX CONCURRENTLY idx_users_city_id_p%s ON users_p%s(city_id)', i, i);
END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users CASCADE;
-- +goose StatementEnd
