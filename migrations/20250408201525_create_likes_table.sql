-- +goose Up
-- +goose StatementBegin
CREATE TABLE likes (
    id bigserial PRIMARY KEY,
    from_user_id bigint NOT NULL,
    to_user_id bigint NOT NULL,
    message text,
    created_at timestamp with time zone DEFAULT now(),
    expires_at timestamp with time zone DEFAULT (now() + interval '3 days'),
    CONSTRAINT fk_from_user FOREIGN KEY (from_user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_to_user FOREIGN KEY (to_user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (from_user_id, to_user_id)
);

CREATE OR REPLACE FUNCTION set_like_expires_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.expires_at := NEW.created_at + interval '3 days';
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_like_expires_at
BEFORE INSERT ON likes
FOR EACH ROW EXECUTE FUNCTION set_like_expires_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_like_expires_at ON likes;
DROP FUNCTION IF EXISTS set_like_expires_at();
DROP TABLE IF EXISTS likes CASCADE;
-- +goose StatementEnd