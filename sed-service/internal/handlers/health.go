package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/industrial-sed/sed-service/internal/httpx"
)

// HealthDeps зависимости /ready.
type HealthDeps struct {
	DB *pgxpool.Pool
}

// Health liveness.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready readiness (БД).
func Ready(deps *HealthDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if deps == nil || deps.DB == nil {
			c.JSON(http.StatusOK, gin.H{"status": "ready"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := deps.DB.Ping(ctx); err != nil {
			httpx.ErrorJSON(c, http.StatusServiceUnavailable, "база недоступна", http.StatusServiceUnavailable)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}
