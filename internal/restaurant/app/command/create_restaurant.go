package command

import (
	"context"
	"fmt"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"log/slog"
)

// CreateRestaurant registers a new restaurant (command payload).
type CreateRestaurant struct {
	Name string
}

type CreateRestaurantHandler decorator.CommandResultHandler[CreateRestaurant, *restaurant.Restaurant]

type createRestaurantHandler struct {
	repo restaurant.Repository
}

func NewCreateRestaurantHandler(repo restaurant.Repository, log *slog.Logger, metrics decorator.MetricsClient) CreateRestaurantHandler {
	if repo == nil {
		panic("nil restaurant.Repository")
	}
	h := createRestaurantHandler{repo: repo}
	return decorator.ApplyCommandResultDecorators[CreateRestaurant, *restaurant.Restaurant](h, log, metrics)
}

func (h createRestaurantHandler) Handle(ctx context.Context, cmd CreateRestaurant) (*restaurant.Restaurant, error) {
	_ = ctx
	id := common.RestaurantID(fmt.Sprintf("rest_%d", time.Now().UnixNano()))

	r, err := restaurant.NewRestaurant(id, cmd.Name)
	if err != nil {
		return nil, err
	}

	if err := h.repo.Save(r); err != nil {
		return nil, err
	}

	return r, nil
}
