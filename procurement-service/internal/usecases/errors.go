package usecases

import "errors"

var (
	ErrNotFound   = errors.New("не найдено")
	ErrForbidden  = errors.New("запрещено")
	ErrValidation = errors.New("ошибка валидации")
	ErrWrongState = errors.New("неверное состояние")
	ErrConflict   = errors.New("конфликт")
	ErrWarehouse  = errors.New("ошибка склада")
)

