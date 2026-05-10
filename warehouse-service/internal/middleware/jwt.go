package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/warehouse-service/internal/httpx"
	"github.com/industrial-sed/warehouse-service/internal/jwtverify"
	"github.com/industrial-sed/warehouse-service/internal/models"
)

const (
	CtxClaimsKey           = "claims"
	HeaderServiceSecret    = "X-Service-Secret"
	HeaderServiceTenantID  = "X-Tenant-Id"
)

// JWTAuth Bearer JWT или сервисный вызов (X-Service-Secret + X-Tenant-Id) от sed-service.
func JWTAuth(parser *jwtverify.Parser, serviceSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sec := c.GetHeader(HeaderServiceSecret)
		tid := c.GetHeader(HeaderServiceTenantID)
		if serviceSecret != "" && sec == serviceSecret && tid != "" {
			c.Set(CtxClaimsKey, &models.Claims{
				Sub:        "service:sed",
				Username:   "sed-service",
				TenantID:   tid,
				RealmRoles: []string{models.RoleWarehouseAdmin},
			})
			c.Next()
			return
		}

		raw := extractBearer(c.GetHeader("Authorization"))
		if raw == "" {
			httpx.ErrorJSON(c, http.StatusUnauthorized, "требуется авторизация", http.StatusUnauthorized)
			c.Abort()
			return
		}
		claims, err := parser.ParseAccessToken(c.Request.Context(), raw)
		if err != nil {
			httpx.ErrorJSON(c, http.StatusUnauthorized, "невалидный токен", http.StatusUnauthorized)
			c.Abort()
			return
		}
		if claims.TenantID == "" {
			httpx.ErrorJSON(c, http.StatusForbidden, "в токене отсутствует tenant_id", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Set(CtxClaimsKey, claims)
		c.Next()
	}
}

// Claims из контекста Gin.
func Claims(c *gin.Context) *models.Claims {
	v, ok := c.Get(CtxClaimsKey)
	if !ok {
		return nil
	}
	cl, _ := v.(*models.Claims)
	return cl
}

func extractBearer(h string) string {
	if h == "" {
		return ""
	}
	const p = "Bearer "
	if len(h) > len(p) && strings.EqualFold(h[:len(p)], p) {
		return strings.TrimSpace(h[len(p):])
	}
	return ""
}
