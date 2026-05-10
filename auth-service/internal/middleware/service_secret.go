package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/auth-service/internal/httpx"
)

const HeaderServiceSecret = "X-Service-Secret"

// ServiceSecret проверяет заголовок для internal API.
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
