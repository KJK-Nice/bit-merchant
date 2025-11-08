package memory

import (
	"errors"
	"sync"

	"bitmerchant/internal/domain"
)

// MemoryPaymentRepository implements PaymentRepository with in-memory storage
type MemoryPaymentRepository struct {
	mu       sync.RWMutex
	payments map[domain.PaymentID]*domain.Payment
}

// NewMemoryPaymentRepository creates a new in-memory payment repository
func NewMemoryPaymentRepository() *MemoryPaymentRepository {
	return &MemoryPaymentRepository{
		payments: make(map[domain.PaymentID]*domain.Payment),
	}
}

// Save saves a payment
func (r *MemoryPaymentRepository) Save(payment *domain.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.payments[payment.ID] = payment
	return nil
}

// FindByID finds a payment by ID
func (r *MemoryPaymentRepository) FindByID(id domain.PaymentID) (*domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	payment, exists := r.payments[id]
	if !exists {
		return nil, errors.New("payment not found")
	}
	return payment, nil
}

// FindByInvoiceID finds a payment by invoice ID
func (r *MemoryPaymentRepository) FindByInvoiceID(invoiceID string) (*domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, payment := range r.payments {
		if payment.InvoiceID == invoiceID {
			return payment, nil
		}
	}
	return nil, errors.New("payment not found")
}

// FindByRestaurantID finds all payments for a restaurant
func (r *MemoryPaymentRepository) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Payment
	for _, payment := range r.payments {
		if payment.RestaurantID == restaurantID {
			result = append(result, payment)
		}
	}
	return result, nil
}

// FindPendingSettlements finds pending settlement payments for a restaurant
func (r *MemoryPaymentRepository) FindPendingSettlements(restaurantID domain.RestaurantID) ([]*domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Payment
	for _, payment := range r.payments {
		if payment.RestaurantID == restaurantID &&
			payment.Status == domain.PaymentStatusPaid &&
			payment.SettlementStatus == domain.SettlementStatusPending {
			result = append(result, payment)
		}
	}
	return result, nil
}

// Update updates a payment
func (r *MemoryPaymentRepository) Update(payment *domain.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.payments[payment.ID]; !exists {
		return errors.New("payment not found")
	}
	r.payments[payment.ID] = payment
	return nil
}
