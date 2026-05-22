package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/sed-service/internal/httpx"
)

const (
	HeaderServiceSecret   = "X-Service-Secret"
	HeaderServiceTenantID = "X-Tenant-Id"
)

// ServiceSecret защищает internal API (RAG corpus и т.п.).
func ServiceSecret(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret == "" || c.GetHeader(HeaderServiceSecret) != secret {
			httpx.ErrorJSON(c, http.StatusUnauthorized, "неверные учётные данные", http.StatusUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}
