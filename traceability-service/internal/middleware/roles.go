package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/traceability-service/internal/httpx"
	"github.com/industrial-sed/traceability-service/internal/models"
)

func RequireViewTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanViewTrace(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

