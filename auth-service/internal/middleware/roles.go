package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/auth-service/internal/httpx"
)

// RequireRealmRoles требует хотя бы одну из realm-ролей Keycloak.
func RequireRealmRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil {
			httpx.ErrorJSON(c, http.StatusUnauthorized, "нет claims", http.StatusUnauthorized)
			c.Abort()
			return
		}
		for _, need := range roles {
			if cl.HasRole(need) {
				c.Next()
				return
			}
		}
		httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
		c.Abort()
	}
}
