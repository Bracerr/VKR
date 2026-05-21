package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/platform/outbox"
)

// OutboxAdmin replay/list failed outbox (service secret).
type OutboxAdmin struct {
	Store *outbox.Store
}

func (h *OutboxAdmin) ListFailed(c *gin.Context) {
	rows, err := h.Store.ListFailed(c.Request.Context(), 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

func (h *OutboxAdmin) Retry(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id"})
		return
	}
	if err := h.Store.ResetToPending(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
