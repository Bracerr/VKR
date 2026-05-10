package usecases

import "errors"

// Доменные ошибки usecase-слоя.
var (
	ErrNotFound    = errors.New("not found")
	ErrForbidden   = errors.New("forbidden")
	ErrConflict    = errors.New("conflict")
	ErrValidation  = errors.New("validation")
	ErrUnauthorized = errors.New("unauthorized")
)
