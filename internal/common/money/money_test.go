package money_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bitmerchant/internal/common/money"
)

func TestParse(t *testing.T) {
	t.Run("known codes", func(t *testing.T) {
		for _, code := range []string{"USD", "THB", "SAT", "usd", "thb", "sat"} {
			c, err := money.Parse(code)
			require.NoError(t, err)
			assert.NotEmpty(t, c.Code)
		}
	})
	t.Run("empty defaults to USD", func(t *testing.T) {
		c, err := money.Parse("")
		require.NoError(t, err)
		assert.Equal(t, money.USD, c)
	})
	t.Run("unknown errors", func(t *testing.T) {
		_, err := money.Parse("XYZ")
		var unknown money.ErrUnknownCurrency
		require.ErrorAs(t, err, &unknown)
		assert.Equal(t, "XYZ", unknown.Code)
	})
}

func TestAll(t *testing.T) {
	all := money.All()
	require.Len(t, all, 3)
	assert.Equal(t, money.USD, all[0])
	assert.Equal(t, money.SAT, all[2])
}

func TestFormat(t *testing.T) {
	cases := []struct {
		name  string
		money money.Money
		want  string
	}{
		{"USD whole", money.New(1200, money.USD), "$12.00"},
		{"USD with cents", money.New(1234, money.USD), "$12.34"},
		{"THB", money.New(42000, money.THB), "฿420.00"},
		{"SAT small", money.New(500, money.SAT), "500 sats"},
		{"SAT thousands", money.New(5000, money.SAT), "5,000 sats"},
		{"SAT millions", money.New(2_500_000, money.SAT), "2,500,000 sats"},
		{"SAT zero", money.New(0, money.SAT), "0 sats"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.money.Format())
		})
	}
}

func TestFormatNoSymbol(t *testing.T) {
	assert.Equal(t, "12.34", money.New(1234, money.USD).FormatNoSymbol())
	assert.Equal(t, "5000", money.New(5000, money.SAT).FormatNoSymbol())
}

func TestArithmetic(t *testing.T) {
	t.Run("add same currency", func(t *testing.T) {
		sum, err := money.New(100, money.USD).Add(money.New(250, money.USD))
		require.NoError(t, err)
		assert.Equal(t, money.New(350, money.USD), sum)
	})
	t.Run("add mismatch errors", func(t *testing.T) {
		_, err := money.New(100, money.USD).Add(money.New(100, money.SAT))
		require.Error(t, err)
		assert.True(t, errors.Is(err, money.ErrCurrencyMismatch))
	})
	t.Run("mul keeps currency", func(t *testing.T) {
		got := money.New(1000, money.SAT).Mul(3)
		assert.Equal(t, money.New(3000, money.SAT), got)
	})
}

func TestFromMajorRoundsHalfUp(t *testing.T) {
	assert.Equal(t, int64(1235), money.FromMajor(12.345, money.USD).Amount)
	assert.Equal(t, int64(5000), money.FromMajor(5000.0, money.SAT).Amount)
}

func TestNoopConverter(t *testing.T) {
	c := money.NoopConverter{}
	t.Run("same currency passthrough", func(t *testing.T) {
		got, err := c.Convert(context.Background(), money.New(500, money.USD), money.USD)
		require.NoError(t, err)
		assert.Equal(t, money.New(500, money.USD), got)
	})
	t.Run("cross-currency errors", func(t *testing.T) {
		_, err := c.Convert(context.Background(), money.New(500, money.USD), money.SAT)
		require.ErrorIs(t, err, money.ErrConversionNotSupported)
	})
}
