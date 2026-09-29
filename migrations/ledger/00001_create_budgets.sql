-- +goose Up
CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    category VARCHAR(100) NOT NULL,
    limit_amount BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT budgets_category_not_empty
        CHECK (BTRIM(category) <> ''),

    CONSTRAINT budgets_limit_positive
        CHECK (limit_amount > 0),

    CONSTRAINT budgets_user_category_unique
        UNIQUE (user_id, category)
);

-- +goose Down
DROP TABLE budgets;