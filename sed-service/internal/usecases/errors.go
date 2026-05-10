package usecases

import "errors"

var (
	ErrNotFound      = errors.New("не найдено")
	ErrForbidden     = errors.New("запрещено")
	ErrConflict      = errors.New("конфликт")
	ErrValidation    = errors.New("некорректные данные")
	ErrWrongState    = errors.New("недопустимый статус документа")
	ErrWarehouse     = errors.New("ошибка интеграции со складом")
)
