package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/auth-service/internal/httpx"
	"github.com/industrial-sed/auth-service/internal/jwtverify"
)

const (
	// CookieAccessToken httpOnly cookie с access JWT.
	CookieAccessToken = "access_token"
	// CookieRefreshToken httpOnly cookie с refresh.
	CookieRefreshToken = "refresh_token"
	// CookieIDToken cookie с id_token (для logout / end_session).
	CookieIDToken = "id_token"
)

// JWTAuth проверяет Bearer или cookie access_token.
func JWTAuth(parser *jwtverify.Parser) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := extractBearer(c.GetHeader("Authorization"))
		if raw == "" {
			raw, _ = c.Cookie(CookieAccessToken)
		}
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
		c.Set(CtxClaimsKey, claims)
		c.Next()
	}
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
