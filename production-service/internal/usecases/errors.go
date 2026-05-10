package usecases

import "errors"

var (
	ErrNotFound    = errors.New("не найдено")
	ErrForbidden   = errors.New("запрещено")
	ErrValidation  = errors.New("ошибка валидации")
	ErrWrongState  = errors.New("неверное состояние")
	ErrWarehouse   = errors.New("ошибка склада")
	ErrConflict    = errors.New("конфликт")
)
