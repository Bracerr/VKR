package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type sedSignedReq struct {
	Event       string `json:"event" binding:"required"`
	TenantCode  string `json:"tenant_code" binding:"required"`
	DocumentID  string `json:"document_id" binding:"required"`
	TypeCode    string `json:"document_type_code"`
}

// PostSedEvents POST /internal/sed-events (X-Service-Secret).
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
