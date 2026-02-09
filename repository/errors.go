package repository

import "errors"

var (
	ErrNotFound          = errors.New("record not found")
	ErrInvalidID         = errors.New("invalid ID")
	ErrDatabaseError     = errors.New("database error")
	ErrAlreadyExists     = errors.New("record already exists")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrValidationFailed  = errors.New("validation failed")
	ErrTransactionFailed = errors.New("transaction failed")
)
