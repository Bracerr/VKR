package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/sed-service/internal/middleware"
)

// ListDocumentFiles GET /documents/:id/files
func (h *HTTP) ListDocumentFiles(c *gin.Context) {
	docID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	list, err := h.App.ListDocumentFiles(c.Request.Context(), cl.TenantID, docID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// UploadDocumentFile POST /documents/:id/files
func (h *HTTP) UploadDocumentFile(c *gin.Context) {
	docID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "нужен multipart file «file»"})
		return
	}
	src, err := fh.Open()
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	defer src.Close()

	cl := middleware.Claims(c)
	ct := fh.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	meta, err := h.App.UploadFile(c.Request.Context(), cl.TenantID, cl.Sub, docID, fh.Filename, ct, fh.Size, src)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusCreated, meta)
}

// DownloadDocumentFile GET /documents/:id/files/:file_id
func (h *HTTP) DownloadDocumentFile(c *gin.Context) {
	docID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	fileID, ok := parseUUIDParam(c, "file_id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	meta, err := h.App.GetFileMeta(c.Request.Context(), cl.TenantID, docID, fileID)
	if err != nil {
		writeUsecaseError(c, err)
		return
	}
	rc, err := h.App.OpenFileStream(c.Request.Context(), meta.ObjectKey)
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

// DeleteDocumentFile DELETE /documents/:id/files/:file_id
func (h *HTTP) DeleteDocumentFile(c *gin.Context) {
	docID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	fileID, ok := parseUUIDParam(c, "file_id")
	if !ok {
		return
	}
	cl := middleware.Claims(c)
	if err := h.App.DeleteFile(c.Request.Context(), cl.TenantID, cl.Sub, docID, fileID); err != nil {
		writeUsecaseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
