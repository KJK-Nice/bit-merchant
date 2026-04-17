package command

import (
	"context"
	"fmt"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/menu/domain/menu"
	"log/slog"
)

// CreateMenuItem creates a new menu item in a category.
type CreateMenuItem struct {
	RestaurantID common.RestaurantID
	CategoryID   common.CategoryID
	Name         string
	Description  string
	Price        float64
	Available    bool
}

type CreateMenuItemHandler decorator.CommandResultHandler[CreateMenuItem, *menu.MenuItem]

type createMenuItemHandler struct {
	repo menu.ItemRepository
}

func NewCreateMenuItemHandler(repo menu.ItemRepository, log *slog.Logger, metrics decorator.MetricsClient) CreateMenuItemHandler {
	if repo == nil {
		panic("nil menu.ItemRepository")
	}
	h := createMenuItemHandler{repo: repo}
	return decorator.ApplyCommandResultDecorators[CreateMenuItem, *menu.MenuItem](h, log, metrics)
}

func (h createMenuItemHandler) Handle(ctx context.Context, cmd CreateMenuItem) (*menu.MenuItem, error) {
	_ = ctx
	id := common.ItemID(fmt.Sprintf("item_%d", time.Now().UnixNano()))

	item, err := menu.NewMenuItem(id, cmd.CategoryID, cmd.RestaurantID, cmd.Name, cmd.Price)
	if err != nil {
		return nil, err
	}

	if err := item.SetDescription(cmd.Description); err != nil {
		return nil, err
	}
	item.SetAvailable(cmd.Available)

	maxOrder := -1
	siblings, err := h.repo.FindByCategoryID(cmd.CategoryID)
	if err != nil {
		return nil, err
	}
	for _, s := range siblings {
		if s.DisplayOrder > maxOrder {
			maxOrder = s.DisplayOrder
		}
	}
	_ = item.SetDisplayOrder(maxOrder + 1)

	if err := h.repo.Save(item); err != nil {
		return nil, err
	}

	return item, nil
}
