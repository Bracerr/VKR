package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/auth-service/internal/httpx"
)

const HeaderTestSecret = "X-Test-Secret"

// TestSecret защищает тестовые эндпоинты.
func TestSecret(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret == "" || c.GetHeader(HeaderTestSecret) != secret {
			httpx.ErrorJSON(c, http.StatusUnauthorized, "неверные учётные данные", http.StatusUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}
