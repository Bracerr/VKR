package usecases

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestABCClass(t *testing.T) {
	total := decimal.NewFromInt(100)
	require.Equal(t, "A", abcClass(decimal.NewFromInt(50), total))
	require.Equal(t, "A", abcClass(decimal.NewFromInt(80), total))
	require.Equal(t, "B", abcClass(decimal.NewFromInt(81), total))
	require.Equal(t, "B", abcClass(decimal.NewFromInt(95), total))
	require.Equal(t, "C", abcClass(decimal.NewFromInt(96), total))
}
