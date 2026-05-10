package server

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/industrial-sed/procurement-service/internal/config"
	"github.com/industrial-sed/procurement-service/internal/handlers"
	"github.com/industrial-sed/procurement-service/internal/jwtverify"
	"github.com/industrial-sed/procurement-service/internal/middleware"
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
		view.Use(middleware.RequireViewProc())
		{
			view.GET("/suppliers", d.H.ListSuppliers)
			view.GET("/purchase-requests", d.H.ListPR)
			view.GET("/purchase-requests/:id", d.H.GetPR)
			view.GET("/purchase-orders", d.H.ListPO)
			view.GET("/purchase-orders/:id", d.H.GetPO)
		}

		buy := v1.Group("")
		buy.Use(middleware.RequireBuyer())
		{
			buy.POST("/suppliers", d.H.CreateSupplier)
			buy.PUT("/suppliers/:id", d.H.UpdateSupplier)
			buy.DELETE("/suppliers/:id", d.H.DeleteSupplier)

			buy.POST("/purchase-requests", d.H.CreatePR)
			buy.POST("/purchase-requests/:id/lines", d.H.AddPRLine)
			buy.POST("/purchase-requests/:id/submit", d.H.SubmitPR)
			buy.POST("/purchase-requests/:id/cancel", d.H.CancelPR)

			buy.POST("/purchase-orders", d.H.CreatePO)
			buy.POST("/purchase-orders/from-pr/:id", d.H.CreatePOFromPR)
			buy.POST("/purchase-orders/:id/lines", d.H.AddPOLine)
			buy.POST("/purchase-orders/:id/submit", d.H.SubmitPO)
			buy.POST("/purchase-orders/:id/release", d.H.ReleasePO)
			buy.POST("/purchase-orders/:id/cancel", d.H.CancelPO)
			buy.POST("/purchase-orders/:id/receive", d.H.ReceivePO)
		}
	}

	return r
}

