-- +goose Up
CREATE TABLE refresh_sessions (
    token_hash TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,

    CONSTRAINT refresh_sessions_token_hash_not_empty
        CHECK (BTRIM(token_hash) <> ''),
    CONSTRAINT refresh_sessions_expiry_valid
        CHECK (expires_at > created_at)
);

CREATE INDEX refresh_sessions_user_id_idx ON refresh_sessions (user_id);
CREATE INDEX refresh_sessions_active_idx
    ON refresh_sessions (token_hash, expires_at)
    WHERE revoked_at IS NULL;

-- +goose Down
DROP TABLE refresh_sessions;
