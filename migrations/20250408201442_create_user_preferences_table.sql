-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_preferences (
                                  user_id bigint NOT NULL,
                                  hash_key bigint NOT NULL,
                                  min_age integer NOT NULL DEFAULT 10 CHECK (min_age >= 10),
                                  max_age integer NOT NULL DEFAULT 99 CHECK (max_age <= 100),
                                  gender_preference char(1) NOT NULL CHECK (gender_preference IN ('m', 'f', 'a')),
                                  same_country_only boolean DEFAULT TRUE,
                                  max_distance_km integer DEFAULT 50 CHECK (max_distance_km > 0),
                                  updated_at timestamp with time zone DEFAULT now(),
                                  PRIMARY KEY (user_id, hash_key),
                                  CONSTRAINT fk_user_prefs FOREIGN KEY (user_id, hash_key)
                                      REFERENCES users(id, hash_key) ON DELETE CASCADE
) PARTITION BY LIST (hash_key);

DO $$
BEGIN
FOR i IN 0..15 LOOP
        EXECUTE format('CREATE TABLE user_prefs_p%s PARTITION OF user_preferences FOR VALUES IN (%s)', i, i);
EXECUTE format('CREATE INDEX CONCURRENTLY idx_user_prefs_age_gender_p%s ON user_prefs_p%s(min_age, max_age, gender_preference)', i, i);
END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_preferences CASCADE;
-- +goose StatementEnd
