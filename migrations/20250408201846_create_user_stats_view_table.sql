-- +goose Up
-- +goose StatementBegin
CREATE MATERIALIZED VIEW user_stats AS
SELECT
    gender,
    FLOOR(age / 10) * 10 AS age_group,
    COUNT(*) AS user_count,
    COUNT(*) FILTER (WHERE is_premium) AS premium_count
FROM users
         JOIN cities ON users.city_id = cities.id
WHERE is_active = TRUE
GROUP BY gender, FLOOR(age / 10) * 10
    WITH DATA;

SELECT cron.schedule(
               'refresh-user-stats',
               '0 2 * * *',
               'REFRESH MATERIALIZED VIEW CONCURRENTLY user_stats'
       );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS user_stats;
SELECT cron.unschedule('refresh-user-stats');
-- +goose StatementEnd
