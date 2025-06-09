package usecase

import "errors"

var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserNotFound         = errors.New("user not found")
	ErrUserExists           = errors.New("user already exists")
	ErrInvalidCaptcha       = errors.New("invalid captcha")
	ErrTooManyLoginAttempts = errors.New("too many login attempts, please try again later")
	ErrAccountLocked        = errors.New("account is locked due to too many failed attempts")
	ErrLoanNotFound         = errors.New("loan not found")
)
