package cart_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"bitmerchant/internal/common/money"
	"bitmerchant/internal/ordering/app/cart"
)

func TestComputeBreakdown_USDDefaultTip(t *testing.T) {
	c := &cart.Cart{Total: 22.25, Currency: money.USD}

	bd := cart.ComputeBreakdown(c, 0.08, 15)

	// Subtotal $22.25, tax 8% = $1.78, tip 15% = $3.34, total = $27.37.
	assert.Equal(t, int64(2225), bd.Subtotal.Amount, "subtotal in cents")
	assert.Equal(t, int64(178), bd.Tax.Amount, "tax @ 8%")
	assert.Equal(t, int64(334), bd.Tip.Amount, "tip @ 15%")
	assert.Equal(t, int64(2737), bd.Total.Amount, "total = subtotal + tax + tip")
	assert.Equal(t, "$22.25", bd.Subtotal.Format())
	assert.Equal(t, "$1.78", bd.Tax.Format())
	assert.Equal(t, "$3.34", bd.Tip.Format())
	assert.Equal(t, "$27.37", bd.Total.Format())
}

func TestComputeBreakdown_NoTip(t *testing.T) {
	c := &cart.Cart{Total: 22.25, Currency: money.USD}

	bd := cart.ComputeBreakdown(c, 0.08, 0)

	assert.Equal(t, int64(0), bd.Tip.Amount)
	assert.Equal(t, int64(2403), bd.Total.Amount, "$22.25 + $1.78 tax + $0 tip = $24.03")
	assert.Equal(t, "$24.03", bd.Total.Format())
}

func TestComputeBreakdown_AllTipTiers(t *testing.T) {
	c := &cart.Cart{Total: 22.25, Currency: money.USD}

	cases := map[int]int64{
		0:  0,
		10: 223, // round(22.25 * 0.10 * 100) = 222.5 → 223 (banker's: round-half-away)
		15: 334,
		20: 445,
	}
	for tip, wantCents := range cases {
		bd := cart.ComputeBreakdown(c, 0.08, tip)
		assert.Equal(t, wantCents, bd.Tip.Amount, "tip %d%%", tip)
	}
}

func TestComputeBreakdown_SATCurrency(t *testing.T) {
	c := &cart.Cart{Total: 15_000, Currency: money.SAT}

	bd := cart.ComputeBreakdown(c, 0.08, 15)

	// SAT has scale 0, so amounts are integer sats: 1,200 tax, 2,250 tip.
	assert.Equal(t, int64(15_000), bd.Subtotal.Amount)
	assert.Equal(t, int64(1_200), bd.Tax.Amount)
	assert.Equal(t, int64(2_250), bd.Tip.Amount)
	assert.Equal(t, int64(18_450), bd.Total.Amount)
	assert.Equal(t, "18,450 sats", bd.Total.Format())
}

func TestComputeBreakdown_ZeroCart(t *testing.T) {
	c := &cart.Cart{Total: 0, Currency: money.USD}

	bd := cart.ComputeBreakdown(c, 0.08, 15)

	assert.Equal(t, int64(0), bd.Subtotal.Amount)
	assert.Equal(t, int64(0), bd.Tax.Amount)
	assert.Equal(t, int64(0), bd.Tip.Amount)
	assert.Equal(t, int64(0), bd.Total.Amount)
}

func TestIsAllowedTipPercent(t *testing.T) {
	for _, ok := range []int{0, 10, 15, 20} {
		assert.True(t, cart.IsAllowedTipPercent(ok))
	}
	for _, bad := range []int{-1, 5, 12, 18, 25, 100} {
		assert.False(t, cart.IsAllowedTipPercent(bad))
	}
}
