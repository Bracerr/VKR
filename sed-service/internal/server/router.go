package server

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/industrial-sed/sed-service/internal/config"
	"github.com/industrial-sed/sed-service/internal/handlers"
	"github.com/industrial-sed/sed-service/internal/jwtverify"
	"github.com/industrial-sed/sed-service/internal/middleware"
	"github.com/industrial-sed/sed-service/internal/usecases"
)

// Deps зависимости HTTP.
type Deps struct {
	Log    *slog.Logger
	Parser *jwtverify.Parser
	App    *usecases.App
	Cfg    *config.Config
	DB     *pgxpool.Pool
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// NewRouter собирает маршруты.
func NewRouter(d Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(middleware.RequestID())
	r.Use(middleware.SlogAccessLog(d.Log))
	r.Use(middleware.PerIPRateLimit(d.Cfg.RateLimitPerMinute))

	h := &handlers.HTTP{App: d.App}

	r.GET("/health", handlers.Health)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/ready", handlers.Ready(&handlers.HealthDeps{DB: d.DB}))

	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth(d.Parser))

	admin := v1.Group("")
	admin.Use(middleware.RequireSedAdmin())
	{
		admin.GET("/document-types", h.ListDocumentTypes)
		admin.POST("/document-types", h.CreateDocumentType)
		admin.GET("/document-types/:id", h.GetDocumentType)
		admin.PUT("/document-types/:id", h.UpdateDocumentType)
		admin.DELETE("/document-types/:id", h.DeleteDocumentType)

		admin.GET("/workflows", h.ListWorkflows)
		admin.POST("/workflows", h.CreateWorkflow)
		admin.PUT("/workflows/:id", h.UpdateWorkflow)
		admin.DELETE("/workflows/:id", h.DeleteWorkflow)
		admin.GET("/workflows/:id/steps", h.ListWorkflowSteps)
		admin.POST("/workflows/:id/steps", h.AddWorkflowStep)
		admin.DELETE("/workflow-steps/:id", h.DeleteWorkflowStep)
	}

	view := v1.Group("")
	view.Use(middleware.RequireViewSED())
	{
		view.GET("/documents", h.ListDocuments)
		view.GET("/documents/:id", h.GetDocument)
		view.GET("/documents/:id/history", h.DocumentHistory)
		view.GET("/documents/:id/files", h.ListDocumentFiles)
		view.GET("/documents/:id/files/:file_id", h.DownloadDocumentFile)
	}

	authDoc := v1.Group("")
	authDoc.Use(middleware.RequireAuthor())
	{
		authDoc.POST("/documents", h.CreateDocument)
		authDoc.PATCH("/documents/:id", h.PatchDocument)
		authDoc.POST("/documents/:id/submit", h.SubmitDocument)
		authDoc.POST("/documents/:id/sign", h.SignDocument)
		authDoc.POST("/documents/:id/cancel", h.CancelDocument)
		authDoc.POST("/documents/:id/files", h.UploadDocumentFile)
		authDoc.DELETE("/documents/:id/files/:file_id", h.DeleteDocumentFile)
	}

	appr := v1.Group("")
	appr.Use(middleware.RequireApprover())
	{
		appr.GET("/tasks", h.ListTasks)
		appr.POST("/documents/:id/approve", h.ApproveDocument)
		appr.POST("/documents/:id/reject", h.RejectDocument)
	}

	return r
}
