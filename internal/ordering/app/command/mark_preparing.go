package command

import (
	"context"
	"errors"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

type MarkOrderPreparingUseCase struct {
	repo     order.Repository
	eventBus common.EventBus
}

func NewMarkOrderPreparingUseCase(repo order.Repository, eventBus common.EventBus) *MarkOrderPreparingUseCase {
	return &MarkOrderPreparingUseCase{repo: repo, eventBus: eventBus}
}

func (uc *MarkOrderPreparingUseCase) Execute(ctx context.Context, orderID common.OrderID) (*order.Order, error) {
	o, err := uc.repo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("order not found")
	}

	if err := o.StartPreparing(); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(o); err != nil {
		return nil, err
	}

	event := order.OrderPreparing{
		OrderID:      o.ID,
		RestaurantID: o.RestaurantID,
		OrderNumber:  o.OrderNumber,
		PreparingAt:  time.Now(),
	}
	if err := uc.eventBus.Publish(ctx, event.EventName(), event); err != nil {
		return nil, err
	}

	return o, nil
}
