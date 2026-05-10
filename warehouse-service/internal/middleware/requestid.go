package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/warehouse-service/internal/logger"
)

const HeaderRequestID = "X-Request-ID"

// RequestID добавляет request_id.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Writer.Header().Set(HeaderRequestID, rid)
		ctx := logger.WithRequestID(c.Request.Context(), rid)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
