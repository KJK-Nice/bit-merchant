package command

import (
	"context"
	"fmt"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/menu/domain/menu"
	"log/slog"
)

// ReorderMenuItems persists item order within a category.
type ReorderMenuItems struct {
	RestaurantID   common.RestaurantID
	CategoryID     common.CategoryID
	OrderedItemIDs []common.ItemID
}

type ReorderMenuItemsHandler decorator.CommandHandler[ReorderMenuItems]

type reorderMenuItemsHandler struct {
	itemRepo menu.ItemRepository
	catRepo  menu.CategoryRepository
}

func NewReorderMenuItemsHandler(itemRepo menu.ItemRepository, catRepo menu.CategoryRepository, log *slog.Logger, metrics decorator.MetricsClient) ReorderMenuItemsHandler {
	if itemRepo == nil || catRepo == nil {
		panic("nil menu repository")
	}
	h := reorderMenuItemsHandler{itemRepo: itemRepo, catRepo: catRepo}
	return decorator.ApplyCommandDecorators[ReorderMenuItems](h, log, metrics)
}

func (h reorderMenuItemsHandler) Handle(ctx context.Context, cmd ReorderMenuItems) error {
	_ = ctx
	cat, err := h.catRepo.FindByID(cmd.CategoryID)
	if err != nil {
		return err
	}
	if cat.RestaurantID != cmd.RestaurantID {
		return fmt.Errorf("category does not belong to restaurant")
	}
	return h.itemRepo.ReorderItemsInCategory(cmd.RestaurantID, cmd.CategoryID, cmd.OrderedItemIDs)
}
