package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/sed-service/internal/httpx"
	"github.com/industrial-sed/sed-service/internal/usecases"
)

func writeUsecaseError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecases.ErrNotFound):
		httpx.ErrorJSON(c, http.StatusNotFound, err.Error(), http.StatusNotFound)
	case errors.Is(err, usecases.ErrForbidden):
		httpx.ErrorJSON(c, http.StatusForbidden, err.Error(), http.StatusForbidden)
	case errors.Is(err, usecases.ErrValidation):
		httpx.ErrorJSON(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
	case errors.Is(err, usecases.ErrWrongState):
		httpx.ErrorJSON(c, http.StatusConflict, err.Error(), http.StatusConflict)
	case errors.Is(err, usecases.ErrConflict):
		httpx.ErrorJSON(c, http.StatusConflict, err.Error(), http.StatusConflict)
	case errors.Is(err, usecases.ErrWarehouse):
		httpx.ErrorJSON(c, http.StatusBadGateway, err.Error(), http.StatusBadGateway)
	default:
		httpx.ErrorJSON(c, http.StatusInternalServerError, err.Error(), http.StatusInternalServerError)
	}
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	s := c.Param(name)
	id, err := uuid.Parse(s)
	if err != nil {
		httpx.ErrorJSON(c, http.StatusBadRequest, "некорректный UUID", http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}
