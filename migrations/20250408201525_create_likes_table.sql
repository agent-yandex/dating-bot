-- +goose Up
-- +goose StatementBegin
CREATE TABLE likes (
                       id bigserial,
                       from_user_id bigint NOT NULL,
                       to_user_id bigint NOT NULL,
                       from_user_hash bigint NOT NULL,
                       message text,
                       created_at timestamp with time zone DEFAULT now(),
                       expires_at timestamp with time zone GENERATED ALWAYS AS (created_at + interval '3 days') STORED,
    PRIMARY KEY (id, from_user_hash),
    CONSTRAINT fk_from_user FOREIGN KEY (from_user_id, from_user_hash)
        REFERENCES users(id, hash_key) ON DELETE CASCADE,
    CONSTRAINT fk_to_user FOREIGN KEY (to_user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (from_user_id, to_user_id)
) PARTITION BY LIST (from_user_hash);

DO $$
BEGIN
FOR i IN 0..15 LOOP
        EXECUTE format('CREATE TABLE likes_p%s PARTITION OF likes FOR VALUES IN (%s)', i, i);
EXECUTE format('CREATE INDEX CONCURRENTLY idx_likes_to_user_p%s ON likes_p%s(to_user_id)', i, i);
EXECUTE format('CREATE INDEX CONCURRENTLY idx_likes_expires_at_p%s ON likes_p%s(expires_at)', i, i);
END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS likes CASCADE;
-- +goose StatementEnd
