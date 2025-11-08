package domain_test

import (
	"testing"

	"bitmerchant/internal/domain"
)

func TestNewPayment_Validation(t *testing.T) {
	tests := []struct {
		name          string
		id            domain.PaymentID
		restaurantID  domain.RestaurantID
		orderID       *domain.OrderID
		invoiceID     string
		invoice       string
		amountSatoshis int64
		amountFiat    float64
		exchangeRate  float64
		wantError     bool
	}{
		{
			name:          "valid payment",
			id:            "pay_001",
			restaurantID:  "rest_001",
			orderID:       nil,
			invoiceID:     "inv_001",
			invoice:       "lnbc123...",
			amountSatoshis: 1000,
			amountFiat:    10.00,
			exchangeRate:  100000,
			wantError:     false,
		},
		{
			name:          "empty invoice ID should fail",
			id:            "pay_002",
			restaurantID:  "rest_001",
			orderID:       nil,
			invoiceID:     "",
			invoice:       "lnbc123...",
			amountSatoshis: 1000,
			amountFiat:    10.00,
			exchangeRate:  100000,
			wantError:     true,
		},
		{
			name:          "zero satoshis should fail",
			id:            "pay_003",
			restaurantID:  "rest_001",
			orderID:       nil,
			invoiceID:     "inv_001",
			invoice:       "lnbc123...",
			amountSatoshis: 0,
			amountFiat:    10.00,
			exchangeRate:  100000,
			wantError:     true,
		},
		{
			name:          "zero fiat amount should fail",
			id:            "pay_004",
			restaurantID:  "rest_001",
			orderID:       nil,
			invoiceID:     "inv_001",
			invoice:       "lnbc123...",
			amountSatoshis: 1000,
			amountFiat:    0,
			exchangeRate:  100000,
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment, err := domain.NewPayment(
				tt.id,
				tt.restaurantID,
				tt.orderID,
				tt.invoiceID,
				tt.invoice,
				tt.amountSatoshis,
				tt.amountFiat,
				tt.exchangeRate,
			)
			if (err != nil) != tt.wantError {
				t.Errorf("NewPayment() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && payment == nil {
				t.Error("NewPayment() returned nil payment without error")
			}
			if !tt.wantError && payment.Status != domain.PaymentStatusPending {
				t.Errorf("NewPayment() status = %v, want %v", payment.Status, domain.PaymentStatusPending)
			}
		})
	}
}

func TestPayment_MarkAsPaid(t *testing.T) {
	payment, _ := domain.NewPayment(
		"pay_001",
		"rest_001",
		nil,
		"inv_001",
		"lnbc123...",
		1000,
		10.00,
		100000,
	)

	payment.MarkAsPaid()
	if payment.Status != domain.PaymentStatusPaid {
		t.Errorf("MarkAsPaid() status = %v, want %v", payment.Status, domain.PaymentStatusPaid)
	}
	if payment.PaidAt == nil {
		t.Error("MarkAsPaid() did not set PaidAt")
	}
}

func TestPayment_MarkPaid(t *testing.T) {
	payment, _ := domain.NewPayment(
		"pay_001",
		"rest_001",
		nil,
		"inv_001",
		"lnbc123...",
		1000,
		10.00,
		100000,
	)

	orderID := domain.OrderID("ord_001")
	err := payment.MarkPaid(orderID)
	if err != nil {
		t.Fatalf("MarkPaid() error = %v", err)
	}
	if payment.Status != domain.PaymentStatusPaid {
		t.Errorf("MarkPaid() status = %v, want %v", payment.Status, domain.PaymentStatusPaid)
	}
	if payment.OrderID == nil || *payment.OrderID != orderID {
		t.Error("MarkPaid() did not set OrderID correctly")
	}
}

func TestPayment_MarkPaid_AlreadyPaid(t *testing.T) {
	payment, _ := domain.NewPayment(
		"pay_001",
		"rest_001",
		nil,
		"inv_001",
		"lnbc123...",
		1000,
		10.00,
		100000,
	)

	payment.MarkAsPaid()
	orderID := domain.OrderID("ord_001")
	err := payment.MarkPaid(orderID)
	if err == nil {
		t.Error("MarkPaid() on already paid payment should return error")
	}
}

func TestPayment_MarkAsExpired(t *testing.T) {
	payment, _ := domain.NewPayment(
		"pay_001",
		"rest_001",
		nil,
		"inv_001",
		"lnbc123...",
		1000,
		10.00,
		100000,
	)

	payment.MarkAsExpired("Invoice expired")
	if payment.Status != domain.PaymentStatusExpired {
		t.Errorf("MarkAsExpired() status = %v, want %v", payment.Status, domain.PaymentStatusExpired)
	}
	if payment.FailureReason != "Invoice expired" {
		t.Errorf("MarkAsExpired() FailureReason = %v, want 'Invoice expired'", payment.FailureReason)
	}
}

func TestPayment_MarkFailed(t *testing.T) {
	payment, _ := domain.NewPayment(
		"pay_001",
		"rest_001",
		nil,
		"inv_001",
		"lnbc123...",
		1000,
		10.00,
		100000,
	)

	err := payment.MarkFailed("Payment failed")
	if err != nil {
		t.Fatalf("MarkFailed() error = %v", err)
	}
	if payment.Status != domain.PaymentStatusFailed {
		t.Errorf("MarkFailed() status = %v, want %v", payment.Status, domain.PaymentStatusFailed)
	}
	if payment.FailureReason != "Payment failed" {
		t.Errorf("MarkFailed() FailureReason = %v, want 'Payment failed'", payment.FailureReason)
	}
}

func TestPayment_MarkSettled(t *testing.T) {
	payment, _ := domain.NewPayment(
		"pay_001",
		"rest_001",
		nil,
		"inv_001",
		"lnbc123...",
		1000,
		10.00,
		100000,
	)

	payment.MarkAsPaid()
	err := payment.MarkSettled("tx_123")
	if err != nil {
		t.Fatalf("MarkSettled() error = %v", err)
	}
	if payment.SettlementStatus != domain.SettlementStatusSettled {
		t.Errorf("MarkSettled() SettlementStatus = %v, want %v", payment.SettlementStatus, domain.SettlementStatusSettled)
	}
	if payment.SettlementTransactionID != "tx_123" {
		t.Errorf("MarkSettled() SettlementTransactionID = %v, want 'tx_123'", payment.SettlementTransactionID)
	}
	if payment.SettledAt == nil {
		t.Error("MarkSettled() did not set SettledAt")
	}
}

func TestPayment_MarkSettled_NotPaid(t *testing.T) {
	payment, _ := domain.NewPayment(
		"pay_001",
		"rest_001",
		nil,
		"inv_001",
		"lnbc123...",
		1000,
		10.00,
		100000,
	)

	err := payment.MarkSettled("tx_123")
	if err == nil {
		t.Error("MarkSettled() on unpaid payment should return error")
	}
}

