package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/production-service/internal/httpx"
	"github.com/industrial-sed/production-service/internal/models"
)

func RequireTechnologist() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanTechnologist(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequirePlanner() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanPlan(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireMaster() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanMaster(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireWorkerOrMaster() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanWorker(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireViewPROD() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil || !models.CanViewPROD(cl) {
			httpx.ErrorJSON(c, http.StatusForbidden, "недостаточно прав", http.StatusForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
