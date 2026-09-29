-- +goose Up
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    amount BIGINT NOT NULL,
    category VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT transactions_amount_positive
        CHECK (amount > 0),

    CONSTRAINT transactions_category_not_empty
        CHECK (BTRIM(category) <> '')
);

CREATE INDEX transactions_user_occurred_at_idx
    ON transactions (user_id, occurred_at);

CREATE INDEX transactions_user_category_occurred_at_idx
    ON transactions (user_id, category, occurred_at);

-- +goose Down
DROP TABLE transactions;