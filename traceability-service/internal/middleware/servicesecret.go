package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/traceability-service/internal/config"
	"github.com/industrial-sed/traceability-service/internal/httpx"
)

const HeaderServiceSecret = "X-Service-Secret"

// ServiceSecretAuth проверка секрета для internal ingest (warehouse/sales/proc/prod → traceability).
func ServiceSecretAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		sec := c.GetHeader(HeaderServiceSecret)
		if cfg.TraceIngestSecret == "" || sec != cfg.TraceIngestSecret {
			httpx.ErrorJSON(c, http.StatusUnauthorized, "неверный сервисный секрет", http.StatusUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

