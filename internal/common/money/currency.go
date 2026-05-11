// Package money provides Currency and Money value objects with strict
// per-currency arithmetic. Amounts are stored as int64 minor units
// (cents for fiat, sats for SAT). Operations refuse to mix currencies —
// conversion goes through the explicit Converter boundary.
package money

import (
	"fmt"
	"strings"
)

// Currency identifies a unit of account.
type Currency struct {
	Code   string
	Symbol string
	Scale  int
}

var (
	USD = Currency{Code: "USD", Symbol: "$", Scale: 2}
	THB = Currency{Code: "THB", Symbol: "฿", Scale: 2}
	SAT = Currency{Code: "SAT", Symbol: "sats", Scale: 0}
)

var registry = map[string]Currency{
	USD.Code: USD,
	THB.Code: THB,
	SAT.Code: SAT,
}

// ErrUnknownCurrency is returned by Parse when the code is not in the registry.
type ErrUnknownCurrency struct{ Code string }

func (e ErrUnknownCurrency) Error() string {
	return fmt.Sprintf("unknown currency: %q", e.Code)
}

// Parse returns the registered Currency for the given code (case-insensitive).
// An empty code defaults to USD so legacy callers and rows backfilled with the
// schema default behave correctly.
func Parse(code string) (Currency, error) {
	if code == "" {
		return USD, nil
	}
	if c, ok := registry[strings.ToUpper(code)]; ok {
		return c, nil
	}
	return Currency{}, ErrUnknownCurrency{Code: code}
}

// MustParse is Parse without the error return — for tests and constants.
func MustParse(code string) Currency {
	c, err := Parse(code)
	if err != nil {
		panic(err)
	}
	return c
}

// All returns the registered currencies in a stable order suitable for UI
// dropdowns. USD first (default), then THB, then SAT.
func All() []Currency {
	return []Currency{USD, THB, SAT}
}

// IsZero reports whether the Currency is the zero value (uninitialized).
func (c Currency) IsZero() bool {
	return c.Code == ""
}
