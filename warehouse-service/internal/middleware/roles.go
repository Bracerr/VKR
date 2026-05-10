package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/warehouse-service/internal/httpx"
	"github.com/industrial-sed/warehouse-service/internal/models"
)

// RequireAnyRole одна из ролей.
func RequireAnyRole(roles ...string) gin.HandlerFunc {
	set := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		set[r] = struct{}{}
	}
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil {
			httpx.ErrorJSON(c, http.StatusUnauthorized, "требуется авторизация", http.StatusUnauthorized)
			c.Abort()
			return
		}
		for _, r := range cl.RealmRoles {
			if _, ok := set[r]; ok {
				c.Next()
				return
			}
		}
		httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
		c.Abort()
	}
}

// RequireWarehouseAdmin только админ склада.
func RequireWarehouseAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanAdminCatalog(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireOperate storekeeper или admin.
func RequireOperate() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanOperate(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireView любой складской доступ.
func RequireView() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanView(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
