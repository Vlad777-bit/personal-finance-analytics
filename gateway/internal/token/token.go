package token

import "context"

type Identity struct {
	UserID string
	Email  string
}

type Verifier interface {
	Verify(ctx context.Context, value string) (Identity, error)
}
