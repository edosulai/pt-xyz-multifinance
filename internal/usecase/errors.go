package usecase

import "errors"

// Error types
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

type ConflictError struct {
	Message string
}

func (e ConflictError) Error() string {
	return e.Message
}

func NewValidationError(msg string) error {
	return ValidationError{Message: msg}
}

func NewConflictError(msg string) error {
	return ConflictError{Message: msg}
}

// Common errors
var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserNotFound         = errors.New("user not found")
	ErrUserExists           = NewConflictError("user already exists")
	ErrInvalidEmail         = NewValidationError("invalid email format")
	ErrInvalidPassword      = NewValidationError("password must be at least 8 characters long and contain at least one uppercase letter, one number, and one special character")
	ErrMissingRequired      = NewValidationError("all required fields must be provided")
	ErrInvalidCaptcha       = errors.New("invalid captcha")
	ErrTooManyLoginAttempts = errors.New("too many login attempts, please try again later")
	ErrAccountLocked        = errors.New("account is locked due to too many failed attempts")
	ErrLoanNotFound         = errors.New("loan not found")
)
