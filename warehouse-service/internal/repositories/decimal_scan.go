package repositories

import (
	"github.com/shopspring/decimal"
)

func decPtr(s *string) *decimal.Decimal {
	if s == nil || *s == "" {
		return nil
	}
	d, err := decimal.NewFromString(*s)
	if err != nil {
		return nil
	}
	return &d
}

func decStr(d *decimal.Decimal) interface{} {
	if d == nil {
		return nil
	}
	return d.StringFixed(4)
}
