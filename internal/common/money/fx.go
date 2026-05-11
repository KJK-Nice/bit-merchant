package money

import (
	"context"
	"errors"
)

// ErrConversionNotSupported indicates the Converter cannot exchange between
// the requested currencies. The Lightning rollout will replace NoopConverter
// with a real rate provider; until then, only same-currency "conversions"
// succeed.
var ErrConversionNotSupported = errors.New("currency conversion not supported")

// Converter exchanges Money between currencies. Implementations should be
// safe for concurrent use.
type Converter interface {
	Convert(ctx context.Context, from Money, to Currency) (Money, error)
}

// NoopConverter is a Converter that only succeeds when the source and target
// currencies match. It is the default until an FX provider is wired.
type NoopConverter struct{}

func (NoopConverter) Convert(_ context.Context, from Money, to Currency) (Money, error) {
	if from.Currency.Code == to.Code {
		return from, nil
	}
	return Money{}, ErrConversionNotSupported
}
