package adapters

import (
	"errors"
	"sync"

	"bitmerchant/internal/common"
	"bitmerchant/internal/payment/domain/payment"
)

type MemoryPaymentRepository struct {
	mu       sync.RWMutex
	payments map[common.PaymentID]*payment.Payment
}

func NewMemoryPaymentRepository() *MemoryPaymentRepository {
	return &MemoryPaymentRepository{
		payments: make(map[common.PaymentID]*payment.Payment),
	}
}

func (r *MemoryPaymentRepository) Save(p *payment.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.payments[p.ID] = p
	return nil
}

func (r *MemoryPaymentRepository) FindByID(id common.PaymentID) (*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, exists := r.payments[id]
	if !exists {
		return nil, errors.New("payment not found")
	}
	return p, nil
}

func (r *MemoryPaymentRepository) FindByOrderID(orderID common.OrderID) (*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.payments {
		if p.OrderID == orderID {
			return p, nil
		}
	}
	return nil, errors.New("payment not found")
}

func (r *MemoryPaymentRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*payment.Payment
	for _, p := range r.payments {
		if p.RestaurantID == restaurantID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *MemoryPaymentRepository) Update(p *payment.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.payments[p.ID]; !exists {
		return errors.New("payment not found")
	}
	r.payments[p.ID] = p
	return nil
}
