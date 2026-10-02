package user

import (
	db "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/repository"
)

var _ repository.UserRepository = (*Repository)(nil)

type Repository struct {
	db db.DB
}

func New(database db.DB) *Repository {
	return &Repository{db: database}
}
