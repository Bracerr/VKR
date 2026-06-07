package handlers

// Аннотации swag для маршрутов internal/server/router.go.

// Health godoc
// @Summary Liveness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func swaggerProdHealth() {}

// Ready godoc
// @Summary Readiness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]interface{}
// @Router /ready [get]
func swaggerProdReady() {}

// PostSedEvents godoc
// @Summary События от СЭД (internal)
// @Tags internal
// @Accept json
// @Param X-Service-Secret header string true "Секрет сервиса"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/internal/sed-events [post]
func swaggerProdPostSedEvents() {}

// ListWorkcenters godoc
// @Summary Список рабочих центров
// @Tags workcenters
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/workcenters [get]
func swaggerProdListWorkcenters() {}

// CreateWorkcenter godoc
// @Summary Создать рабочий центр
// @Tags workcenters
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/workcenters [post]
func swaggerProdCreateWorkcenter() {}

// UpdateWorkcenter godoc
// @Summary Обновить рабочий центр
// @Tags workcenters
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workcenters/{id} [put]
func swaggerProdUpdateWorkcenter() {}

// DeleteWorkcenter godoc
// @Summary Удалить рабочий центр
// @Tags workcenters
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 204
// @Router /api/v1/workcenters/{id} [delete]
func swaggerProdDeleteWorkcenter() {}

// ListScrapReasons godoc
// @Summary Причины брака
// @Tags scrap
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/scrap-reasons [get]
func swaggerProdListScrapReasons() {}

// CreateScrapReason godoc
// @Summary Добавить причину брака
// @Tags scrap
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/scrap-reasons [post]
func swaggerProdCreateScrapReason() {}

// ListBOMs godoc
// @Summary Список спецификаций
// @Tags bom
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/boms [get]
func swaggerProdListBOMs() {}

// GetBOM godoc
// @Summary Спецификация по ID
// @Tags bom
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/boms/{id} [get]
func swaggerProdGetBOM() {}

// CreateBOM godoc
// @Summary Создать спецификацию
// @Tags bom
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/boms [post]
func swaggerProdCreateBOM() {}

// PatchBOM godoc
// @Summary Обновить спецификацию
// @Tags bom
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/boms/{id} [patch]
func swaggerProdPatchBOM() {}

// AddBOMLine godoc
// @Summary Добавить строку спецификации
// @Tags bom
// @Security BearerAuth
// @Param id path string true "UUID спецификации"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/boms/{id}/lines [post]
func swaggerProdAddBOMLine() {}

// DeleteBOMLine godoc
// @Summary Удалить строку спецификации
// @Tags bom
// @Security BearerAuth
// @Param id path string true "UUID спецификации"
// @Param line_id path string true "UUID строки"
// @Success 204
// @Router /api/v1/boms/{id}/lines/{line_id} [delete]
func swaggerProdDeleteBOMLine() {}

// SubmitBOM godoc
// @Summary Утвердить спецификацию
// @Tags bom
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/boms/{id}/submit [post]
func swaggerProdSubmitBOM() {}

// ArchiveBOM godoc
// @Summary Архивировать спецификацию
// @Tags bom
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/boms/{id}/archive [post]
func swaggerProdArchiveBOM() {}

// ListRoutings godoc
// @Summary Список техкарт
// @Tags routing
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/routings [get]
func swaggerProdListRoutings() {}

// GetRouting godoc
// @Summary Техкарта по ID
// @Tags routing
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/routings/{id} [get]
func swaggerProdGetRouting() {}

// CreateRouting godoc
// @Summary Создать техкарту
// @Tags routing
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/routings [post]
func swaggerProdCreateRouting() {}

// AddRoutingOperation godoc
// @Summary Добавить операцию в техкарту
// @Tags routing
// @Security BearerAuth
// @Param id path string true "UUID техкарты"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/routings/{id}/operations [post]
func swaggerProdAddRoutingOperation() {}

// SubmitRouting godoc
// @Summary Утвердить техкарту
// @Tags routing
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/routings/{id}/submit [post]
func swaggerProdSubmitRouting() {}

// ListOrders godoc
// @Summary Список производственных заказов
// @Tags orders
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/orders [get]
func swaggerProdListOrders() {}

// GetOrder godoc
// @Summary Заказ по ID
// @Tags orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/orders/{id} [get]
func swaggerProdGetOrder() {}

// CreateOrder godoc
// @Summary Создать заказ
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/orders [post]
func swaggerProdCreateOrder() {}

// ReleaseOrder godoc
// @Summary Выпустить заказ в производство
// @Tags orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/orders/{id}/release [post]
func swaggerProdReleaseOrder() {}

// CancelOrder godoc
// @Summary Отменить заказ
// @Tags orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/orders/{id}/cancel [post]
func swaggerProdCancelOrder() {}

// CompleteOrder godoc
// @Summary Завершить заказ
// @Tags orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/orders/{id}/complete [post]
func swaggerProdCompleteOrder() {}

// ListShiftTasks godoc
// @Summary Сменные задания
// @Tags shift-tasks
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/shift-tasks [get]
func swaggerProdListShiftTasks() {}

// CreateShiftTask godoc
// @Summary Создать сменное задание
// @Tags shift-tasks
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/shift-tasks [post]
func swaggerProdCreateShiftTask() {}

// DeleteShiftTask godoc
// @Summary Удалить сменное задание
// @Tags shift-tasks
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 204
// @Router /api/v1/shift-tasks/{id} [delete]
func swaggerProdDeleteShiftTask() {}

// MeShiftTasks godoc
// @Summary Мои сменные задания
// @Tags shift-tasks
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/me/shift-tasks [get]
func swaggerProdMeShiftTasks() {}

// StartOperation godoc
// @Summary Начать операцию
// @Tags operations
// @Security BearerAuth
// @Param id path string true "UUID заказа"
// @Param op_id path string true "UUID операции"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/orders/{id}/operations/{op_id}/start [post]
func swaggerProdStartOperation() {}

// ReportOperation godoc
// @Summary Отчёт по операции
// @Tags operations
// @Security BearerAuth
// @Param id path string true "UUID заказа"
// @Param op_id path string true "UUID операции"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/orders/{id}/operations/{op_id}/report [post]
func swaggerProdReportOperation() {}

// FinishOperation godoc
// @Summary Завершить операцию
// @Tags operations
// @Security BearerAuth
// @Param id path string true "UUID заказа"
// @Param op_id path string true "UUID операции"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/orders/{id}/operations/{op_id}/finish [post]
func swaggerProdFinishOperation() {}
