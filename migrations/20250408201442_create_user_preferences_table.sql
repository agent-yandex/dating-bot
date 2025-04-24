-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_preferences (
    user_id bigint NOT NULL PRIMARY KEY,
    min_age integer NOT NULL DEFAULT 10 CHECK (min_age >= 10),
    max_age integer NOT NULL DEFAULT 99 CHECK (max_age <= 100),
    gender_preference char(1) NOT NULL default 'a' CHECK (gender_preference IN ('m', 'f', 'a')),
    max_distance_km integer DEFAULT 100 CHECK (max_distance_km > 0),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT fk_user_prefs FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_preferences CASCADE;
-- +goose StatementEnd
