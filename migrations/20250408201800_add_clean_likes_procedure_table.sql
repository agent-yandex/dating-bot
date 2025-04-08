-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE clean_expired_likes()
LANGUAGE plpgsql
AS $$
DECLARE
deleted_rows bigint;
    total_deleted bigint := 0;
    batch_size int := 10000;
BEGIN
FOR i IN 0..15 LOOP
        LOOP
            EXECUTE format('
                WITH deleted AS (
                    DELETE FROM likes_p%s
                    WHERE expires_at < now()
                    LIMIT %s
                    RETURNING *
                ) SELECT count(*) FROM deleted', i, batch_size)
            INTO deleted_rows;

            total_deleted := total_deleted + deleted_rows;
            RAISE NOTICE 'Partition likes_p%: % rows deleted in this batch', i, deleted_rows;
            EXIT WHEN deleted_rows = 0;
END LOOP;
END LOOP;

    RAISE NOTICE 'Total rows deleted: %', total_deleted;
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
