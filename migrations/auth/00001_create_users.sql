-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_email_not_empty
        CHECK (BTRIM(email) <> ''),

    CONSTRAINT users_email_normalized
        CHECK (email = LOWER(BTRIM(email))),

    CONSTRAINT users_email_length_valid
        CHECK (CHAR_LENGTH(email) <= 254),

    CONSTRAINT users_password_hash_not_empty
        CHECK (BTRIM(password_hash) <> ''),

    CONSTRAINT users_email_unique
        UNIQUE (email)
);

-- +goose Down
DROP TABLE users;
