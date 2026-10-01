package report

import (
	db "github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/database"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository"
)

var _ repository.ReportRepository = (*Repository)(nil)

type Repository struct {
	db db.DB
}

func New(database db.DB) *Repository {
	return &Repository{db: database}
}
