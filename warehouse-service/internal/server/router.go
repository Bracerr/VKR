package server

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/industrial-sed/warehouse-service/internal/config"
	"github.com/industrial-sed/warehouse-service/internal/handlers"
	"github.com/industrial-sed/warehouse-service/internal/jwtverify"
	"github.com/industrial-sed/warehouse-service/internal/middleware"
	"github.com/industrial-sed/warehouse-service/internal/usecases"
)

// Deps зависимости HTTP.
type Deps struct {
	Log    *slog.Logger
	Parser *jwtverify.Parser
	UC     *usecases.UC
	Cfg    *config.Config
	DB     *pgxpool.Pool
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Request-ID, X-Service-Secret, X-Tenant-Id")
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

	cat := &handlers.Catalog{UC: d.UC}
	ops := &handlers.Ops{UC: d.UC}
	inv := &handlers.Inv{UC: d.UC}
	res := &handlers.Res{UC: d.UC}
	rep := &handlers.Rep{UC: d.UC}
	imp := &handlers.Imp{UC: d.UC, Cfg: d.Cfg}

	r.GET("/health", handlers.Health)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/ready", handlers.Ready(&handlers.HealthDeps{
		DB:            d.DB,
		KeycloakURL:   d.Cfg.KeycloakURL,
		KeycloakRealm: d.Cfg.KeycloakRealm,
	}))

	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth(d.Parser, d.Cfg.ServiceSecret))

	// Чтение (viewer+)
	v1GET := v1.Group("")
	v1GET.Use(middleware.RequireView())
	{
		v1GET.GET("/products", cat.ListProducts)
		v1GET.GET("/products/:id", cat.GetProduct)
		v1GET.GET("/warehouses", cat.ListWarehouses)
		v1GET.GET("/warehouses/:warehouse_id/bins", cat.ListBins)
		v1GET.GET("/batches/:id", cat.GetBatch)
		v1GET.GET("/products/:id/prices", cat.ListPrices)
		v1GET.GET("/serials", cat.ListSerials)
		v1GET.GET("/serials/:id/history", cat.SerialHistory)
		v1GET.GET("/balances", rep.Balances)
		v1GET.GET("/movements", rep.Movements)
		v1GET.GET("/reports/stock-on-date", rep.StockOnDate)
		v1GET.GET("/reports/turnover", rep.Turnover)
		v1GET.GET("/reports/abc", rep.ABC)
		v1GET.GET("/reports/expiring", rep.Expiring)
		v1GET.GET("/reports/price-on-date", rep.PriceOnDate)
		v1GET.GET("/reports/average-cost", rep.AvgCostOnDate)
		v1GET.GET("/reservations", res.List)
		v1GET.GET("/reservations/:id", res.Get)
		v1GET.GET("/inventory/:id", inv.Get)
		v1GET.GET("/import/jobs/:id", imp.GetImportJob)
		v1GET.GET("/export/movements.csv", imp.ExportMovementsCSV)
	}

	// Админ справочников
	adm := v1.Group("")
	adm.Use(middleware.RequireWarehouseAdmin())
	{
		adm.POST("/products", cat.CreateProduct)
		adm.PUT("/products/:id", cat.UpdateProduct)
		adm.DELETE("/products/:id", cat.DeleteProduct)
		adm.POST("/warehouses", cat.CreateWarehouse)
		adm.PUT("/warehouses/:id", cat.UpdateWarehouse)
		adm.DELETE("/warehouses/:id", cat.DeleteWarehouse)
		adm.POST("/warehouses/:warehouse_id/bins", cat.CreateBin)
		adm.PUT("/bins/:id", cat.UpdateBin)
		adm.DELETE("/bins/:id", cat.DeleteBin)
		adm.POST("/products/:id/prices", cat.CreatePrice)
		adm.DELETE("/prices/:id", cat.DeletePrice)
		adm.POST("/import/products", imp.ImportProductsCSV)
	}

	// Операции кладовщика
	op := v1.Group("")
	op.Use(middleware.RequireOperate())
	{
		op.POST("/operations/receipt", ops.Receipt)
		op.POST("/operations/issue", ops.Issue)
		op.POST("/operations/transfer", ops.Transfer)
		op.POST("/operations/relocate", ops.Relocate)
		op.POST("/inventory", inv.Start)
		op.PATCH("/inventory/lines/:line_id", inv.SetCounted)
		op.POST("/inventory/:id/post", inv.Post)
		op.POST("/reservations", res.Create)
		op.POST("/reservations/:id/release", res.Release)
		op.POST("/reservations/:id/consume", res.Consume)
	}

	return r
}
