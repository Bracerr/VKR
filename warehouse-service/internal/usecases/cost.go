package usecases

import (
	"github.com/shopspring/decimal"
)

// MovementValue рассчитывает стоимость строки движения для расхода (средняя по остатку).
func MovementValue(qty, balanceQty, balanceVal decimal.Decimal) decimal.Decimal {
	if balanceQty.IsZero() {
		return decimal.Zero
	}
	unit := balanceVal.Div(balanceQty)
	return qty.Mul(unit)
}

// ReceiptLineValue стоимость прихода.
func ReceiptLineValue(qty decimal.Decimal, unitCost *decimal.Decimal) decimal.Decimal {
	if unitCost == nil {
		return decimal.Zero
	}
	return qty.Mul(*unitCost)
}
