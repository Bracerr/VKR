package server

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/industrial-sed/traceability-service/internal/config"
	"github.com/industrial-sed/traceability-service/internal/handlers"
	"github.com/industrial-sed/traceability-service/internal/jwtverify"
	"github.com/industrial-sed/traceability-service/internal/middleware"
)

type Deps struct {
	Log    *slog.Logger
	Parser *jwtverify.Parser
	H      *handlers.HTTP
	Cfg    *config.Config
	DB     *pgxpool.Pool
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID, X-Service-Secret")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func NewRouter(d Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(middleware.RequestID())
	r.Use(middleware.SlogAccessLog(d.Log))
	r.Use(middleware.PerIPRateLimit(d.Cfg.RateLimitPerMinute))

	r.GET("/health", handlers.Health)
	r.GET("/ready", handlers.Ready(&handlers.HealthDeps{DB: d.DB}))

	internal := r.Group("/api/v1/internal")
	internal.Use(middleware.ServiceSecretAuth(d.Cfg))
	{
		internal.POST("/events", d.H.PostInternalEvents)
	}

	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth(d.Parser))
	v1.Use(middleware.RequireViewTrace())
	{
		v1.GET("/trace/search", d.H.Search)
		v1.GET("/trace/graph", d.H.Graph)
	}

	return r
}

