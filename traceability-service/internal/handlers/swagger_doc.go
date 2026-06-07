package handlers

// Аннотации swag для маршрутов internal/server/router.go.

// Health godoc
// @Summary Liveness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func swaggerTraceHealth() {}

// Ready godoc
// @Summary Readiness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]interface{}
// @Router /ready [get]
func swaggerTraceReady() {}

// PostInternalEvents godoc
// @Summary События от других сервисов (internal)
// @Tags internal
// @Accept json
// @Param X-Service-Secret header string true "Секрет сервиса"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/internal/events [post]
func swaggerTracePostInternalEvents() {}

// Search godoc
// @Summary Поиск по прослеживаемости
// @Tags trace
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/trace/search [get]
func swaggerTraceSearch() {}

// Graph godoc
// @Summary Граф прослеживаемости
// @Tags trace
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/trace/graph [get]
func swaggerTraceGraph() {}
