package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/traceability-service/internal/httpx"
	"github.com/industrial-sed/traceability-service/internal/usecases"
)

func writeUsecaseError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecases.ErrNotFound):
		httpx.ErrorJSON(c, http.StatusNotFound, err.Error(), http.StatusNotFound)
	case errors.Is(err, usecases.ErrForbidden):
		httpx.ErrorJSON(c, http.StatusForbidden, err.Error(), http.StatusForbidden)
	case errors.Is(err, usecases.ErrValidation):
		httpx.ErrorJSON(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
	case errors.Is(err, usecases.ErrConflict):
		httpx.ErrorJSON(c, http.StatusConflict, err.Error(), http.StatusConflict)
	default:
		httpx.ErrorJSON(c, http.StatusInternalServerError, err.Error(), http.StatusInternalServerError)
	}
}

