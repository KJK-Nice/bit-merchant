package money

import (
	"errors"
	"fmt"
	"math"
)

// Money is an amount in a specific currency, stored as int64 minor units
// (cents for USD/THB, sats for SAT).
type Money struct {
	Amount   int64
	Currency Currency
}

// New constructs Money from an amount already in minor units.
func New(amount int64, c Currency) Money {
	return Money{Amount: amount, Currency: c}
}

// FromMajor converts a major-unit float (e.g. 12.34 USD) into Money. Used
// only on read paths that still hold legacy float columns — new code should
// stay in minor units end-to-end.
func FromMajor(major float64, c Currency) Money {
	scale := math.Pow10(c.Scale)
	return Money{Amount: int64(math.Round(major * scale)), Currency: c}
}

// Major returns the amount as a major-unit float. Lossy for SAT past 2^53;
// avoid in arithmetic.
func (m Money) Major() float64 {
	return float64(m.Amount) / math.Pow10(m.Currency.Scale)
}

// IsZero reports whether the amount is exactly zero.
func (m Money) IsZero() bool { return m.Amount == 0 }

// IsPositive reports whether the amount is strictly greater than zero.
func (m Money) IsPositive() bool { return m.Amount > 0 }

// ErrCurrencyMismatch is returned when two Money values of different
// currencies are combined without an explicit conversion.
var ErrCurrencyMismatch = errors.New("currency mismatch")

// Add returns m + other. Errors if currencies differ.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency.Code != other.Currency.Code {
		return Money{}, fmt.Errorf("%w: %s vs %s", ErrCurrencyMismatch, m.Currency.Code, other.Currency.Code)
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Sub returns m - other. Errors if currencies differ.
func (m Money) Sub(other Money) (Money, error) {
	if m.Currency.Code != other.Currency.Code {
		return Money{}, fmt.Errorf("%w: %s vs %s", ErrCurrencyMismatch, m.Currency.Code, other.Currency.Code)
	}
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}, nil
}

// Mul returns m multiplied by an integer factor (e.g. quantity).
func (m Money) Mul(factor int) Money {
	return Money{Amount: m.Amount * int64(factor), Currency: m.Currency}
}
