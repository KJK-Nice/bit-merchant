package strike

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// CreateInvoiceRequest represents Strike invoice creation request
type CreateInvoiceRequest struct {
	Amount        Amount `json:"amount"`
	Description   string `json:"description"`
	CorrelationID string `json:"correlationId,omitempty"`
}

// Amount represents currency and amount
type Amount struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

// CreateInvoiceResponse represents Strike invoice creation response
type CreateInvoiceResponse struct {
	InvoiceID string `json:"invoiceId"`
	Invoice   string `json:"invoice"`
	State     string `json:"state"`
}

// CreateInvoice creates a Lightning invoice via Strike API
func (c *Client) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/invoices", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var result CreateInvoiceResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetInvoiceStatusResponse represents invoice status response
type GetInvoiceStatusResponse struct {
	InvoiceID string `json:"invoiceId"`
	State     string `json:"state"`
}

// GetInvoiceStatus checks invoice payment status via Strike API
func (c *Client) GetInvoiceStatus(ctx context.Context, invoiceID string) (*GetInvoiceStatusResponse, error) {
	endpoint := fmt.Sprintf("/invoices/%s", invoiceID)
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result GetInvoiceStatusResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
