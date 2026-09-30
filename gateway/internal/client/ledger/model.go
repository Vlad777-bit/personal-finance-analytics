package ledger

import "time"

type CreateTransactionInput struct {
	UserID      string
	Amount      int64
	Category    string
	Description string
	OccurredAt  time.Time
}

type Transaction struct {
	ID          string
	UserID      string
	Amount      int64
	Category    string
	Description string
	OccurredAt  time.Time
	CreatedAt   time.Time
}

type CreateBudgetInput struct {
	UserID   string
	Category string
	Limit    int64
}

type Budget struct {
	ID       string
	UserID   string
	Category string
	Limit    int64
}
