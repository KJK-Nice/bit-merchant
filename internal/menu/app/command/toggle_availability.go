package command

import (
	"context"
	"fmt"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/menu/domain/menu"
	"log/slog"
)

// ToggleMenuItemAvailability flips an item's availability flag.
type ToggleMenuItemAvailability struct {
	RestaurantID common.RestaurantID
	ItemID       common.ItemID
}

type ToggleMenuItemAvailabilityHandler decorator.CommandHandler[ToggleMenuItemAvailability]

type toggleMenuItemAvailabilityHandler struct {
	repo menu.ItemRepository
}

func NewToggleMenuItemAvailabilityHandler(repo menu.ItemRepository, log *slog.Logger, metrics decorator.MetricsClient) ToggleMenuItemAvailabilityHandler {
	if repo == nil {
		panic("nil menu.ItemRepository")
	}
	h := toggleMenuItemAvailabilityHandler{repo: repo}
	return decorator.ApplyCommandDecorators[ToggleMenuItemAvailability](h, log, metrics)
}

func (h toggleMenuItemAvailabilityHandler) Handle(ctx context.Context, cmd ToggleMenuItemAvailability) error {
	_ = ctx
	item, err := h.repo.FindByID(cmd.ItemID)
	if err != nil {
		return err
	}
	if item.RestaurantID != cmd.RestaurantID {
		return fmt.Errorf("item does not belong to restaurant")
	}
	item.SetAvailable(!item.IsAvailable)
	return h.repo.Update(item)
}
