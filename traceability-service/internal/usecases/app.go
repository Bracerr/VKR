package usecases

import (
	"github.com/industrial-sed/traceability-service/internal/config"
	"github.com/industrial-sed/traceability-service/internal/repositories"
)

type App struct {
	Store *repositories.Store
	Cfg   *config.Config
}

