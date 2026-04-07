package command

import (
	"context"
	"fmt"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

type ReorderMenuItemsUseCase struct {
	itemRepo menu.ItemRepository
	catRepo  menu.CategoryRepository
}

func NewReorderMenuItemsUseCase(itemRepo menu.ItemRepository, catRepo menu.CategoryRepository) *ReorderMenuItemsUseCase {
	return &ReorderMenuItemsUseCase{itemRepo: itemRepo, catRepo: catRepo}
}

func (uc *ReorderMenuItemsUseCase) Execute(ctx context.Context, restaurantID common.RestaurantID, categoryID common.CategoryID, orderedItemIDs []common.ItemID) error {
	cat, err := uc.catRepo.FindByID(categoryID)
	if err != nil {
		return err
	}
	if cat.RestaurantID != restaurantID {
		return fmt.Errorf("category does not belong to restaurant")
	}
	return uc.itemRepo.ReorderItemsInCategory(restaurantID, categoryID, orderedItemIDs)
}
