package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/warehouse-service/internal/httpx"
	"github.com/industrial-sed/warehouse-service/internal/usecases"
)

// RespondError JSON-ошибка.
func RespondError(c *gin.Context, httpStatus int, message string, code int) {
	httpx.ErrorJSON(c, httpStatus, message, code)
}

// RespondUsecaseError маппинг доменных ошибок.
func RespondUsecaseError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, usecases.ErrNotFound), errors.Is(err, pgx.ErrNoRows):
		RespondError(c, http.StatusNotFound, "не найдено", http.StatusNotFound)
	case errors.Is(err, usecases.ErrForbidden):
		RespondError(c, http.StatusForbidden, "запрещено", http.StatusForbidden)
	case errors.Is(err, usecases.ErrConflict):
		RespondError(c, http.StatusConflict, "конфликт", http.StatusConflict)
	case errors.Is(err, usecases.ErrInsufficient):
		RespondError(c, http.StatusConflict, "недостаточно товара", http.StatusConflict)
	case errors.Is(err, usecases.ErrCapacityExceeded):
		RespondError(c, http.StatusUnprocessableEntity, "превышена вместимость ячейки", http.StatusUnprocessableEntity)
	case errors.Is(err, usecases.ErrValidation):
		RespondError(c, http.StatusUnprocessableEntity, err.Error(), http.StatusUnprocessableEntity)
	default:
		RespondError(c, http.StatusInternalServerError, "внутренняя ошибка", http.StatusInternalServerError)
	}
	return true
}
