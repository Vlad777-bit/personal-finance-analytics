package transaction

import (
	db "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository"
)

var _ repository.TransactionRepository = (*Repository)(nil)

type Repository struct {
	db db.DB
}

func New(database db.DB) *Repository {
	return &Repository{
		db: database,
	}
}
