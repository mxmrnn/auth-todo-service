package domain

import "errors"

var (
	ErrNameInvalid          = errors.New("full name must be between 2 and 100 characters")
	ErrInvalidEmail         = errors.New("invalid email format")
	ErrPasswordRequired     = errors.New("password is required")
	ErrPasswordTooWeak      = errors.New("password is too weak")
	ErrUserAlreadyExists    = errors.New("user with this email already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrNonValidUserID       = errors.New("non valid user ID")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenRevoked  = errors.New("refresh token was revoked")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrInvalidRefreshToken  = errors.New("invalid refresh token")
)
