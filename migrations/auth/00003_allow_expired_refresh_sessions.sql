-- +goose Up
ALTER TABLE refresh_sessions
    DROP CONSTRAINT refresh_sessions_expiry_valid;

-- +goose Down
ALTER TABLE refresh_sessions
    ADD CONSTRAINT refresh_sessions_expiry_valid
    CHECK (expires_at > created_at);
