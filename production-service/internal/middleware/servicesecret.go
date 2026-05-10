package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/production-service/internal/config"
	"github.com/industrial-sed/production-service/internal/httpx"
)

const HeaderServiceSecret = "X-Service-Secret"

// ServiceSecretAuth проверка секрета для internal callback (sed-service → production).
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
