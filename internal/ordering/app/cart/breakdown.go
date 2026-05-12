package cart

import (
	"math"

	"bitmerchant/internal/common/money"
)

// AllowedTipPercents are the four canonical tip tiers offered on confirm.
var AllowedTipPercents = []int{0, 10, 15, 20}

// DefaultTipPercent is selected on first render of the confirm page.
const DefaultTipPercent = 15

// Breakdown captures the four lines shown on the confirm page receipt:
// Subtotal + Tax + Tip = Total. All amounts share the cart's currency.
type Breakdown struct {
	Subtotal money.Money
	Tax      money.Money
	Tip      money.Money
	Total    money.Money
	Currency money.Currency
}

// IsAllowedTipPercent reports whether p is one of the canonical tip tiers.
func IsAllowedTipPercent(p int) bool {
	for _, allowed := range AllowedTipPercents {
		if p == allowed {
			return true
		}
	}
	return false
}

// ComputeBreakdown builds the Subtotal/Tax/Tip/Total breakdown for the cart.
// Tax is computed against the pre-tip subtotal; tip is computed against the
// pre-tax subtotal (US tipping convention). Both rounded to minor units.
func ComputeBreakdown(c *Cart, taxRate float64, tipPercent int) Breakdown {
	cur := c.Currency
	if cur.IsZero() {
		cur = money.USD
	}
	subtotal := money.FromMajor(c.Total, cur)
	tax := money.FromMajor(round(c.Total*taxRate, cur.Scale), cur)
	tip := money.FromMajor(round(c.Total*float64(tipPercent)/100.0, cur.Scale), cur)

	// Add is safe — all three share cur. Errors only on currency mismatch.
	withTax, _ := subtotal.Add(tax)
	total, _ := withTax.Add(tip)

	return Breakdown{
		Subtotal: subtotal,
		Tax:      tax,
		Tip:      tip,
		Total:    total,
		Currency: cur,
	}
}

// round rounds f to `scale` decimal places to match minor-unit precision.
// Without this, math.Round inside FromMajor can drift one minor unit due to
// float64 representation (e.g. 22.25 * 0.08 == 1.7800000000000002).
func round(f float64, scale int) float64 {
	mult := math.Pow10(scale)
	return math.Round(f*mult) / mult
}
