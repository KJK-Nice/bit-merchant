package payment

import (
	"context"
	"errors"
	"fmt"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/strike"
)

// CheckPaymentStatusUseCase checks payment status via Strike API
type CheckPaymentStatusUseCase struct {
	strikeClient *strike.Client
	paymentRepo  domain.PaymentRepository
}

// NewCheckPaymentStatusUseCase creates a new CheckPaymentStatusUseCase
func NewCheckPaymentStatusUseCase(
	strikeClient *strike.Client,
	paymentRepo domain.PaymentRepository,
) *CheckPaymentStatusUseCase {
	return &CheckPaymentStatusUseCase{
		strikeClient: strikeClient,
		paymentRepo:  paymentRepo,
	}
}

// CheckStatusResponse represents payment status check response
type CheckStatusResponse struct {
	PaymentID domain.PaymentID
	Status    domain.PaymentStatus
	InvoiceID string
}

// Execute checks payment status
func (uc *CheckPaymentStatusUseCase) Execute(ctx context.Context, invoiceID string) (*CheckStatusResponse, error) {
	// Get payment from repository
	payment, err := uc.paymentRepo.FindByInvoiceID(invoiceID)
	if err != nil {
		return nil, errors.New("payment not found")
	}

	// Check status via Strike API
	statusResp, err := uc.strikeClient.GetInvoiceStatus(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to check invoice status: %w", err)
	}

	// Update payment status based on Strike response
	switch statusResp.State {
	case "PAID":
		if payment.Status != domain.PaymentStatusPaid {
			payment.MarkAsPaid()
			_ = uc.paymentRepo.Update(payment)
		}
	case "UNPAID", "PENDING":
		// Status remains pending
	case "CANCELLED", "EXPIRED":
		if payment.Status != domain.PaymentStatusExpired {
			payment.MarkAsExpired("Invoice expired or cancelled")
			_ = uc.paymentRepo.Update(payment)
		}
	default:
		// Unknown status, keep current status
	}

	return &CheckStatusResponse{
		PaymentID: payment.ID,
		Status:    payment.Status,
		InvoiceID: invoiceID,
	}, nil
}
