package command

import (
	"context"
	"fmt"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/menu/domain/menu"
	"log/slog"
)

// UpdateMenuItem updates an item and may move it to another category.
type UpdateMenuItem struct {
	RestaurantID common.RestaurantID
	ItemID       common.ItemID
	CategoryID   common.CategoryID
	Name         string
	Description  string
	Price        float64
	Available    bool
}

type UpdateMenuItemHandler decorator.CommandHandler[UpdateMenuItem]

type updateMenuItemHandler struct {
	itemRepo menu.ItemRepository
	catRepo  menu.CategoryRepository
}

func NewUpdateMenuItemHandler(itemRepo menu.ItemRepository, catRepo menu.CategoryRepository, log *slog.Logger, metrics decorator.MetricsClient) UpdateMenuItemHandler {
	if itemRepo == nil || catRepo == nil {
		panic("nil menu repository")
	}
	h := updateMenuItemHandler{itemRepo: itemRepo, catRepo: catRepo}
	return decorator.ApplyCommandDecorators[UpdateMenuItem](h, log, metrics)
}

func (h updateMenuItemHandler) Handle(ctx context.Context, cmd UpdateMenuItem) error {
	_ = ctx
	item, err := h.itemRepo.FindByID(cmd.ItemID)
	if err != nil {
		return err
	}

	cat, err := h.catRepo.FindByID(cmd.CategoryID)
	if err != nil {
		return err
	}

	if err := validateItemAndCategoryOwnership(item, cat, cmd.RestaurantID); err != nil {
		return err
	}
	if err := menu.ValidateItemName(cmd.Name); err != nil {
		return err
	}
	if err := menu.ValidatePriceForCurrency(cmd.Price, item.Currency); err != nil {
		return err
	}
	if err := menu.ValidateDescription(cmd.Description); err != nil {
		return err
	}

	oldCat := item.CategoryID
	item.CategoryID = cmd.CategoryID
	item.Name = cmd.Name
	item.Price = cmd.Price
	item.SetAvailable(cmd.Available)
	if err := item.SetDescription(cmd.Description); err != nil {
		return err
	}

	if oldCat != cmd.CategoryID {
		if err := h.moveItemToCategoryEnd(item, cmd.CategoryID); err != nil {
			return err
		}
	}

	return h.itemRepo.Update(item)
}

func validateItemAndCategoryOwnership(item *menu.MenuItem, cat *menu.MenuCategory, restaurantID common.RestaurantID) error {
	if item.RestaurantID != restaurantID {
		return fmt.Errorf("item does not belong to restaurant")
	}
	if cat.RestaurantID != restaurantID {
		return fmt.Errorf("category does not belong to restaurant")
	}
	return nil
}

func (h updateMenuItemHandler) moveItemToCategoryEnd(item *menu.MenuItem, categoryID common.CategoryID) error {
	maxOrder, err := h.maxDisplayOrderInCategoryExcluding(categoryID, item.ID)
	if err != nil {
		return err
	}
	return item.SetDisplayOrder(maxOrder + 1)
}

func (h updateMenuItemHandler) maxDisplayOrderInCategoryExcluding(categoryID common.CategoryID, excludeItemID common.ItemID) (int, error) {
	maxOrder := -1
	siblings, err := h.itemRepo.FindByCategoryID(categoryID)
	if err != nil {
		return 0, err
	}
	for _, s := range siblings {
		if s.ID != excludeItemID && s.DisplayOrder > maxOrder {
			maxOrder = s.DisplayOrder
		}
	}
	return maxOrder, nil
}
