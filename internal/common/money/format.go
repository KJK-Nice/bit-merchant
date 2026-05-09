package money

import (
	"fmt"
	"strconv"
	"strings"
)

// Format renders Money for human display.
//   - SAT: "5,000 sats" (thousands separator, no decimals, suffix label).
//   - Fiat (USD/THB/...): "$12.34" / "฿420.00" (symbol prefix, fixed decimals).
func (m Money) Format() string {
	if m.Currency.Code == SAT.Code {
		return fmt.Sprintf("%s sats", thousands(m.Amount))
	}
	return fmt.Sprintf("%s%.*f", m.Currency.Symbol, m.Currency.Scale, m.Major())
}

// FormatNoSymbol renders the bare numeric amount in major units, suitable for
// pre-filling input fields.
func (m Money) FormatNoSymbol() string {
	if m.Currency.Code == SAT.Code {
		return strconv.FormatInt(m.Amount, 10)
	}
	return fmt.Sprintf("%.*f", m.Currency.Scale, m.Major())
}

func thousands(n int64) string {
	negative := n < 0
	if negative {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		if negative {
			return "-" + s
		}
		return s
	}
	var b strings.Builder
	first := len(s) % 3
	if first == 0 {
		first = 3
	}
	b.WriteString(s[:first])
	for i := first; i < len(s); i += 3 {
		b.WriteByte(',')
		b.WriteString(s[i : i+3])
	}
	if negative {
		return "-" + b.String()
	}
	return b.String()
}
