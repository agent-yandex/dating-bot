-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE clean_expired_likes() 
LANGUAGE plpgsql 
AS $$
DECLARE 
    deleted_rows bigint;
BEGIN
    WITH deleted AS (
        DELETE FROM likes
        WHERE expires_at < now()
        RETURNING *
    ) 
    SELECT count(*) INTO deleted_rows FROM deleted;

    RAISE NOTICE 'Total rows deleted: %', deleted_rows;
    COMMIT;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Error in clean_expired_likes: %', SQLERRM;
        ROLLBACK;
END;
$$;

SELECT cron.schedule(
    'clean-expired-likes',
    '0 2 * * *',
    'CALL clean_expired_likes()'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP PROCEDURE IF EXISTS clean_expired_likes;
SELECT cron.unschedule('clean-expired-likes');
-- +goose StatementEnd
