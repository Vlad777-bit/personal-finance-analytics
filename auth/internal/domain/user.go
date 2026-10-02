package domain

import (
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

const minimumPasswordLength = 8

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(email, passwordHash string) (User, error) {
	normalizedEmail, err := NormalizeEmail(email)
	if err != nil {
		return User{}, err
	}

	if strings.TrimSpace(passwordHash) == "" {
		return User{}, ErrPasswordHashEmpty
	}

	return User{
		Email:        normalizedEmail,
		PasswordHash: passwordHash,
	}, nil
}

func NormalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	if email == "" {
		return "", ErrEmailRequired
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", ErrInvalidEmail
	}

	return email, nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return ErrPasswordRequired
	}
	if utf8.RuneCountInString(password) < minimumPasswordLength {
		return ErrPasswordTooShort
	}

	return nil
}
