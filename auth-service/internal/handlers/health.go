package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthDeps зависимости health/ready.
type HealthDeps struct {
	DB            *pgxpool.Pool
	KeycloakURL   string
	KeycloakRealm string
}

// Health liveness.
// @Summary Liveness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready readiness: PostgreSQL + Keycloak realm.
// @Summary Readiness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} httpx.ErrorBody
// @Router /ready [get]
func Ready(deps *HealthDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		out := gin.H{"status": "ok", "postgres": "ok", "keycloak": "ok"}
		if deps.DB != nil {
			if err := deps.DB.Ping(ctx); err != nil {
				RespondError(c, http.StatusServiceUnavailable, "postgres: "+err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
		base := strings.TrimRight(deps.KeycloakURL, "/")
		url := base + "/realms/" + deps.KeycloakRealm
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode >= 400 {
			if resp != nil {
				_ = resp.Body.Close()
			}
			RespondError(c, http.StatusServiceUnavailable, "keycloak недоступен", http.StatusServiceUnavailable)
			return
		}
		_ = resp.Body.Close()
		c.JSON(http.StatusOK, out)
	}
}
