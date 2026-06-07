package handlers

// Аннотации swag для маршрутов internal/server/router.go.

// Health godoc
// @Summary Liveness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func swaggerSalesHealth() {}

// Ready godoc
// @Summary Readiness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]interface{}
// @Router /ready [get]
func swaggerSalesReady() {}

// PostSedEvents godoc
// @Summary События от СЭД (internal)
// @Tags internal
// @Accept json
// @Param X-Service-Secret header string true "Секрет сервиса"
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/internal/sed-events [post]
func swaggerSalesPostSedEvents() {}

// ListCustomers godoc
// @Summary Список клиентов
// @Tags customers
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/customers [get]
func swaggerSalesListCustomers() {}

// CreateCustomer godoc
// @Summary Создать клиента
// @Tags customers
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/customers [post]
func swaggerSalesCreateCustomer() {}

// UpdateCustomer godoc
// @Summary Обновить клиента
// @Tags customers
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/customers/{id} [put]
func swaggerSalesUpdateCustomer() {}

// DeleteCustomer godoc
// @Summary Удалить клиента
// @Tags customers
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 204
// @Router /api/v1/customers/{id} [delete]
func swaggerSalesDeleteCustomer() {}

// ListSO godoc
// @Summary Список заказов на продажу
// @Tags sales-orders
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/sales-orders [get]
func swaggerSalesListSO() {}

// GetSO godoc
// @Summary Заказ на продажу по ID
// @Tags sales-orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sales-orders/{id} [get]
func swaggerSalesGetSO() {}

// CreateSO godoc
// @Summary Создать заказ на продажу
// @Tags sales-orders
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/sales-orders [post]
func swaggerSalesCreateSO() {}

// AddSOLine godoc
// @Summary Добавить строку заказа
// @Tags sales-orders
// @Security BearerAuth
// @Param id path string true "UUID заказа"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/sales-orders/{id}/lines [post]
func swaggerSalesAddSOLine() {}

// SubmitSO godoc
// @Summary Отправить заказ на согласование
// @Tags sales-orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sales-orders/{id}/submit [post]
func swaggerSalesSubmitSO() {}

// ReleaseSO godoc
// @Summary Утвердить заказ
// @Tags sales-orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sales-orders/{id}/release [post]
func swaggerSalesReleaseSO() {}

// CancelSO godoc
// @Summary Отменить заказ
// @Tags sales-orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sales-orders/{id}/cancel [post]
func swaggerSalesCancelSO() {}

// ReserveSO godoc
// @Summary Зарезервировать товар по заказу
// @Tags sales-orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sales-orders/{id}/reserve [post]
func swaggerSalesReserveSO() {}

// ShipSO godoc
// @Summary Отгрузить заказ
// @Tags sales-orders
// @Security BearerAuth
// @Param id path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sales-orders/{id}/ship [post]
func swaggerSalesShipSO() {}
