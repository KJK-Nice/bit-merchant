package command

import (
	"context"
	"errors"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

type MarkOrderPaidUseCase struct {
	repo     order.Repository
	eventBus common.EventBus
}

func NewMarkOrderPaidUseCase(repo order.Repository, eventBus common.EventBus) *MarkOrderPaidUseCase {
	return &MarkOrderPaidUseCase{repo: repo, eventBus: eventBus}
}

func (uc *MarkOrderPaidUseCase) Execute(ctx context.Context, orderID common.OrderID) (*order.Order, error) {
	o, err := uc.repo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("order not found")
	}

	o.MarkPaid()

	if err := uc.repo.Update(o); err != nil {
		return nil, err
	}

	event := order.OrderPaid{
		OrderID:      o.ID,
		RestaurantID: o.RestaurantID,
		OrderNumber:  o.OrderNumber,
		TotalAmount:  o.TotalAmount,
		PaidAt:       time.Now(),
	}
	if err := uc.eventBus.Publish(ctx, event.EventName(), event); err != nil {
		return nil, err
	}

	return o, nil
}
