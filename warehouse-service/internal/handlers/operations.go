package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/warehouse-service/internal/usecases"
)

// Ops складские операции.
type Ops struct {
	UC *usecases.UC
}

type receiptLineReq struct {
	ProductID      string   `json:"product_id" binding:"required"`
	Qty            string   `json:"qty" binding:"required"`
	Series         *string  `json:"series"`
	ManufacturedAt *string  `json:"manufactured_at"`
	ExpiresAt      *string  `json:"expires_at"`
	UnitCost       *string  `json:"unit_cost"`
	Currency       *string  `json:"currency"`
	SerialNumbers  []string `json:"serial_numbers"`
}

type receiptReq struct {
	WarehouseID string           `json:"warehouse_id" binding:"required"`
	BinID       string           `json:"bin_id" binding:"required"`
	Lines       []receiptLineReq `json:"lines" binding:"required,min=1"`
}

func (h *Ops) Receipt(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	var req receiptReq
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
	var lines []usecases.ReceiptLineIn
	for _, ln := range req.Lines {
		pid, err := uuid.Parse(ln.ProductID)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "product_id", http.StatusBadRequest)
			return
		}
		qty, err := parseDecimal(ln.Qty)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "qty", http.StatusBadRequest)
			return
		}
		rl := usecases.ReceiptLineIn{ProductID: pid, Qty: qty, Series: ln.Series, SerialNumbers: ln.SerialNumbers}
		if ln.UnitCost != nil {
			u, err := parseDecimal(*ln.UnitCost)
			if err != nil {
				RespondError(c, http.StatusBadRequest, "unit_cost", http.StatusBadRequest)
				return
			}
			rl.UnitCost = &u
		}
		rl.Currency = ln.Currency
		if ln.ManufacturedAt != nil && *ln.ManufacturedAt != "" {
			t, err := time.Parse("2006-01-02", *ln.ManufacturedAt)
			if err != nil {
				RespondError(c, http.StatusBadRequest, "manufactured_at", http.StatusBadRequest)
				return
			}
			rl.ManufacturedAt = &t
		}
		if ln.ExpiresAt != nil && *ln.ExpiresAt != "" {
			t, err := time.Parse("2006-01-02", *ln.ExpiresAt)
			if err != nil {
				RespondError(c, http.StatusBadRequest, "expires_at", http.StatusBadRequest)
				return
			}
			rl.ExpiresAt = &t
		}
		lines = append(lines, rl)
	}
	docID, err := h.UC.Receipt(c.Request.Context(), tn, userName(c), whID, binID, lines)
	if err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"document_id": docID})
}

type issueLineReq struct {
	ProductID     string   `json:"product_id" binding:"required"`
	Qty           string   `json:"qty"`
	BatchID       *string  `json:"batch_id"`
	SerialNumbers []string `json:"serial_numbers"`
}

type issueReq struct {
	WarehouseID string         `json:"warehouse_id" binding:"required"`
	BinID       string         `json:"bin_id" binding:"required"`
	Lines       []issueLineReq `json:"lines" binding:"required,min=1"`
}

func (h *Ops) Issue(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	var req issueReq
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
	var lines []usecases.IssueIn
	for _, ln := range req.Lines {
		pid, err := uuid.Parse(ln.ProductID)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "product_id", http.StatusBadRequest)
			return
		}
		issue := usecases.IssueIn{ProductID: pid, SerialNumbers: ln.SerialNumbers}
		if len(ln.SerialNumbers) == 0 {
			if ln.Qty == "" {
				RespondError(c, http.StatusBadRequest, "qty обязателен", http.StatusBadRequest)
				return
			}
			q, err := parseDecimal(ln.Qty)
			if err != nil {
				RespondError(c, http.StatusBadRequest, "qty", http.StatusBadRequest)
				return
			}
			issue.Qty = q
		}
		if ln.BatchID != nil && *ln.BatchID != "" {
			bid, err := uuid.Parse(*ln.BatchID)
			if err != nil {
				RespondError(c, http.StatusBadRequest, "batch_id", http.StatusBadRequest)
				return
			}
			issue.BatchID = &bid
		}
		lines = append(lines, issue)
	}
	docID, err := h.UC.Issue(c.Request.Context(), tn, userName(c), whID, binID, lines)
	if err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"document_id": docID})
}

type transferLineReq struct {
	ProductID     string   `json:"product_id" binding:"required"`
	Qty           string   `json:"qty"`
	BatchID       *string  `json:"batch_id"`
	SerialNumbers []string `json:"serial_numbers"`
}

type transferReq struct {
	WarehouseFromID string            `json:"warehouse_from_id" binding:"required"`
	BinFromID       string            `json:"bin_from_id" binding:"required"`
	WarehouseToID   string            `json:"warehouse_to_id" binding:"required"`
	BinToID         string            `json:"bin_to_id" binding:"required"`
	Lines           []transferLineReq `json:"lines" binding:"required,min=1"`
}

func (h *Ops) Transfer(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	var req transferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	wf, err := uuid.Parse(req.WarehouseFromID)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "warehouse_from_id", http.StatusBadRequest)
		return
	}
	bf, err := uuid.Parse(req.BinFromID)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "bin_from_id", http.StatusBadRequest)
		return
	}
	wt, err := uuid.Parse(req.WarehouseToID)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "warehouse_to_id", http.StatusBadRequest)
		return
	}
	bt, err := uuid.Parse(req.BinToID)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "bin_to_id", http.StatusBadRequest)
		return
	}
	lines, ok := parseTransferLines(c, req.Lines)
	if !ok {
		return
	}
	docID, err := h.UC.Transfer(c.Request.Context(), tn, userName(c), wf, bf, wt, bt, lines)
	if err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"document_id": docID})
}

type relocateReq struct {
	WarehouseID string            `json:"warehouse_id" binding:"required"`
	BinFromID   string            `json:"bin_from_id" binding:"required"`
	BinToID     string            `json:"bin_to_id" binding:"required"`
	Lines       []transferLineReq `json:"lines" binding:"required,min=1"`
}

func (h *Ops) Relocate(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	var req relocateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	whID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "warehouse_id", http.StatusBadRequest)
		return
	}
	bf, err := uuid.Parse(req.BinFromID)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "bin_from_id", http.StatusBadRequest)
		return
	}
	bt, err := uuid.Parse(req.BinToID)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "bin_to_id", http.StatusBadRequest)
		return
	}
	lines, ok := parseTransferLines(c, req.Lines)
	if !ok {
		return
	}
	docID, err := h.UC.Relocate(c.Request.Context(), tn, userName(c), whID, bf, bt, lines)
	if err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"document_id": docID})
}

func parseTransferLines(c *gin.Context, in []transferLineReq) ([]usecases.TransferLine, bool) {
	var out []usecases.TransferLine
	for _, ln := range in {
		pid, err := uuid.Parse(ln.ProductID)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "product_id", http.StatusBadRequest)
			return nil, false
		}
		tl := usecases.TransferLine{ProductID: pid, SerialNumbers: ln.SerialNumbers}
		if len(ln.SerialNumbers) == 0 {
			q, err := parseDecimal(ln.Qty)
			if err != nil {
				RespondError(c, http.StatusBadRequest, "qty", http.StatusBadRequest)
				return nil, false
			}
			tl.Qty = q
		}
		if ln.BatchID != nil && *ln.BatchID != "" {
			bid, err := uuid.Parse(*ln.BatchID)
			if err != nil {
				RespondError(c, http.StatusBadRequest, "batch_id", http.StatusBadRequest)
				return nil, false
			}
			tl.BatchID = &bid
		}
		out = append(out, tl)
	}
	return out, true
}
