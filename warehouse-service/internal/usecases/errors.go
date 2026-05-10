package usecases

import "errors"

var (
	ErrNotFound         = errors.New("не найдено")
	ErrForbidden        = errors.New("запрещено")
	ErrConflict         = errors.New("конфликт")
	ErrValidation       = errors.New("некорректные данные")
	ErrInsufficient     = errors.New("недостаточно товара")
	ErrCapacityExceeded = errors.New("превышена вместимость ячейки")
)
