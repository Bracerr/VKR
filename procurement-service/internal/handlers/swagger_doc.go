package handlers

// Аннотации swag для маршрутов internal/server/router.go.

// Health godoc
// @Summary Liveness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func swaggerProcHealth() {}

// Ready godoc
// @Summary Readiness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]interface{}
// @Router /ready [get]
func swaggerProcReady() {}

// PostSedEvents godoc
// @Summary События от СЭД (internal)
// @Tags internal
// @Accept json
// @Param X-Service-Secret header string true "Секрет сервиса"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/internal/sed-events [post]
func swaggerProcPostSedEvents() {}

// ListSuppliers godoc
// @Summary Список поставщиков
// @Tags suppliers
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/suppliers [get]
func swaggerProcListSuppliers() {}

// CreateSupplier godoc
// @Summary Создать поставщика
// @Tags suppliers
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/suppliers [post]
func swaggerProcCreateSupplier() {}

// UpdateSupplier godoc
// @Summary Обновить поставщика
// @Tags suppliers
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/suppliers/{id} [put]
func swaggerProcUpdateSupplier() {}

// DeleteSupplier godoc
// @Summary Удалить поставщика
// @Tags suppliers
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 204
// @Router /api/v1/suppliers/{id} [delete]
func swaggerProcDeleteSupplier() {}

// ListPR godoc
// @Summary Список заявок на закупку
// @Tags purchase-requests
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/purchase-requests [get]
func swaggerProcListPR() {}

// GetPR godoc
// @Summary Заявка на закупку по ID
// @Tags purchase-requests
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/purchase-requests/{id} [get]
func swaggerProcGetPR() {}

// CreatePR godoc
// @Summary Создать заявку на закупку
// @Tags purchase-requests
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/purchase-requests [post]
func swaggerProcCreatePR() {}

// AddPRLine godoc
// @Summary Добавить строку заявки
// @Tags purchase-requests
// @Security BearerAuth
// @Param id path string true "UUID заявки"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/purchase-requests/{id}/lines [post]
func swaggerProcAddPRLine() {}

// SubmitPR godoc
// @Summary Отправить заявку на согласование
// @Tags purchase-requests
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/purchase-requests/{id}/submit [post]
func swaggerProcSubmitPR() {}

// CancelPR godoc
// @Summary Отменить заявку
// @Tags purchase-requests
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/purchase-requests/{id}/cancel [post]
func swaggerProcCancelPR() {}

// ListPO godoc
// @Summary Список заказов поставщику
// @Tags purchase-orders
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/purchase-orders [get]
func swaggerProcListPO() {}

// GetPO godoc
// @Summary Заказ поставщику по ID
// @Tags purchase-orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/purchase-orders/{id} [get]
func swaggerProcGetPO() {}

// CreatePO godoc
// @Summary Создать заказ поставщику
// @Tags purchase-orders
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/purchase-orders [post]
func swaggerProcCreatePO() {}

// CreatePOFromPR godoc
// @Summary Создать заказ из заявки
// @Tags purchase-orders
// @Security BearerAuth
// @Param id path string true "UUID заявки"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/purchase-orders/from-pr/{id} [post]
func swaggerProcCreatePOFromPR() {}

// AddPOLine godoc
// @Summary Добавить строку заказа
// @Tags purchase-orders
// @Security BearerAuth
// @Param id path string true "UUID заказа"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/purchase-orders/{id}/lines [post]
func swaggerProcAddPOLine() {}

// SubmitPO godoc
// @Summary Отправить заказ на согласование
// @Tags purchase-orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/purchase-orders/{id}/submit [post]
func swaggerProcSubmitPO() {}

// ReleasePO godoc
// @Summary Утвердить заказ
// @Tags purchase-orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/purchase-orders/{id}/release [post]
func swaggerProcReleasePO() {}

// CancelPO godoc
// @Summary Отменить заказ
// @Tags purchase-orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/purchase-orders/{id}/cancel [post]
func swaggerProcCancelPO() {}

// ReceivePO godoc
// @Summary Принять товар по заказу
// @Tags purchase-orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/purchase-orders/{id}/receive [post]
func swaggerProcReceivePO() {}
