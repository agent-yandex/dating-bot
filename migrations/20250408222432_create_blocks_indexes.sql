-- +goose Up
-- +goose NO TRANSACTION
CREATE INDEX CONCURRENTLY idx_blocks_blocker ON blocks(blocker_id);
CREATE INDEX CONCURRENTLY idx_blocks_blocked ON blocks(blocked_id);
CREATE INDEX CONCURRENTLY idx_blocks_created_at ON blocks(created_at);

CREATE INDEX CONCURRENTLY idx_blocks_blocker_blocked ON blocks(blocker_id, blocked_id);


-- +goose Down
-- +goose NO TRANSACTION
DROP INDEX IF EXISTS idx_blocks_blocker;
DROP INDEX IF EXISTS idx_blocks_blocked;
DROP INDEX IF EXISTS idx_blocks_created_at;
DROP INDEX IF EXISTS idx_blocks_blocker_blocked;
