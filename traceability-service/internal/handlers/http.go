package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/traceability-service/internal/middleware"
	"github.com/industrial-sed/traceability-service/internal/usecases"
)

type HTTP struct {
	App *usecases.App
}

type ingestReq struct {
	EventType      string          `json:"event_type" binding:"required"`
	TenantCode     string          `json:"tenant_code" binding:"required"`
	IdempotencyKey *string         `json:"idempotency_key"`
	Payload        json.RawMessage `json:"payload" binding:"required"`
}

func (h *HTTP) PostInternalEvents(c *gin.Context) {
	var req ingestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.App.Ingest(c.Request.Context(), &usecases.IngestEvent{
		EventType:      req.EventType,
		TenantCode:     req.TenantCode,
		IdempotencyKey: req.IdempotencyKey,
		Payload:        req.Payload,
	}); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) Search(c *gin.Context) {
	cl := middleware.Claims(c)
	serialNo := c.Query("serial_no")
	batchID := c.Query("batch_id")
	productID := c.Query("product_id")
	from := c.Query("from")
	to := c.Query("to")

	out, err := h.App.Search(c.Request.Context(), cl.TenantID, serialNo, batchID, productID, from, to)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *HTTP) Graph(c *gin.Context) {
	cl := middleware.Claims(c)
	anchorType := c.Query("anchor_type")
	anchorID := c.Query("anchor_id")
	from := c.Query("from")
	to := c.Query("to")
	depth := c.Query("depth")

	out, err := h.App.Graph(c.Request.Context(), cl.TenantID, anchorType, anchorID, from, to, depth)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

