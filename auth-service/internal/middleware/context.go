package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/auth-service/internal/models"
)

const (
	// CtxClaimsKey ключ в gin.Context для JWT-claims.
	CtxClaimsKey = "jwt_claims"
)

// Claims из контекста Gin (после JWTAuth).
func Claims(c *gin.Context) *models.Claims {
	v, ok := c.Get(CtxClaimsKey)
	if !ok {
		return nil
	}
	cl, _ := v.(*models.Claims)
	return cl
}
