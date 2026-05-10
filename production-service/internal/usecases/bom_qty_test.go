package usecases

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/industrial-sed/production-service/internal/models"
)

func TestBomLineQtyNeed(t *testing.T) {
	ln := models.BOMLine{
		ID:                 uuid.New(),
		QtyPer:             "2",
		ScrapPct:           "10",
	}
	qp := decimal.RequireFromString("100")
	got := bomLineQtyNeed(ln, qp)
	require.Equal(t, "220", got.StringFixed(0))
}
