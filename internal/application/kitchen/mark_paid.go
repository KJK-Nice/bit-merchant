package kitchen

import (
	"bitmerchant/internal/domain"
	"context"
	"errors"
	"time"
)

type MarkOrderPaidUseCase struct {
	repo     domain.OrderRepository
	eventBus domain.EventBus
}

func NewMarkOrderPaidUseCase(repo domain.OrderRepository, eventBus domain.EventBus) *MarkOrderPaidUseCase {
	return &MarkOrderPaidUseCase{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (uc *MarkOrderPaidUseCase) Execute(ctx context.Context, orderID domain.OrderID) (*domain.Order, error) {
	order, err := uc.repo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	// Update status
	order.PaymentStatus = domain.PaymentStatusPaid
	now := time.Now()
	order.PaidAt = &now
	order.UpdatedAt = now

	if err := uc.repo.Update(order); err != nil {
		return nil, err
	}

	// Publish event
	event := domain.OrderPaid{
		OrderID:      order.ID,
		RestaurantID: order.RestaurantID,
		OrderNumber:  order.OrderNumber,
		TotalAmount:  order.TotalAmount,
		PaidAt:       now,
	}

	if err := uc.eventBus.Publish(ctx, event.EventName(), event); err != nil {
		return nil, err
	}

	return order, nil
}
