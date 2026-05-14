package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/sales-service/internal/middleware"
	"github.com/industrial-sed/sales-service/internal/usecases"
)

type HTTP struct {
	App *usecases.App
}

// --- Customers ---

func (h *HTTP) ListCustomers(c *gin.Context) {
	cl := middleware.Claims(c)
	out, err := h.App.ListCustomers(c.Request.Context(), cl.TenantID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

type customerReq struct {
	Code     string          `json:"code" binding:"required"`
	Name     string          `json:"name" binding:"required"`
	Contacts json.RawMessage `json:"contacts"`
	Active   bool            `json:"active"`
}

func (h *HTTP) CreateCustomer(c *gin.Context) {
	var req customerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	out, err := h.App.CreateCustomer(c.Request.Context(), cl.TenantID, cl.Sub, req.Code, req.Name, req.Contacts, req.Active)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *HTTP) UpdateCustomer(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req customerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.UpdateCustomer(c.Request.Context(), cl.TenantID, cl.Sub, id, req.Code, req.Name, req.Contacts, req.Active); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) DeleteCustomer(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.DeleteCustomer(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Sales Orders ---

func (h *HTTP) ListSO(c *gin.Context) {
	cl := middleware.Claims(c)
	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}
	out, err := h.App.ListSO(c.Request.Context(), cl.TenantID, status)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *HTTP) GetSO(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	so, lines, err := h.App.GetSODetail(c.Request.Context(), cl.TenantID, id)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"so": so, "lines": lines})
}

type createSOReq struct {
	CustomerID         uuid.UUID  `json:"customer_id" binding:"required"`
	ShipFromWarehouseID *uuid.UUID `json:"ship_from_warehouse_id"`
	ShipFromBinID       *uuid.UUID `json:"ship_from_bin_id"`
	Note               *string    `json:"note"`
}

func (h *HTTP) CreateSO(c *gin.Context) {
	var req createSOReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	so, err := h.App.CreateSO(c.Request.Context(), cl.TenantID, cl.Sub, req.CustomerID, req.ShipFromWarehouseID, req.ShipFromBinID, req.Note)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, so)
}

type soLineReq struct {
	LineNo    int       `json:"line_no" binding:"required,min=1"`
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Qty       string    `json:"qty" binding:"required"`
	UOM       string    `json:"uom"`
	Note      *string   `json:"note"`
}

func (h *HTTP) AddSOLine(c *gin.Context) {
	soID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req soLineReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	ln, err := h.App.AddSOLine(c.Request.Context(), cl.TenantID, cl.Sub, soID, req.LineNo, req.ProductID, req.Qty, req.UOM, req.Note)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, ln)
}

type submitReq struct {
	SedDocumentTypeID uuid.UUID `json:"sed_document_type_id" binding:"required"`
	Title             string    `json:"title" binding:"required"`
}

func (h *HTTP) SubmitSO(c *gin.Context) {
	soID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req submitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.SubmitSO(c.Request.Context(), cl.TenantID, cl.Sub, bearerFromRequest(c), soID, req.SedDocumentTypeID, req.Title); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) ReleaseSO(c *gin.Context) {
	soID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.ReleaseSO(c.Request.Context(), cl.TenantID, cl.Sub, soID); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) CancelSO(c *gin.Context) {
	soID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.CancelSO(c.Request.Context(), cl.TenantID, cl.Sub, soID); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) ReserveSO(c *gin.Context) {
	soID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	ids, err := h.App.ReserveSO(c.Request.Context(), cl.TenantID, cl.Sub, soID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"reservation_ids": ids})
}

func (h *HTTP) ShipSO(c *gin.Context) {
	soID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	docID, err := h.App.ShipSO(c.Request.Context(), cl.TenantID, cl.Sub, soID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"warehouse_document_id": docID.String(), "posted_at": time.Now().UTC()})
}

// --- internal callback ---

type sedSignedReq struct {
	Event      string `json:"event" binding:"required"`
	TenantCode string `json:"tenant_code" binding:"required"`
	DocumentID string `json:"document_id" binding:"required"`
	TypeCode   string `json:"document_type_code"`
}

func (h *HTTP) PostSedEvents(c *gin.Context) {
	var req sedSignedReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Event != "DOCUMENT_SIGNED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported event"})
		return
	}
	docID, err := uuid.Parse(req.DocumentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document_id"})
		return
	}
	if err := h.App.HandleSedDocumentSigned(c.Request.Context(), req.TenantCode, docID); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

