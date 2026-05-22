package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/sed-service/internal/middleware"
	"github.com/industrial-sed/sed-service/internal/usecases"
)

// RagInternalHandler internal API для выгрузки корпуса в RAG.
type RagInternalHandler struct {
	App            *usecases.App
	PublicBaseURL  string
}

type setRagContentReq struct {
	RagContent json.RawMessage `json:"rag_content" binding:"required"`
}

type updateFixtureReq struct {
	Title      string          `json:"title" binding:"required"`
	Payload    json.RawMessage `json:"payload"`
	RagContent json.RawMessage `json:"rag_content"`
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
	resp, err := h.App.BuildRagExport(c.Request.Context(), tenant, h.PublicBaseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DownloadFile GET /api/v1/internal/rag/documents/:id/files/:file_id
func (h *RagInternalHandler) DownloadFile(c *gin.Context) {
	tenant := c.GetHeader(middleware.HeaderServiceTenantID)
	if tenant == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "нужен заголовок X-Tenant-Id"})
		return
	}
	docID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	fileID, ok := parseUUIDParam(c, "file_id")
	if !ok {
		return
	}
	meta, rc, err := h.App.OpenDocumentFileInternal(c.Request.Context(), tenant, docID, fileID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	defer rc.Close()
	if meta.ContentType != nil && *meta.ContentType != "" {
		c.Header("Content-Type", *meta.ContentType)
	}
	c.Header("Content-Disposition", `attachment; filename="`+meta.OriginalName+`"`)
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, rc)
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
