package server

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/industrial-sed/production-service/internal/config"
	"github.com/industrial-sed/production-service/internal/handlers"
	"github.com/industrial-sed/production-service/internal/jwtverify"
	"github.com/industrial-sed/production-service/internal/middleware"
)

// Deps зависимости HTTP.
type Deps struct {
	Log    *slog.Logger
	Parser *jwtverify.Parser
	App    *handlers.HTTP
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

// NewRouter маршруты.
func NewRouter(d Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(middleware.RequestID())
	r.Use(middleware.SlogAccessLog(d.Log))
	r.Use(middleware.PerIPRateLimit(d.Cfg.RateLimitPerMinute))

	h := d.App

	r.GET("/health", handlers.Health)
	r.GET("/ready", handlers.Ready(&handlers.HealthDeps{DB: d.DB}))

	internal := r.Group("/api/v1/internal")
	internal.Use(middleware.ServiceSecretAuth(d.Cfg))
	{
		internal.POST("/sed-events", h.PostSedEvents)
	}

	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth(d.Parser))
	{
		view := v1.Group("")
		view.Use(middleware.RequireViewPROD())
		{
			view.GET("/workcenters", h.ListWorkcenters)
			view.GET("/scrap-reasons", h.ListScrapReasons)
			view.GET("/boms", h.ListBOMs)
			view.GET("/boms/:id", h.GetBOM)
			view.GET("/routings", h.ListRoutings)
			view.GET("/routings/:id", h.GetRouting)
			view.GET("/orders", h.ListOrders)
			view.GET("/orders/:id", h.GetOrder)
			view.GET("/shift-tasks", h.ListShiftTasks)
		}

		tech := v1.Group("")
		tech.Use(middleware.RequireTechnologist())
		{
			tech.POST("/workcenters", h.CreateWorkcenter)
			tech.PUT("/workcenters/:id", h.UpdateWorkcenter)
			tech.DELETE("/workcenters/:id", h.DeleteWorkcenter)
			tech.POST("/scrap-reasons", h.CreateScrapReason)
			tech.POST("/boms", h.CreateBOM)
			tech.PATCH("/boms/:id", h.PatchBOM)
			tech.POST("/boms/:id/lines", h.AddBOMLine)
			tech.DELETE("/boms/:id/lines/:line_id", h.DeleteBOMLine)
			tech.POST("/boms/:id/submit", h.SubmitBOM)
			tech.POST("/boms/:id/archive", h.ArchiveBOM)
			tech.POST("/routings", h.CreateRouting)
			tech.POST("/routings/:id/operations", h.AddRoutingOperation)
			tech.POST("/routings/:id/submit", h.SubmitRouting)
		}

		plan := v1.Group("")
		plan.Use(middleware.RequirePlanner())
		{
			plan.POST("/orders", h.CreateOrder)
			plan.POST("/orders/:id/release", h.ReleaseOrder)
			plan.POST("/orders/:id/cancel", h.CancelOrder)
			plan.POST("/orders/:id/complete", h.CompleteOrder)
			plan.POST("/shift-tasks", h.CreateShiftTask)
			plan.DELETE("/shift-tasks/:id", h.DeleteShiftTask)
		}

		wrk := v1.Group("")
		wrk.Use(middleware.RequireWorkerOrMaster())
		{
			wrk.GET("/me/shift-tasks", h.MeShiftTasks)
			wrk.POST("/orders/:id/operations/:op_id/start", h.StartOperation)
			wrk.POST("/orders/:id/operations/:op_id/report", h.ReportOperation)
			wrk.POST("/orders/:id/operations/:op_id/finish", h.FinishOperation)
		}
	}

	return r
}
