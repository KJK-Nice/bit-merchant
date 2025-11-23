package kitchen

import (
	"bitmerchant/internal/domain"
	"context"
	"errors"
	"time"
)

type MarkOrderPreparingUseCase struct {
	repo     domain.OrderRepository
	eventBus domain.EventBus
}

func NewMarkOrderPreparingUseCase(repo domain.OrderRepository, eventBus domain.EventBus) *MarkOrderPreparingUseCase {
	return &MarkOrderPreparingUseCase{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (uc *MarkOrderPreparingUseCase) Execute(ctx context.Context, orderID domain.OrderID) (*domain.Order, error) {
	order, err := uc.repo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	if order.PaymentStatus != domain.PaymentStatusPaid {
		return nil, errors.New("cannot prepare unpaid order")
	}

	if err := order.UpdateFulfillmentStatus(domain.FulfillmentStatusPreparing); err != nil {
		return nil, err
	}
	
	now := time.Now()
	order.PreparingAt = &now

	if err := uc.repo.Update(order); err != nil {
		return nil, err
	}

	event := domain.OrderPreparing{
		OrderID:      order.ID,
		RestaurantID: order.RestaurantID,
		OrderNumber:  order.OrderNumber,
		PreparingAt:  now,
	}

	if err := uc.eventBus.Publish(ctx, event.EventName(), event); err != nil {
		return nil, err
	}

	return order, nil
}
