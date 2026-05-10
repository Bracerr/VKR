package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/procurement-service/internal/httpx"
	"github.com/industrial-sed/procurement-service/internal/models"
)

func RequireBuyer() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanBuy(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireViewProc() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanViewProc(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

