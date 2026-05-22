package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/sed-service/internal/clients"
	"github.com/industrial-sed/sed-service/internal/middleware"
	"github.com/industrial-sed/sed-service/internal/usecases"
)

// RagInternalHandler internal API для RAG-корпуса.
type RagInternalHandler struct {
	App       *usecases.App
	AuthUsers *clients.AuthUsersClient
}

type setRagContentReq struct {
	RagContent json.RawMessage `json:"rag_content" binding:"required"`
}

type updateFixtureReq struct {
	Title      string          `json:"title" binding:"required"`
	Payload    json.RawMessage `json:"payload"`
	RagContent json.RawMessage `json:"rag_content" binding:"required"`
}

// GetCorpus GET /api/v1/internal/rag/corpus
func (h *RagInternalHandler) GetCorpus(c *gin.Context) {
	tenant := c.GetHeader(middleware.HeaderServiceTenantID)
	if tenant == "" {
		tenant = c.Query("tenant")
	}
	if tenant == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "нужен заголовок X-Tenant-Id или query tenant"})
		return
	}
	if h.AuthUsers == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auth users client not configured"})
		return
	}
	resp, err := h.App.BuildRagCorpus(c.Request.Context(), tenant, h.AuthUsers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateDocumentFixture PUT /api/v1/internal/rag/documents/:id/fixture
func (h *RagInternalHandler) UpdateDocumentFixture(c *gin.Context) {
	tenant := c.GetHeader(middleware.HeaderServiceTenantID)
	if tenant == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "нужен заголовок X-Tenant-Id"})
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req updateFixtureReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.App.UpdateDocumentFixture(c.Request.Context(), tenant, id, req.Title, req.Payload, req.RagContent); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// SetDocumentRagContent PUT /api/v1/internal/rag/documents/:id/content
func (h *RagInternalHandler) SetDocumentRagContent(c *gin.Context) {
	tenant := c.GetHeader(middleware.HeaderServiceTenantID)
	if tenant == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "нужен заголовок X-Tenant-Id"})
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req setRagContentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.App.SetDocumentRagContent(c.Request.Context(), tenant, id, req.RagContent); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
