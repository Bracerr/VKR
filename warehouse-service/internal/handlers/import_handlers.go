package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/industrial-sed/warehouse-service/internal/config"
	"github.com/industrial-sed/warehouse-service/internal/usecases"
)

// Imp импорт/экспорт.
type Imp struct {
	UC  *usecases.UC
	Cfg *config.Config
}

func (h *Imp) ImportProductsCSV(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, err := h.UC.CreateImportJob(c.Request.Context(), tn, userName(c), "PRODUCTS", c.Request.Body, h.Cfg.ImportMaxRows)
	if err != nil {
		if RespondUsecaseError(c, err) {
			return
		}
		RespondError(c, http.StatusBadRequest, err.Error(), http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"job_id": id})
}

func (h *Imp) GetImportJob(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	j, err := h.UC.GetImportJob(c.Request.Context(), tn, id)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	if j == nil {
		RespondError(c, http.StatusNotFound, "не найдено", http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, j)
}

func (h *Imp) ExportMovementsCSV(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	from, err := time.Parse(time.RFC3339, c.Query("from"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, "from", http.StatusBadRequest)
		return
	}
	to, err := time.Parse(time.RFC3339, c.Query("to"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, "to", http.StatusBadRequest)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=movements.csv")
	_ = h.UC.ExportMovementsCSV(c.Request.Context(), c.Writer, tn, from, to)
}
