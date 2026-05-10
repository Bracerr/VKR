package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/procurement-service/internal/middleware"
	"github.com/industrial-sed/procurement-service/internal/usecases"
)

type HTTP struct {
	App *usecases.App
}

// --- Suppliers ---

func (h *HTTP) ListSuppliers(c *gin.Context) {
	cl := middleware.Claims(c)
	out, err := h.App.ListSuppliers(c.Request.Context(), cl.TenantID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

type supplierReq struct {
	Code     string          `json:"code" binding:"required"`
	Name     string          `json:"name" binding:"required"`
	INN      *string         `json:"inn"`
	KPP      *string         `json:"kpp"`
	Contacts json.RawMessage `json:"contacts"`
	Active   bool            `json:"active"`
}

func (h *HTTP) CreateSupplier(c *gin.Context) {
	var req supplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	sp, err := h.App.CreateSupplier(c.Request.Context(), cl.TenantID, cl.Sub, req.Code, req.Name, req.INN, req.KPP, req.Contacts, req.Active)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, sp)
}

func (h *HTTP) UpdateSupplier(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req supplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.UpdateSupplier(c.Request.Context(), cl.TenantID, cl.Sub, id, req.Code, req.Name, req.INN, req.KPP, req.Contacts, req.Active); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) DeleteSupplier(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.DeleteSupplier(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- PR ---

func (h *HTTP) ListPR(c *gin.Context) {
	cl := middleware.Claims(c)
	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}
	out, err := h.App.ListPR(c.Request.Context(), cl.TenantID, status)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *HTTP) GetPR(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	pr, lines, err := h.App.GetPRDetail(c.Request.Context(), cl.TenantID, id)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"pr": pr, "lines": lines})
}

type createPRReq struct {
	NeededBy *time.Time `json:"needed_by"`
	Note     *string    `json:"note"`
}

func (h *HTTP) CreatePR(c *gin.Context) {
	var req createPRReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	pr, err := h.App.CreatePR(c.Request.Context(), cl.TenantID, cl.Sub, req.NeededBy, req.Note)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, pr)
}

type prLineReq struct {
	LineNo            int        `json:"line_no" binding:"required,min=1"`
	ProductID         uuid.UUID  `json:"product_id" binding:"required"`
	Qty               string     `json:"qty" binding:"required"`
	UOM               string     `json:"uom"`
	TargetWarehouseID *uuid.UUID `json:"target_warehouse_id"`
	TargetBinID       *uuid.UUID `json:"target_bin_id"`
	Note              *string    `json:"note"`
}

func (h *HTTP) AddPRLine(c *gin.Context) {
	prID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req prLineReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	ln, err := h.App.AddPRLine(c.Request.Context(), cl.TenantID, cl.Sub, prID, req.LineNo, req.ProductID, req.Qty, req.UOM, req.TargetWarehouseID, req.TargetBinID, req.Note)
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

func (h *HTTP) SubmitPR(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req submitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.SubmitPR(c.Request.Context(), cl.TenantID, cl.Sub, bearerFromRequest(c), id, req.SedDocumentTypeID, req.Title); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) CancelPR(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.CancelPR(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- PO ---

func (h *HTTP) ListPO(c *gin.Context) {
	cl := middleware.Claims(c)
	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}
	out, err := h.App.ListPO(c.Request.Context(), cl.TenantID, status)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *HTTP) GetPO(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	po, lines, err := h.App.GetPODetail(c.Request.Context(), cl.TenantID, id)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"po": po, "lines": lines})
}

type createPOReq struct {
	SupplierID uuid.UUID  `json:"supplier_id" binding:"required"`
	Currency   string     `json:"currency"`
	ExpectedAt *time.Time `json:"expected_at"`
}

func (h *HTTP) CreatePO(c *gin.Context) {
	var req createPOReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	po, err := h.App.CreatePO(c.Request.Context(), cl.TenantID, cl.Sub, req.SupplierID, req.Currency, req.ExpectedAt, nil)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, po)
}

type createPOFromPRReq struct {
	SupplierID uuid.UUID `json:"supplier_id" binding:"required"`
}

func (h *HTTP) CreatePOFromPR(c *gin.Context) {
	prID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req createPOFromPRReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	po, err := h.App.CreatePOFromPR(c.Request.Context(), cl.TenantID, cl.Sub, prID, req.SupplierID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, po)
}

type poLineReq struct {
	LineNo            int        `json:"line_no" binding:"required,min=1"`
	ProductID         uuid.UUID  `json:"product_id" binding:"required"`
	QtyOrdered        string     `json:"qty_ordered" binding:"required"`
	Price             string     `json:"price"`
	VATRate           string     `json:"vat_rate"`
	TargetWarehouseID *uuid.UUID `json:"target_warehouse_id"`
	TargetBinID       *uuid.UUID `json:"target_bin_id"`
}

func (h *HTTP) AddPOLine(c *gin.Context) {
	poID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req poLineReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	ln, err := h.App.AddPOLine(c.Request.Context(), cl.TenantID, cl.Sub, poID, req.LineNo, req.ProductID, req.QtyOrdered, req.Price, req.VATRate, req.TargetWarehouseID, req.TargetBinID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, ln)
}

func (h *HTTP) SubmitPO(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req submitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.SubmitPO(c.Request.Context(), cl.TenantID, cl.Sub, bearerFromRequest(c), id, req.SedDocumentTypeID, req.Title); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) ReleasePO(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.ReleasePO(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) CancelPO(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.CancelPO(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type receiveReq struct {
	WarehouseID uuid.UUID `json:"warehouse_id" binding:"required"`
	BinID       uuid.UUID `json:"bin_id" binding:"required"`
}

func (h *HTTP) ReceivePO(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req receiveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	docID, err := h.App.ReceivePO(c.Request.Context(), cl.TenantID, cl.Sub, id, req.WarehouseID, req.BinID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"warehouse_document_id": docID.String()})
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

