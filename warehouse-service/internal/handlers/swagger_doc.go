package handlers

// Аннотации swag для маршрутов internal/server/router.go.

// Health godoc
// @Summary Liveness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func swaggerWHHealth() {}

// Ready godoc
// @Summary Readiness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]interface{}
// @Router /ready [get]
func swaggerWHReady() {}

// ListProducts godoc
// @Summary Список номенклатуры
// @Tags catalog
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/products [get]
func swaggerWHListProducts() {}

// GetProduct godoc
// @Summary Номенклатура по ID
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID товара"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/products/{id} [get]
func swaggerWHGetProduct() {}

// CreateProduct godoc
// @Summary Создать номенклатуру
// @Tags catalog
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/products [post]
func swaggerWHCreateProduct() {}

// UpdateProduct godoc
// @Summary Обновить номенклатуру
// @Tags catalog
// @Security BearerAuth
// @Accept json
// @Param id path string true "UUID товара"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/products/{id} [put]
func swaggerWHUpdateProduct() {}

// DeleteProduct godoc
// @Summary Удалить номенклатуру
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID товара"
// @Success 204
// @Router /api/v1/products/{id} [delete]
func swaggerWHDeleteProduct() {}

// ListWarehouses godoc
// @Summary Список складов
// @Tags catalog
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/warehouses [get]
func swaggerWHListWarehouses() {}

// CreateWarehouse godoc
// @Summary Создать склад
// @Tags catalog
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/warehouses [post]
func swaggerWHCreateWarehouse() {}

// UpdateWarehouse godoc
// @Summary Обновить склад
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID склада"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/warehouses/{id} [put]
func swaggerWHUpdateWarehouse() {}

// DeleteWarehouse godoc
// @Summary Удалить склад
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID склада"
// @Success 204
// @Router /api/v1/warehouses/{id} [delete]
func swaggerWHDeleteWarehouse() {}

// ListBins godoc
// @Summary Ячейки склада
// @Tags catalog
// @Security BearerAuth
// @Param warehouse_id path string true "UUID склада"
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/warehouses/{warehouse_id}/bins [get]
func swaggerWHListBins() {}

// CreateBin godoc
// @Summary Создать ячейку
// @Tags catalog
// @Security BearerAuth
// @Param warehouse_id path string true "UUID склада"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/warehouses/{warehouse_id}/bins [post]
func swaggerWHCreateBin() {}

// UpdateBin godoc
// @Summary Обновить ячейку
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID ячейки"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/bins/{id} [put]
func swaggerWHUpdateBin() {}

// DeleteBin godoc
// @Summary Удалить ячейку
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID ячейки"
// @Success 204
// @Router /api/v1/bins/{id} [delete]
func swaggerWHDeleteBin() {}

// GetBatch godoc
// @Summary Партия по ID
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID партии"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/batches/{id} [get]
func swaggerWHGetBatch() {}

// ListPrices godoc
// @Summary Цены товара
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID товара"
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/products/{id}/prices [get]
func swaggerWHListPrices() {}

// CreatePrice godoc
// @Summary Добавить цену
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID товара"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/products/{id}/prices [post]
func swaggerWHCreatePrice() {}

// DeletePrice godoc
// @Summary Удалить цену
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID цены"
// @Success 204
// @Router /api/v1/prices/{id} [delete]
func swaggerWHDeletePrice() {}

// ListSerials godoc
// @Summary Серийные номера
// @Tags catalog
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/serials [get]
func swaggerWHListSerials() {}

// SerialHistory godoc
// @Summary История серийного номера
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "UUID серии"
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/serials/{id}/history [get]
func swaggerWHSerialHistory() {}

// Balances godoc
// @Summary Остатки
// @Tags reports
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/balances [get]
func swaggerWHBalances() {}

// Movements godoc
// @Summary Движения
// @Tags reports
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/movements [get]
func swaggerWHMovements() {}

// StockOnDate godoc
// @Summary Остатки на дату
// @Tags reports
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/reports/stock-on-date [get]
func swaggerWHStockOnDate() {}

// Turnover godoc
// @Summary Оборотная ведомость
// @Tags reports
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/reports/turnover [get]
func swaggerWHTurnover() {}

// ABC godoc
// @Summary ABC-анализ
// @Tags reports
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/reports/abc [get]
func swaggerWHABC() {}

// Expiring godoc
// @Summary Истекающие партии
// @Tags reports
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/reports/expiring [get]
func swaggerWHExpiring() {}

// PriceOnDate godoc
// @Summary Цена на дату
// @Tags reports
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/reports/price-on-date [get]
func swaggerWHPriceOnDate() {}

// AvgCostOnDate godoc
// @Summary Средняя себестоимость на дату
// @Tags reports
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/reports/average-cost [get]
func swaggerWHAvgCostOnDate() {}

// ListReservations godoc
// @Summary Список резервов
// @Tags reservations
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/reservations [get]
func swaggerWHListReservations() {}

// GetReservation godoc
// @Summary Резерв по ID
// @Tags reservations
// @Security BearerAuth
// @Param id path string true "UUID резерва"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/reservations/{id} [get]
func swaggerWHGetReservation() {}

// CreateReservation godoc
// @Summary Создать резерв
// @Tags reservations
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/reservations [post]
func swaggerWHCreateReservation() {}

// ReleaseReservation godoc
// @Summary Снять резерв
// @Tags reservations
// @Security BearerAuth
// @Param id path string true "UUID резерва"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/reservations/{id}/release [post]
func swaggerWHReleaseReservation() {}

// ConsumeReservation godoc
// @Summary Списать резерв
// @Tags reservations
// @Security BearerAuth
// @Param id path string true "UUID резерва"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/reservations/{id}/consume [post]
func swaggerWHConsumeReservation() {}

// GetInventory godoc
// @Summary Инвентаризация по ID
// @Tags inventory
// @Security BearerAuth
// @Param id path string true "UUID инвентаризации"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/inventory/{id} [get]
func swaggerWHGetInventory() {}

// StartInventory godoc
// @Summary Начать инвентаризацию
// @Tags inventory
// @Security BearerAuth
// @Accept json
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/inventory [post]
func swaggerWHStartInventory() {}

// SetInventoryCounted godoc
// @Summary Указать фактическое количество
// @Tags inventory
// @Security BearerAuth
// @Param line_id path string true "UUID строки"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/inventory/lines/{line_id} [patch]
func swaggerWHSetInventoryCounted() {}

// PostInventory godoc
// @Summary Провести инвентаризацию
// @Tags inventory
// @Security BearerAuth
// @Param id path string true "UUID инвентаризации"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/inventory/{id}/post [post]
func swaggerWHPostInventory() {}

// Receipt godoc
// @Summary Приход
// @Tags operations
// @Security BearerAuth
// @Accept json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/operations/receipt [post]
func swaggerWHReceipt() {}

// Issue godoc
// @Summary Расход
// @Tags operations
// @Security BearerAuth
// @Accept json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/operations/issue [post]
func swaggerWHIssue() {}

// IssueFromReservations godoc
// @Summary Расход по резервам
// @Tags operations
// @Security BearerAuth
// @Accept json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/operations/issue-from-reservations [post]
func swaggerWHIssueFromReservations() {}

// Transfer godoc
// @Summary Перемещение между складами
// @Tags operations
// @Security BearerAuth
// @Accept json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/operations/transfer [post]
func swaggerWHTransfer() {}

// Relocate godoc
// @Summary Перемещение внутри склада
// @Tags operations
// @Security BearerAuth
// @Accept json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/operations/relocate [post]
func swaggerWHRelocate() {}

// ImportProductsCSV godoc
// @Summary Импорт номенклатуры из CSV
// @Tags import
// @Security BearerAuth
// @Accept multipart/form-data
// @Success 202 {object} map[string]interface{}
// @Router /api/v1/import/products [post]
func swaggerWHImportProductsCSV() {}

// GetImportJob godoc
// @Summary Статус импорта
// @Tags import
// @Security BearerAuth
// @Param id path string true "UUID задания"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/import/jobs/{id} [get]
func swaggerWHGetImportJob() {}

// ExportMovementsCSV godoc
// @Summary Экспорт движений в CSV
// @Tags import
// @Security BearerAuth
// @Produce text/csv
// @Success 200 {file} file
// @Router /api/v1/export/movements.csv [get]
func swaggerWHExportMovementsCSV() {}
