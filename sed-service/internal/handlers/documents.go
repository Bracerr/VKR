package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/sed-service/internal/middleware"
)

type createDocumentReq struct {
	TypeID  uuid.UUID       `json:"type_id" binding:"required"`
	Title   string          `json:"title" binding:"required"`
	Payload json.RawMessage `json:"payload"`
}

type patchDocumentReq struct {
	Title   *string         `json:"title"`
	Payload json.RawMessage `json:"payload"`
}

type commentReq struct {
	Comment string `json:"comment"`
}

// ListDocuments GET /documents
func (h *HTTP) ListDocuments(c *gin.Context) {
	cl := middleware.Claims(c)
	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}
	var author *string
	if s := c.Query("author_sub"); s != "" {
		author = &s
	}
	list, err := h.App.ListDocuments(c.Request.Context(), cl.TenantID, status, author)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetDocument GET /documents/:id
func (h *HTTP) GetDocument(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	d, err := h.App.GetDocument(c.Request.Context(), cl.TenantID, id)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

// CreateDocument POST /documents
func (h *HTTP) CreateDocument(c *gin.Context) {
	var req createDocumentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	d, err := h.App.CreateDocument(c.Request.Context(), cl.TenantID, cl.Sub, req.TypeID, req.Title, req.Payload)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, d)
}

// PatchDocument PATCH /documents/:id
func (h *HTTP) PatchDocument(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req patchDocumentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.UpdateDocument(c.Request.Context(), cl.TenantID, cl.Sub, id, req.Title, req.Payload); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// SubmitDocument POST /documents/:id/submit
func (h *HTTP) SubmitDocument(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.SubmitDocument(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ApproveDocument POST /documents/:id/approve
func (h *HTTP) ApproveDocument(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req commentReq
	_ = c.ShouldBindJSON(&req)
	cl := middleware.Claims(c)
	if err := h.App.ApproveDocument(c.Request.Context(), cl.TenantID, cl.Sub, cl.RealmRoles, id, req.Comment); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RejectDocument POST /documents/:id/reject
func (h *HTTP) RejectDocument(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req commentReq
	_ = c.ShouldBindJSON(&req)
	cl := middleware.Claims(c)
	if err := h.App.RejectDocument(c.Request.Context(), cl.TenantID, cl.Sub, cl.RealmRoles, id, req.Comment); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// SignDocument POST /documents/:id/sign
func (h *HTTP) SignDocument(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.SignDocument(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// CancelDocument POST /documents/:id/cancel
func (h *HTTP) CancelDocument(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.CancelDocument(c.Request.Context(), cl.TenantID, cl.Sub, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DocumentHistory GET /documents/:id/history
func (h *HTTP) DocumentHistory(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if _, err := h.App.GetDocument(c.Request.Context(), cl.TenantID, id); err != nil {
		writeUsecaseError(c, err)
		return
	}
	list, err := h.App.ListDocumentHistory(c.Request.Context(), id)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// ListTasks GET /tasks
func (h *HTTP) ListTasks(c *gin.Context) {
	cl := middleware.Claims(c)
	list, err := h.App.ListTasks(c.Request.Context(), cl.TenantID, cl.Sub, cl.RealmRoles)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}
