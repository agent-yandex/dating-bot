-- +goose Up
-- +goose StatementBegin
CREATE TABLE blocks (
                        id bigserial,
                        blocker_id bigint NOT NULL,
                        blocked_id bigint NOT NULL,
                        blocker_hash bigint NOT NULL,
                        created_at timestamp with time zone DEFAULT now(),
                        PRIMARY KEY (id, blocker_hash),
                        UNIQUE (blocker_id, blocked_id),
                        CONSTRAINT fk_blocker FOREIGN KEY (blocker_id, blocker_hash)
                            REFERENCES users(id, hash_key) ON DELETE CASCADE,
                        CONSTRAINT fk_blocked FOREIGN KEY (blocked_id)
                            REFERENCES users(id) ON DELETE CASCADE
) PARTITION BY LIST (blocker_hash);

DO $$
BEGIN
FOR i IN 0..15 LOOP
        EXECUTE format('CREATE TABLE blocks_p%s PARTITION OF blocks FOR VALUES IN (%s)', i, i);
EXECUTE format('CREATE INDEX CONCURRENTLY idx_blocks_blocked_id_p%s ON blocks_p%s(blocked_id)', i, i);
END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS blocks CASCADE;
-- +goose StatementEnd
