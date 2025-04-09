-- +goose Up
-- +goose NO TRANSACTION
CREATE INDEX CONCURRENTLY idx_user_prefs_age_gender ON user_preferences(min_age, max_age, gender_preference);
CREATE INDEX CONCURRENTLY idx_user_prefs_gender ON user_preferences(gender_preference);
CREATE INDEX CONCURRENTLY idx_user_prefs_distance ON user_preferences(max_distance_km);
CREATE INDEX CONCURRENTLY idx_user_prefs_updated ON user_preferences(updated_at);

CREATE INDEX CONCURRENTLY idx_user_prefs_composite ON user_preferences(
    gender_preference, 
    min_age, 
    max_age,
    max_distance_km
);


-- +goose Down
-- +goose NO TRANSACTION
DROP INDEX IF EXISTS idx_user_prefs_age_gender;
DROP INDEX IF EXISTS idx_user_prefs_gender;
DROP INDEX IF EXISTS idx_user_prefs_distance;
DROP INDEX IF EXISTS idx_user_prefs_updated;
DROP INDEX IF EXISTS idx_user_prefs_composite;
