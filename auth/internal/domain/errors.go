package domain

import "errors"

var (
	ErrEmailRequired       = errors.New("email is required")
	ErrInvalidEmail        = errors.New("email is invalid")
	ErrEmailTooLong        = errors.New("email must contain at most 254 characters")
	ErrPasswordRequired    = errors.New("password is required")
	ErrPasswordTooShort    = errors.New("password must contain at least 8 characters")
	ErrPasswordHashEmpty   = errors.New("password hash is required")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)
