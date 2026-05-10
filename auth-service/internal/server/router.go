package server

import (
	"log/slog"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/industrial-sed/auth-service/internal/config"
	"github.com/industrial-sed/auth-service/internal/handlers"
	"github.com/industrial-sed/auth-service/internal/jwtverify"
	"github.com/industrial-sed/auth-service/internal/keycloak"
	"github.com/industrial-sed/auth-service/internal/middleware"
)

// Deps зависимости HTTP-слоя.
type Deps struct {
	Config   *config.Config
	Log      *slog.Logger
	DB       *pgxpool.Pool
	Parser   *jwtverify.Parser
	TenantUC *handlers.TenantHandler
	UserUC   *handlers.UserHandler
	Auth     *handlers.AuthHandler
	Test     *handlers.TestHandler
	OTel     bool
}

// NewRouter собирает Gin.
func NewRouter(d Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.SlogAccessLog(d.Log))
	if d.OTel {
		r.Use(otelgin.Middleware(d.Config.OTelServiceName))
	}

	corsCfg := cors.Config{
		AllowOrigins:     []string{d.Config.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", middleware.HeaderServiceSecret, middleware.HeaderTestSecret, middleware.HeaderRequestID},
		AllowCredentials: true,
	}
	r.Use(cors.New(corsCfg))

	r.GET("/health", handlers.Health)
	r.GET("/ready", handlers.Ready(&handlers.HealthDeps{
		DB:            d.DB,
		KeycloakURL:   d.Config.KeycloakURL,
		KeycloakRealm: d.Config.KeycloakRealm,
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.GET("/login", middleware.AuthLoginRateLimit(d.Config.AuthLoginRateLimit), d.Auth.Login)
	auth.GET("/callback", d.Auth.Callback)
	auth.POST("/refresh", d.Auth.Refresh)
	auth.POST("/logout", d.Auth.Logout)

	me := v1.Group("/auth")
	me.Use(middleware.JWTAuth(d.Parser))
	me.GET("/me", handlers.Me)

	ten := v1.Group("/tenants")
	ten.Use(middleware.JWTAuth(d.Parser), middleware.RequireRealmRoles(keycloak.RoleSuperAdmin))
	ten.POST("", d.TenantUC.Create)
	ten.GET("", d.TenantUC.List)
	ten.DELETE("/:code", d.TenantUC.Delete)
	ten.POST("/:code/ent-admin", d.TenantUC.BootstrapEntAdmin)

	users := v1.Group("/users")
	users.Use(middleware.JWTAuth(d.Parser), middleware.RequireRealmRoles(keycloak.RoleEntAdmin))
	users.POST("", d.UserUC.Create)
	users.GET("", d.UserUC.List)
	users.PUT("/:id/roles", d.UserUC.UpdateRoles)
	users.DELETE("/:id", d.UserUC.Delete)

	internal := v1.Group("/internal")
	internal.Use(middleware.ServiceSecret(d.Config.ServiceSecret))
	internal.GET("/userinfo", middleware.JWTAuth(d.Parser), handlers.UserInfo)

	if d.Config.EnableTestEndpoints {
		test := v1.Group("/internal/test")
		test.Use(middleware.TestSecret(d.Config.TestSecret))
		test.POST("/login", d.Test.Login)
		test.DELETE("/cleanup", d.Test.Cleanup)
	}

	return r
}
