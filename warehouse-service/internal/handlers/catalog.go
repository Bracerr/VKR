package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/warehouse-service/internal/models"
	"github.com/industrial-sed/warehouse-service/internal/usecases"
)

// Catalog HTTP справочников.
type Catalog struct {
	UC *usecases.UC
}

type productReq struct {
	SKU              string  `json:"sku" binding:"required"`
	Name             string  `json:"name" binding:"required"`
	Unit             string  `json:"unit"`
	TrackingMode     string  `json:"tracking_mode"`
	HasExpiration    bool    `json:"has_expiration"`
	ValuationMethod  string  `json:"valuation_method"`
	DefaultCurrency  string  `json:"default_currency"`
	StandardCost     *string `json:"standard_cost"`
}

func (h *Catalog) ListProducts(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	list, err := h.UC.ListProducts(c.Request.Context(), tn)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Catalog) GetProduct(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	p, err := h.UC.GetProduct(c.Request.Context(), tn, id)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	if p == nil {
		RespondError(c, http.StatusNotFound, "не найдено", http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Catalog) CreateProduct(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	var req productReq
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	p := &models.Product{
		ID:              uuid.New(),
		TenantCode:      tn,
		SKU:             req.SKU,
		Name:            req.Name,
		Unit:            req.Unit,
		TrackingMode:    req.TrackingMode,
		HasExpiration:   req.HasExpiration,
		ValuationMethod: req.ValuationMethod,
		DefaultCurrency: req.DefaultCurrency,
	}
	if p.Unit == "" {
		p.Unit = "pcs"
	}
	if p.TrackingMode == "" {
		p.TrackingMode = models.TrackingNone
	}
	if p.ValuationMethod == "" {
		p.ValuationMethod = models.ValAverage
	}
	if p.DefaultCurrency == "" {
		p.DefaultCurrency = h.UC.DefaultCurrency
	}
	if req.StandardCost != nil {
		d, err := parseDecimal(*req.StandardCost)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "standard_cost", http.StatusBadRequest)
			return
		}
		p.StandardCost = &d
	}
	if err := h.UC.CreateProduct(c.Request.Context(), p); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *Catalog) UpdateProduct(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req productReq
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	p := &models.Product{
		ID:              id,
		TenantCode:      tn,
		SKU:             req.SKU,
		Name:            req.Name,
		Unit:            req.Unit,
		TrackingMode:    req.TrackingMode,
		HasExpiration:   req.HasExpiration,
		ValuationMethod: req.ValuationMethod,
		DefaultCurrency: req.DefaultCurrency,
	}
	if req.StandardCost != nil {
		d, err := parseDecimal(*req.StandardCost)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "standard_cost", http.StatusBadRequest)
			return
		}
		p.StandardCost = &d
	}
	if err := h.UC.UpdateProduct(c.Request.Context(), p); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Catalog) DeleteProduct(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.UC.DeleteProduct(c.Request.Context(), tn, id); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Warehouses ---

func (h *Catalog) ListWarehouses(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	list, err := h.UC.ListWarehouses(c.Request.Context(), tn)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

type whReq struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func (h *Catalog) CreateWarehouse(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	var req whReq
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	w := &models.Warehouse{ID: uuid.New(), TenantCode: tn, Code: req.Code, Name: req.Name}
	if err := h.UC.CreateWarehouse(c.Request.Context(), w); err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, w)
}

func (h *Catalog) UpdateWarehouse(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	w := &models.Warehouse{ID: id, TenantCode: tn, Name: req.Name}
	if err := h.UC.UpdateWarehouse(c.Request.Context(), w); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *Catalog) DeleteWarehouse(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.UC.DeleteWarehouse(c.Request.Context(), tn, id); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Bins ---

func (h *Catalog) ListBins(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	whID, ok := parseUUID(c, "warehouse_id")
	if !ok {
		return
	}
	list, err := h.UC.ListBins(c.Request.Context(), tn, whID)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

type binReq struct {
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name"`
	BinType     string  `json:"bin_type"`
	ParentBinID *string `json:"parent_bin_id"`
	CapacityQty *string `json:"capacity_qty"`
}

func (h *Catalog) CreateBin(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	whID, ok := parseUUID(c, "warehouse_id")
	if !ok {
		return
	}
	var req binReq
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	b := &models.Bin{ID: uuid.New(), TenantCode: tn, WarehouseID: whID, Code: req.Code, Name: req.Name, BinType: req.BinType}
	if b.BinType == "" {
		b.BinType = "STORAGE"
	}
	if req.ParentBinID != nil && *req.ParentBinID != "" {
		pid, err := uuid.Parse(*req.ParentBinID)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "parent_bin_id", http.StatusBadRequest)
			return
		}
		b.ParentBinID = &pid
	}
	if req.CapacityQty != nil {
		d, err := parseDecimal(*req.CapacityQty)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "capacity_qty", http.StatusBadRequest)
			return
		}
		b.CapacityQty = &d
	}
	if err := h.UC.CreateBin(c.Request.Context(), b); err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, b)
}

func (h *Catalog) UpdateBin(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req binReq
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	b := &models.Bin{ID: id, TenantCode: tn, Code: req.Code, Name: req.Name, BinType: req.BinType}
	if req.ParentBinID != nil && *req.ParentBinID != "" {
		pid, err := uuid.Parse(*req.ParentBinID)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "parent_bin_id", http.StatusBadRequest)
			return
		}
		b.ParentBinID = &pid
	}
	if req.CapacityQty != nil {
		d, err := parseDecimal(*req.CapacityQty)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "capacity_qty", http.StatusBadRequest)
			return
		}
		b.CapacityQty = &d
	}
	if err := h.UC.UpdateBin(c.Request.Context(), b); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *Catalog) DeleteBin(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.UC.DeleteBin(c.Request.Context(), tn, id); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Prices ---

func (h *Catalog) ListPrices(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	pid, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	list, err := h.UC.ListPrices(c.Request.Context(), tn, pid)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

type priceReq struct {
	PriceType string `json:"price_type" binding:"required"`
	Currency  string `json:"currency" binding:"required"`
	Price     string `json:"price" binding:"required"`
	ValidFrom string `json:"valid_from" binding:"required"`
	ValidTo   string `json:"valid_to"`
}

func (h *Catalog) CreatePrice(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	pid, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req priceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	price, err := parseDecimal(req.Price)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "price", http.StatusBadRequest)
		return
	}
	vf, err := time.Parse("2006-01-02", req.ValidFrom)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "valid_from", http.StatusBadRequest)
		return
	}
	var vt *time.Time
	if req.ValidTo != "" {
		t, err := time.Parse("2006-01-02", req.ValidTo)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "valid_to", http.StatusBadRequest)
			return
		}
		vt = &t
	}
	pr := &models.ProductPrice{
		ID: uuid.New(), TenantCode: tn, ProductID: pid,
		PriceType: req.PriceType, Currency: req.Currency, Price: price,
		ValidFrom: vf, ValidTo: vt,
	}
	if err := h.UC.CreatePrice(c.Request.Context(), pr); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, pr)
}

func (h *Catalog) DeletePrice(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.UC.DeletePrice(c.Request.Context(), tn, id); err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Catalog) ListSerials(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	var pid *uuid.UUID
	if s := c.Query("product_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "product_id", http.StatusBadRequest)
			return
		}
		pid = &id
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
	list, err := h.UC.ListSerials(c.Request.Context(), tn, pid, st, wh)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Catalog) SerialHistory(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	list, err := h.UC.SerialMovementHistory(c.Request.Context(), id)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Catalog) GetBatch(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	b, err := h.UC.GetBatch(c.Request.Context(), tn, id)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, b)
}
