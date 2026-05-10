package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/warehouse-service/internal/usecases"
)

// Res HTTP резервов.
type Res struct {
	UC *usecases.UC
}

type resCreateReq struct {
	WarehouseID string  `json:"warehouse_id" binding:"required"`
	BinID       string  `json:"bin_id" binding:"required"`
	ProductID   string  `json:"product_id" binding:"required"`
	BatchID     *string `json:"batch_id"`
	SerialNo    *string `json:"serial_no"`
	Qty         string  `json:"qty" binding:"required"`
	Reason      string  `json:"reason"`
	DocRef      string  `json:"doc_ref"`
	ExpiresAt   *string `json:"expires_at"`
}

func (h *Res) Create(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	var req resCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	whID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "warehouse_id", http.StatusBadRequest)
		return
	}
	binID, err := uuid.Parse(req.BinID)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "bin_id", http.StatusBadRequest)
		return
	}
	pid, err := uuid.Parse(req.ProductID)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "product_id", http.StatusBadRequest)
		return
	}
	qty, err := parseDecimal(req.Qty)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "qty", http.StatusBadRequest)
		return
	}
	in := usecases.ReservationIn{
		WarehouseID: whID, BinID: binID, ProductID: pid, SerialNo: req.SerialNo,
		Qty: qty, Reason: req.Reason, DocRef: req.DocRef,
	}
	if req.BatchID != nil && *req.BatchID != "" {
		bid, err := uuid.Parse(*req.BatchID)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "batch_id", http.StatusBadRequest)
			return
		}
		in.BatchID = &bid
	}
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "expires_at RFC3339", http.StatusBadRequest)
			return
		}
		in.ExpiresAt = &t
	}
	id, err := h.UC.CreateReservation(c.Request.Context(), tn, userName(c), in)
	if err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *Res) List(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	var st *string
	if s := c.Query("status"); s != "" {
		st = &s
	}
	var wh *uuid.UUID
	if s := c.Query("warehouse_id"); s != "" {
		id, err := uuid.Parse(s)
		if err == nil {
			wh = &id
		}
	}
	var pr *uuid.UUID
	if s := c.Query("product_id"); s != "" {
		id, err := uuid.Parse(s)
		if err == nil {
			pr = &id
		}
	}
	list, err := h.UC.ListReservations(c.Request.Context(), tn, st, wh, pr)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Res) Get(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	r, err := h.UC.GetReservation(c.Request.Context(), tn, id)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	if r == nil {
		RespondError(c, http.StatusNotFound, "не найдено", http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *Res) Release(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.UC.ReleaseReservation(c.Request.Context(), tn, id); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Res) Consume(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.UC.ConsumeReservation(c.Request.Context(), tn, userName(c), id); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}
