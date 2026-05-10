package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/industrial-sed/warehouse-service/internal/usecases"
)

// Rep отчёты и остатки.
type Rep struct {
	UC *usecases.UC
}

func (h *Rep) Balances(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	wh := optionalUUIDQuery(c, "warehouse_id")
	bin := optionalUUIDQuery(c, "bin_id")
	pr := optionalUUIDQuery(c, "product_id")
	bat := optionalUUIDQuery(c, "batch_id")
	onlyPos := c.DefaultQuery("only_positive", "true") == "true"
	var expBefore *time.Time
	if s := c.Query("expires_before"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err == nil {
			expBefore = &t
		}
	}
	list, err := h.UC.ListBalances(c.Request.Context(), tn, wh, bin, pr, bat, onlyPos, expBefore)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Rep) Movements(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	from, err := time.Parse(time.RFC3339, c.Query("from"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, "from (RFC3339)", http.StatusBadRequest)
		return
	}
	to, err := time.Parse(time.RFC3339, c.Query("to"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, "to (RFC3339)", http.StatusBadRequest)
		return
	}
	var wh, pr *uuid.UUID
	if id := optionalUUIDQuery(c, "warehouse_id"); id != nil {
		wh = id
	}
	if id := optionalUUIDQuery(c, "product_id"); id != nil {
		pr = id
	}
	var mt *string
	if s := c.Query("movement_type"); s != "" {
		mt = &s
	}
	limit := 500
	list, err := h.UC.ListMovements(c.Request.Context(), tn, from, to, wh, pr, mt, limit)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Rep) StockOnDate(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	at, err := time.Parse(time.RFC3339, c.Query("at"))
	if err != nil {
		at, err = time.Parse("2006-01-02", c.Query("at"))
		if err != nil {
			RespondError(c, http.StatusBadRequest, "at", http.StatusBadRequest)
			return
		}
	}
	var wh, pr *uuid.UUID
	if id := optionalUUIDQuery(c, "warehouse_id"); id != nil {
		wh = id
	}
	if id := optionalUUIDQuery(c, "product_id"); id != nil {
		pr = id
	}
	rows, err := h.UC.StockOnDate(c.Request.Context(), tn, at, wh, pr)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *Rep) Turnover(c *gin.Context) {
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
	gb := c.DefaultQuery("group_by", "product")
	rows, err := h.UC.Turnover(c.Request.Context(), tn, from, to, gb)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *Rep) ABC(c *gin.Context) {
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
	metric := c.DefaultQuery("metric", "qty")
	rows, err := h.UC.ABCAnalysis(c.Request.Context(), tn, from, to, metric)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *Rep) Expiring(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	until, err := time.Parse("2006-01-02", c.Query("until"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, "until (date)", http.StatusBadRequest)
		return
	}
	rows, err := h.UC.ExpiringBatches(c.Request.Context(), tn, until)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *Rep) PriceOnDate(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	pid, err := uuid.Parse(c.Query("product_id"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, "product_id", http.StatusBadRequest)
		return
	}
	pt := c.Query("price_type")
	if pt == "" {
		pt = "SALE"
	}
	on, err := time.Parse("2006-01-02", c.Query("on"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, "on (date)", http.StatusBadRequest)
		return
	}
	p, err := h.UC.PriceOnDate(c.Request.Context(), tn, pid, pt, on)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"price": p})
}

func (h *Rep) AvgCostOnDate(c *gin.Context) {
	tn, ok := tenant(c)
	if !ok {
		return
	}
	pid, err := uuid.Parse(c.Query("product_id"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, "product_id", http.StatusBadRequest)
		return
	}
	at, err2 := time.Parse(time.RFC3339, c.Query("at"))
	if err2 != nil {
		at, err2 = time.Parse("2006-01-02", c.Query("at"))
		if err2 != nil {
			RespondError(c, http.StatusBadRequest, "at", http.StatusBadRequest)
			return
		}
	}
	x, err := h.UC.AverageCostOnDate(c.Request.Context(), tn, at, pid)
	if err != nil {
		RespondUsecaseError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"average_unit_cost": x})
}
