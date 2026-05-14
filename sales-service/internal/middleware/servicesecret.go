package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/sales-service/internal/config"
	"github.com/industrial-sed/sales-service/internal/httpx"
)

const HeaderServiceSecret = "X-Service-Secret"

// ServiceSecretAuth проверка секрета для internal callback (sed-service → sales).
func ServiceSecretAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		sec := c.GetHeader(HeaderServiceSecret)
		if cfg.SedCallbackVerifySecret == "" || sec != cfg.SedCallbackVerifySecret {
			httpx.ErrorJSON(c, http.StatusUnauthorized, "неверный сервисный секрет", http.StatusUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

