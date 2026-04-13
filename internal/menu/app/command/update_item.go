package command

import (
	"context"
	"fmt"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

type UpdateMenuItemRequest struct {
	RestaurantID common.RestaurantID
	ItemID       common.ItemID
	CategoryID   common.CategoryID
	Name         string
	Description  string
	Price        float64
	Available    bool
}

type UpdateMenuItemUseCase struct {
	itemRepo menu.ItemRepository
	catRepo  menu.CategoryRepository
}

func NewUpdateMenuItemUseCase(itemRepo menu.ItemRepository, catRepo menu.CategoryRepository) *UpdateMenuItemUseCase {
	return &UpdateMenuItemUseCase{itemRepo: itemRepo, catRepo: catRepo}
}

func (uc *UpdateMenuItemUseCase) Execute(ctx context.Context, req UpdateMenuItemRequest) error {
	item, err := uc.itemRepo.FindByID(req.ItemID)
	if err != nil {
		return err
	}

	cat, err := uc.catRepo.FindByID(req.CategoryID)
	if err != nil {
		return err
	}

	if err := validateItemAndCategoryOwnership(item, cat, req.RestaurantID); err != nil {
		return err
	}
	if err := validateUpdateMenuItemRequest(req); err != nil {
		return err
	}

	oldCat := item.CategoryID
	item.CategoryID = req.CategoryID
	item.Name = req.Name
	item.Price = req.Price
	item.SetAvailable(req.Available)
	if err := item.SetDescription(req.Description); err != nil {
		return err
	}

	if oldCat != req.CategoryID {
		if err := uc.moveItemToCategoryEnd(item, req.CategoryID); err != nil {
			return err
		}
	}

	return uc.itemRepo.Update(item)
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

func validateUpdateMenuItemRequest(req UpdateMenuItemRequest) error {
	if err := menu.ValidateItemName(req.Name); err != nil {
		return err
	}
	if err := menu.ValidatePrice(req.Price); err != nil {
		return err
	}
	if err := menu.ValidateDescription(req.Description); err != nil {
		return err
	}
	return nil
}

func (uc *UpdateMenuItemUseCase) moveItemToCategoryEnd(item *menu.MenuItem, categoryID common.CategoryID) error {
	maxOrder, err := uc.maxDisplayOrderInCategoryExcluding(categoryID, item.ID)
	if err != nil {
		return err
	}
	return item.SetDisplayOrder(maxOrder + 1)
}

func (uc *UpdateMenuItemUseCase) maxDisplayOrderInCategoryExcluding(categoryID common.CategoryID, excludeItemID common.ItemID) (int, error) {
	maxOrder := -1
	siblings, err := uc.itemRepo.FindByCategoryID(categoryID)
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
