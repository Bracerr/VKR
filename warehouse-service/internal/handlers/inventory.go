package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/warehouse-service/internal/usecases"
)

// Inv инвентаризация.
type Inv struct {
	UC *usecases.UC
}

type invStartReq struct {
	WarehouseID string  `json:"warehouse_id" binding:"required"`
	BinID       *string `json:"bin_id"`
}

func (h *Inv) Start(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	var req invStartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	whID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "warehouse_id", http.StatusBadRequest)
		return
	}
	var binID *uuid.UUID
	if req.BinID != nil && *req.BinID != "" {
		b, err := uuid.Parse(*req.BinID)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "bin_id", http.StatusBadRequest)
			return
		}
		binID = &b
	}
	id, err := h.UC.StartInventory(c.Request.Context(), tn, userName(c), whID, binID)
	if err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"document_id": id})
}

type invCountedReq struct {
	Counted string `json:"counted" binding:"required"`
}

func (h *Inv) SetCounted(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	lineID, ok := parseUUID(c, "line_id")
	if !ok {
		return
	}
	var req invCountedReq
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	cnt, err := parseDecimal(req.Counted)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "counted", http.StatusBadRequest)
		return
	}
	if err := h.UC.SetInventoryCounted(c.Request.Context(), tn, lineID, cnt); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Inv) Post(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	docID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.UC.PostInventory(c.Request.Context(), tn, userName(c), docID); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Inv) Get(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	docID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	d, err := h.UC.GetDocument(c.Request.Context(), tn, docID)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	if d == nil || d.DocType != "INVENTORY" {
		RespondError(c, http.StatusNotFound, "не найдено", http.StatusNotFound)
		return
	}
	lines, err := h.UC.ListInventoryLines(c.Request.Context(), tn, docID)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"document": d, "lines": lines})
}
