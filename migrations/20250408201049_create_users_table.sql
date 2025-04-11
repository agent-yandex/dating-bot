-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id bigint not null PRIMARY KEY,
    username varchar(32),
    gender char(1) NOT NULL CHECK (gender IN ('m', 'f')),
    age integer NOT NULL CHECK (age >= 10 AND age <= 100),
    profile_photo_url varchar(255),
    city_id integer REFERENCES cities(id) ON DELETE SET NULL,
    bio text,
    is_active boolean DEFAULT TRUE,
    is_premium boolean DEFAULT FALSE,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users CASCADE;
-- +goose StatementEnd
