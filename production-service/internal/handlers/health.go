package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthDeps зависимости health.
type HealthDeps struct {
	DB *pgxpool.Pool
}

// Health liveness.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready readiness (БД).
func Ready(d *HealthDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d.DB == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "no_db"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := d.DB.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}
