package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/sales-service/internal/logger"
)

const HeaderRequestID = "X-Request-ID"

// RequestID propagates request id.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Writer.Header().Set(HeaderRequestID, rid)
		c.Request = c.Request.WithContext(logger.WithRequestID(c.Request.Context(), rid))
		c.Next()
	}
}

