-- +goose Up
-- +goose NO TRANSACTION
CREATE INDEX CONCURRENTLY idx_likes_from_user ON likes(from_user_id);
CREATE INDEX CONCURRENTLY idx_likes_to_user ON likes(to_user_id);
CREATE INDEX CONCURRENTLY idx_likes_expires_at ON likes(expires_at);

CREATE INDEX CONCURRENTLY idx_likes_from_to_user ON likes(from_user_id, to_user_id);
CREATE INDEX CONCURRENTLY idx_likes_created_at ON likes(created_at);


-- +goose Down
-- +goose NO TRANSACTION
DROP INDEX IF EXISTS idx_likes_from_user;
DROP INDEX IF EXISTS idx_likes_to_user;
DROP INDEX IF EXISTS idx_likes_expires_at;
DROP INDEX IF EXISTS idx_likes_from_to_user;
DROP INDEX IF EXISTS idx_likes_created_at;
