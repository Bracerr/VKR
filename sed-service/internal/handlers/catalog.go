package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/sed-service/internal/middleware"
)

type createDocTypeReq struct {
	Code               string     `json:"code" binding:"required"`
	Name               string     `json:"name" binding:"required"`
	WarehouseAction    string     `json:"warehouse_action"`
	DefaultWorkflowID  *uuid.UUID `json:"default_workflow_id"`
}

type updateDocTypeReq struct {
	Name               string     `json:"name" binding:"required"`
	WarehouseAction    string     `json:"warehouse_action" binding:"required"`
	DefaultWorkflowID  *uuid.UUID `json:"default_workflow_id"`
}

type createWorkflowReq struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type updateWorkflowReq struct {
	Name string `json:"name" binding:"required"`
}

type addStepReq struct {
	OrderNo        int     `json:"order_no" binding:"required,min=1"`
	ParallelGroup  *int    `json:"parallel_group"`
	Name           string  `json:"name" binding:"required"`
	RequiredRole   *string `json:"required_role"`
	RequiredUserSub *string `json:"required_user_sub"`
}

// ListDocumentTypes GET /document-types
func (h *HTTP) ListDocumentTypes(c *gin.Context) {
	cl := middleware.Claims(c)
	list, err := h.App.ListDocumentTypes(c.Request.Context(), cl.TenantID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// CreateDocumentType POST /document-types
func (h *HTTP) CreateDocumentType(c *gin.Context) {
	var req createDocTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	dt, err := h.App.CreateDocumentType(c.Request.Context(), cl.TenantID, req.Code, req.Name, req.WarehouseAction, req.DefaultWorkflowID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dt)
}

// GetDocumentType GET /document-types/:id
func (h *HTTP) GetDocumentType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	dt, err := h.App.GetDocumentType(c.Request.Context(), cl.TenantID, id)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, dt)
}

// UpdateDocumentType PUT /document-types/:id
func (h *HTTP) UpdateDocumentType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req updateDocTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.UpdateDocumentType(c.Request.Context(), cl.TenantID, id, req.Name, req.WarehouseAction, req.DefaultWorkflowID); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteDocumentType DELETE /document-types/:id
func (h *HTTP) DeleteDocumentType(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.DeleteDocumentType(c.Request.Context(), cl.TenantID, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListWorkflows GET /workflows
func (h *HTTP) ListWorkflows(c *gin.Context) {
	cl := middleware.Claims(c)
	list, err := h.App.ListWorkflows(c.Request.Context(), cl.TenantID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// CreateWorkflow POST /workflows
func (h *HTTP) CreateWorkflow(c *gin.Context) {
	var req createWorkflowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	w, err := h.App.CreateWorkflow(c.Request.Context(), cl.TenantID, req.Code, req.Name)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, w)
}

// UpdateWorkflow PUT /workflows/:id
func (h *HTTP) UpdateWorkflow(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req updateWorkflowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.UpdateWorkflow(c.Request.Context(), cl.TenantID, id, req.Name); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteWorkflow DELETE /workflows/:id
func (h *HTTP) DeleteWorkflow(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.DeleteWorkflow(c.Request.Context(), cl.TenantID, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListWorkflowSteps GET /workflows/:id/steps
func (h *HTTP) ListWorkflowSteps(c *gin.Context) {
	wid, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	list, err := h.App.ListWorkflowSteps(c.Request.Context(), cl.TenantID, wid)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// AddWorkflowStep POST /workflows/:id/steps
func (h *HTTP) AddWorkflowStep(c *gin.Context) {
	wid, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req addStepReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	st, err := h.App.AddWorkflowStep(c.Request.Context(), cl.TenantID, wid, req.OrderNo, req.ParallelGroup, req.Name, req.RequiredRole, req.RequiredUserSub)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, st)
}

// DeleteWorkflowStep DELETE /workflow-steps/:id
func (h *HTTP) DeleteWorkflowStep(c *gin.Context) {
	sid, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.DeleteWorkflowStep(c.Request.Context(), cl.TenantID, sid); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
