package handlers

import (
	"github.com/industrial-sed/sed-service/internal/usecases"
)

// HTTP зависимости хендлеров.
type HTTP struct {
	App *usecases.App
}
