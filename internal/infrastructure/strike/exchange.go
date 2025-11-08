package strike

import (
	"context"
	"fmt"
)

// GetExchangeRateResponse represents exchange rate response
type GetExchangeRateResponse struct {
	Currency        string  `json:"currency"`
	Rate            float64 `json:"rate"`
	SatoshisPerUnit int64   `json:"satoshisPerUnit"`
}

// GetExchangeRate gets BTC exchange rate for currency via Strike API
func (c *Client) GetExchangeRate(ctx context.Context, currency string) (*GetExchangeRateResponse, error) {
	endpoint := fmt.Sprintf("/rates/%s", currency)
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result GetExchangeRateResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
