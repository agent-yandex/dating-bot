-- +goose Up
-- +goose StatementBegin
CREATE TABLE blocks (
    id bigserial PRIMARY KEY,
    blocker_id bigint NOT NULL,
    blocked_id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    UNIQUE (blocker_id, blocked_id),
    CONSTRAINT fk_blocker FOREIGN KEY (blocker_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_blocked FOREIGN KEY (blocked_id)
        REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS blocks CASCADE;
-- +goose StatementEnd
