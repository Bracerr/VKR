package usecases

import "errors"

var (
	ErrNotFound   = errors.New("не найдено")
	ErrForbidden  = errors.New("запрещено")
	ErrValidation = errors.New("ошибка валидации")
	ErrConflict   = errors.New("конфликт")
)

