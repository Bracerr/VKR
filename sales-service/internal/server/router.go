package server

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/industrial-sed/sales-service/internal/config"
	"github.com/industrial-sed/sales-service/internal/handlers"
	"github.com/industrial-sed/sales-service/internal/jwtverify"
	"github.com/industrial-sed/sales-service/internal/middleware"
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
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Request-ID, X-Service-Secret")
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
		internal.POST("/sed-events", d.H.PostSedEvents)
	}

	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth(d.Parser))
	{
		view := v1.Group("")
		view.Use(middleware.RequireViewSales())
		{
			view.GET("/customers", d.H.ListCustomers)
			view.GET("/sales-orders", d.H.ListSO)
			view.GET("/sales-orders/:id", d.H.GetSO)
		}

		mgr := v1.Group("")
		mgr.Use(middleware.RequireManager())
		{
			mgr.POST("/customers", d.H.CreateCustomer)
			mgr.PUT("/customers/:id", d.H.UpdateCustomer)
			mgr.DELETE("/customers/:id", d.H.DeleteCustomer)

			mgr.POST("/sales-orders", d.H.CreateSO)
			mgr.POST("/sales-orders/:id/lines", d.H.AddSOLine)
			mgr.POST("/sales-orders/:id/submit", d.H.SubmitSO)
			mgr.POST("/sales-orders/:id/release", d.H.ReleaseSO)
			mgr.POST("/sales-orders/:id/cancel", d.H.CancelSO)
			mgr.POST("/sales-orders/:id/reserve", d.H.ReserveSO)
			mgr.POST("/sales-orders/:id/ship", d.H.ShipSO)
		}
	}

	return r
}

