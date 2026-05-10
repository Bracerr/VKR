package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/production-service/internal/middleware"
	"github.com/industrial-sed/production-service/internal/usecases"
)

// HTTP хендлеры production-service.
type HTTP struct {
	App *usecases.App
}

// --- Workcenters ---

func (h *HTTP) ListWorkcenters(c *gin.Context) {
	cl := middleware.Claims(c)
	list, err := h.App.ListWorkcenters(c.Request.Context(), cl.TenantID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

type wcReq struct {
	Code                    string `json:"code" binding:"required"`
	Name                    string `json:"name" binding:"required"`
	Active                  bool   `json:"active"`
	CapacityMinutesPerShift *int   `json:"capacity_minutes_per_shift"`
}

func (h *HTTP) CreateWorkcenter(c *gin.Context) {
	var req wcReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	w, err := h.App.CreateWorkcenter(c.Request.Context(), cl.TenantID, cl.Sub, req.Code, req.Name, req.Active, req.CapacityMinutesPerShift)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, w)
}

func (h *HTTP) UpdateWorkcenter(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req wcReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.UpdateWorkcenter(c.Request.Context(), cl.TenantID, cl.Sub, id, req.Code, req.Name, req.Active, req.CapacityMinutesPerShift); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) DeleteWorkcenter(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.DeleteWorkcenter(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Scrap ---

func (h *HTTP) ListScrapReasons(c *gin.Context) {
	cl := middleware.Claims(c)
	list, err := h.App.ListScrapReasons(c.Request.Context(), cl.TenantID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

type scrapReq struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func (h *HTTP) CreateScrapReason(c *gin.Context) {
	var req scrapReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	r, err := h.App.CreateScrapReason(c.Request.Context(), cl.TenantID, cl.Sub, req.Code, req.Name)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, r)
}

// --- BOM ---

func (h *HTTP) ListBOMs(c *gin.Context) {
	cl := middleware.Claims(c)
	var productID *uuid.UUID
	if s := c.Query("product_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product_id"})
			return
		}
		productID = &id
	}
	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}
	list, err := h.App.ListBOMs(c.Request.Context(), cl.TenantID, productID, status)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *HTTP) GetBOM(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	b, lines, err := h.App.GetBOMWithLines(c.Request.Context(), cl.TenantID, id)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"bom": b, "lines": lines})
}

type createBOMReq struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
}

func (h *HTTP) CreateBOM(c *gin.Context) {
	var req createBOMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	b, err := h.App.CreateBOM(c.Request.Context(), cl.TenantID, cl.Sub, req.ProductID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, b)
}

type bomDatesReq struct {
	ValidFrom *time.Time `json:"valid_from"`
	ValidTo   *time.Time `json:"valid_to"`
}

func (h *HTTP) PatchBOM(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req bomDatesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.UpdateBOMDates(c.Request.Context(), cl.TenantID, cl.Sub, id, req.ValidFrom, req.ValidTo); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type bomLineReq struct {
	LineNo             int       `json:"line_no" binding:"required,min=1"`
	ComponentProductID uuid.UUID `json:"component_product_id" binding:"required"`
	QtyPer             string    `json:"qty_per" binding:"required"`
	ScrapPct           string    `json:"scrap_pct"`
	OpNo               int       `json:"op_no" binding:"required,min=1"`
	AltGroup           *string   `json:"alt_group"`
}

func (h *HTTP) AddBOMLine(c *gin.Context) {
	bomID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req bomLineReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	scrap := req.ScrapPct
	if scrap == "" {
		scrap = "0"
	}
	cl := middleware.Claims(c)
	ln, err := h.App.AddBOMLine(c.Request.Context(), cl.TenantID, cl.Sub, bomID, req.LineNo, req.ComponentProductID, req.QtyPer, scrap, req.OpNo, req.AltGroup)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, ln)
}

func (h *HTTP) DeleteBOMLine(c *gin.Context) {
	bomID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	lineID, ok := parseUUIDParam(c, "line_id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.DeleteBOMLine(c.Request.Context(), cl.TenantID, cl.Sub, bomID, lineID); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type submitBOMReq struct {
	SedDocumentTypeID uuid.UUID `json:"sed_document_type_id" binding:"required"`
	Title             string    `json:"title" binding:"required"`
}

func (h *HTTP) SubmitBOM(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req submitBOMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	bearer := bearerFromRequest(c)
	if err := h.App.SubmitBOM(c.Request.Context(), cl.TenantID, cl.Sub, bearer, id, req.SedDocumentTypeID, req.Title); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) ArchiveBOM(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.ArchiveBOM(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Routings ---

func (h *HTTP) ListRoutings(c *gin.Context) {
	cl := middleware.Claims(c)
	var productID *uuid.UUID
	if s := c.Query("product_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product_id"})
			return
		}
		productID = &id
	}
	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}
	list, err := h.App.ListRoutings(c.Request.Context(), cl.TenantID, productID, status)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *HTTP) GetRouting(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	r, ops, err := h.App.GetRoutingWithOps(c.Request.Context(), cl.TenantID, id)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"routing": r, "operations": ops})
}

type createRoutingReq struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
}

func (h *HTTP) CreateRouting(c *gin.Context) {
	var req createRoutingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	r, err := h.App.CreateRouting(c.Request.Context(), cl.TenantID, cl.Sub, req.ProductID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, r)
}

type routingOpReq struct {
	OpNo          int       `json:"op_no" binding:"required,min=1"`
	WorkcenterID  uuid.UUID `json:"workcenter_id" binding:"required"`
	Name          string    `json:"name" binding:"required"`
	TimePerUnitMin *string  `json:"time_per_unit_min"`
	SetupTimeMin   *string  `json:"setup_time_min"`
	QCRequired    bool      `json:"qc_required"`
}

func (h *HTTP) AddRoutingOperation(c *gin.Context) {
	routingID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req routingOpReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	op, err := h.App.AddRoutingOperation(c.Request.Context(), cl.TenantID, cl.Sub, routingID, req.OpNo, req.WorkcenterID, req.Name, req.TimePerUnitMin, req.SetupTimeMin, req.QCRequired)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, op)
}

type submitRoutingReq struct {
	SedDocumentTypeID uuid.UUID `json:"sed_document_type_id" binding:"required"`
	Title             string    `json:"title" binding:"required"`
}

func (h *HTTP) SubmitRouting(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req submitRoutingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	bearer := bearerFromRequest(c)
	if err := h.App.SubmitRouting(c.Request.Context(), cl.TenantID, cl.Sub, bearer, id, req.SedDocumentTypeID, req.Title); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Orders ---

func (h *HTTP) ListOrders(c *gin.Context) {
	cl := middleware.Claims(c)
	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}
	list, err := h.App.ListProductionOrders(c.Request.Context(), cl.TenantID, status)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *HTTP) GetOrder(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	o, ops, err := h.App.GetProductionOrderDetail(c.Request.Context(), cl.TenantID, id)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"order": o, "operations": ops})
}

type createOrderReq struct {
	Code          string     `json:"code"`
	ProductID     uuid.UUID  `json:"product_id" binding:"required"`
	BomID         uuid.UUID  `json:"bom_id" binding:"required"`
	RoutingID     uuid.UUID  `json:"routing_id" binding:"required"`
	WarehouseID   uuid.UUID  `json:"warehouse_id" binding:"required"`
	DefaultBinID  uuid.UUID  `json:"default_bin_id" binding:"required"`
	QtyPlanned    string     `json:"qty_planned" binding:"required"`
	StartPlan     *time.Time `json:"start_plan"`
	FinishPlan    *time.Time `json:"finish_plan"`
}

func (h *HTTP) CreateOrder(c *gin.Context) {
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	o, err := h.App.CreateProductionOrder(c.Request.Context(), cl.TenantID, cl.Sub, req.Code, req.ProductID, req.BomID, req.RoutingID, req.WarehouseID, req.DefaultBinID, req.QtyPlanned, req.StartPlan, req.FinishPlan)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, o)
}

func (h *HTTP) ReleaseOrder(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.ReleaseProductionOrder(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) CancelOrder(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.CancelProductionOrder(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) CompleteOrder(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.CompleteProductionOrder(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Ops ---

func (h *HTTP) StartOperation(c *gin.Context) {
	orderID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	opID, ok := parseUUIDParam(c, "op_id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.StartOperation(c.Request.Context(), cl.TenantID, cl.Sub, orderID, opID); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type reportReq struct {
	QtyGood         string  `json:"qty_good" binding:"required"`
	QtyScrap        string  `json:"qty_scrap"`
	ScrapReasonCode *string `json:"scrap_reason_code"`
	Note            *string `json:"note"`
}

func (h *HTTP) ReportOperation(c *gin.Context) {
	orderID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	opID, ok := parseUUIDParam(c, "op_id")
	if !ok {
		return
	}
	var req reportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	scrap := req.QtyScrap
	if scrap == "" {
		scrap = "0"
	}
	cl := middleware.Claims(c)
	if err := h.App.ReportOperation(c.Request.Context(), cl.TenantID, cl.Sub, orderID, opID, req.QtyGood, scrap, req.ScrapReasonCode, req.Note); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTP) FinishOperation(c *gin.Context) {
	orderID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	opID, ok := parseUUIDParam(c, "op_id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.FinishOperation(c.Request.Context(), cl.TenantID, cl.Sub, orderID, opID); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Shift tasks ---

func (h *HTTP) ListShiftTasks(c *gin.Context) {
	cl := middleware.Claims(c)
	var date *time.Time
	if s := c.Query("date"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date"})
			return
		}
		date = &t
	}
	list, err := h.App.ListShiftTasks(c.Request.Context(), cl.TenantID, date)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *HTTP) MeShiftTasks(c *gin.Context) {
	cl := middleware.Claims(c)
	list, err := h.App.MeShiftTasks(c.Request.Context(), cl.TenantID, cl.Sub)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

type shiftTaskReq struct {
	OrderOperationID uuid.UUID  `json:"order_operation_id" binding:"required"`
	ShiftDate        time.Time  `json:"shift_date" binding:"required"`
	ShiftNo          int        `json:"shift_no" binding:"required,min=1,max=3"`
	AssigneeSub      *string    `json:"assignee_sub"`
	QtyPlanned       *string    `json:"qty_planned"`
}

func (h *HTTP) CreateShiftTask(c *gin.Context) {
	var req shiftTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	t, err := h.App.CreateShiftTask(c.Request.Context(), cl.TenantID, cl.Sub, req.OrderOperationID, req.ShiftDate, req.ShiftNo, req.AssigneeSub, req.QtyPlanned)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *HTTP) DeleteShiftTask(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.DeleteShiftTask(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
