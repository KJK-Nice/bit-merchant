package command

import (
	"context"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"log/slog"
)

// UpdateRestaurantTableCount sets the configured table count for QR generation.
type UpdateRestaurantTableCount struct {
	RestaurantID common.RestaurantID
	TableCount   int
}

type UpdateRestaurantTableCountHandler decorator.CommandHandler[UpdateRestaurantTableCount]

type updateRestaurantTableCountHandler struct {
	repo restaurant.Repository
}

func NewUpdateRestaurantTableCountHandler(repo restaurant.Repository, log *slog.Logger, metrics decorator.MetricsClient) UpdateRestaurantTableCountHandler {
	if repo == nil {
		panic("nil restaurant.Repository")
	}
	h := updateRestaurantTableCountHandler{repo: repo}
	return decorator.ApplyCommandDecorators[UpdateRestaurantTableCount](h, log, metrics)
}

func (h updateRestaurantTableCountHandler) Handle(ctx context.Context, cmd UpdateRestaurantTableCount) error {
	_ = ctx
	if err := restaurant.ValidateTableCount(cmd.TableCount); err != nil {
		return err
	}
	rest, err := h.repo.FindByID(cmd.RestaurantID)
	if err != nil {
		return err
	}
	rest.TableCount = cmd.TableCount
	rest.UpdatedAt = time.Now()
	return h.repo.Update(rest)
}
