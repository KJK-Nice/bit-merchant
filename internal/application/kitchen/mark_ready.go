package kitchen

import (
	"bitmerchant/internal/domain"
	"context"
	"errors"
	"time"
)

type MarkOrderReadyUseCase struct {
	repo     domain.OrderRepository
	eventBus domain.EventBus
}

func NewMarkOrderReadyUseCase(repo domain.OrderRepository, eventBus domain.EventBus) *MarkOrderReadyUseCase {
	return &MarkOrderReadyUseCase{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (uc *MarkOrderReadyUseCase) Execute(ctx context.Context, orderID domain.OrderID) (*domain.Order, error) {
	order, err := uc.repo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	if err := order.UpdateFulfillmentStatus(domain.FulfillmentStatusReady); err != nil {
		return nil, err
	}

	now := time.Now()
	order.ReadyAt = &now

	if err := uc.repo.Update(order); err != nil {
		return nil, err
	}

	event := domain.OrderReady{
		OrderID:      order.ID,
		RestaurantID: order.RestaurantID,
		OrderNumber:  order.OrderNumber,
		ReadyAt:      now,
	}

	if err := uc.eventBus.Publish(ctx, event.EventName(), event); err != nil {
		return nil, err
	}
	
	return order, nil
}
