package handlers

import (
	"github.com/shopspring/decimal"
)

func parseDecimal(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}
