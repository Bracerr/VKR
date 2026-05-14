package usecases

import (
	"github.com/industrial-sed/sed-service/internal/clients"
	"github.com/industrial-sed/sed-service/internal/repositories"
)

// App сценарии СЭД (справочники, документы, файлы).
type App struct {
	Store *repositories.Store
	WH    *clients.Warehouse
	Minio *clients.Minio
	Prod  *clients.ProductionCallback // опционально: уведомление production-service после подписи
	Proc  *clients.ProcurementCallback // опционально: уведомление procurement-service после подписи
	Sales *clients.SalesCallback // опционально: уведомление sales-service после подписи
}
