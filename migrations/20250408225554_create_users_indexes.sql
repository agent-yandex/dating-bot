-- +goose Up
-- +goose NO TRANSACTION
CREATE INDEX CONCURRENTLY idx_users_gender_age_active ON users(gender, age) WHERE is_active = TRUE;
CREATE INDEX CONCURRENTLY idx_users_city_id ON users(city_id);
CREATE INDEX CONCURRENTLY idx_users_is_active ON users(is_active);


CREATE INDEX CONCURRENTLY idx_users_age ON users(age);
CREATE INDEX CONCURRENTLY idx_users_created_at ON users(created_at);
CREATE INDEX CONCURRENTLY idx_users_is_premium ON users(is_premium);

CREATE INDEX CONCURRENTLY idx_users_gender_age_city ON users(gender, age, city_id) WHERE is_active = TRUE;

-- +goose Down
-- +goose NO TRANSACTION
DROP INDEX IF EXISTS idx_users_gender_age_active;
DROP INDEX IF EXISTS idx_users_city_id;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_age;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_is_premium;
DROP INDEX IF EXISTS idx_users_gender_age_city;
