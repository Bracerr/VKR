package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/sed-service/internal/httpx"
	"github.com/industrial-sed/sed-service/internal/models"
)

// RequireSedAdmin только sed_admin.
func RequireSedAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanAdminSED(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAuthor автор документов.
func RequireAuthor() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanAuthor(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireApprover согласующий.
func RequireApprover() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanApprove(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireViewSED чтение.
func RequireViewSED() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanViewSED(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
