package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/industrial-sed/traceability-service/internal/httpx"
)

// PerIPRateLimit лимит запросов на IP в минуту.
func PerIPRateLimit(perMinute int) gin.HandlerFunc {
	if perMinute <= 0 {
		perMinute = 120
	}
	rate := limiter.Rate{Period: time.Minute, Limit: int64(perMinute)}
	store := memory.NewStore()
	instance := limiter.New(store, rate)
	return mgin.NewMiddleware(instance,
		mgin.WithLimitReachedHandler(func(c *gin.Context) {
			httpx.ErrorJSON(c, http.StatusTooManyRequests, "слишком много запросов", http.StatusTooManyRequests)
		}),
	)
}

