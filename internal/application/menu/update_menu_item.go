package menu

import (
	"context"
	"fmt"

	"bitmerchant/internal/domain"
)

// UpdateMenuItemRequest updates fields on an existing item.
type UpdateMenuItemRequest struct {
	RestaurantID domain.RestaurantID
	ItemID       domain.ItemID
	CategoryID   domain.CategoryID
	Name         string
	Description  string
	Price        float64
	Available    bool
}

// UpdateMenuItemUseCase persists item edits after tenancy checks.
type UpdateMenuItemUseCase struct {
	itemRepo domain.MenuItemRepository
	catRepo  domain.MenuCategoryRepository
}

// NewUpdateMenuItemUseCase constructs the use case.
func NewUpdateMenuItemUseCase(itemRepo domain.MenuItemRepository, catRepo domain.MenuCategoryRepository) *UpdateMenuItemUseCase {
	return &UpdateMenuItemUseCase{itemRepo: itemRepo, catRepo: catRepo}
}

// Execute validates and updates the menu item.
func (uc *UpdateMenuItemUseCase) Execute(ctx context.Context, req UpdateMenuItemRequest) error {
	item, err := uc.itemRepo.FindByID(req.ItemID)
	if err != nil {
		return err
	}
	if item.RestaurantID != req.RestaurantID {
		return fmt.Errorf("item does not belong to restaurant")
	}

	cat, err := uc.catRepo.FindByID(req.CategoryID)
	if err != nil {
		return err
	}
	if cat.RestaurantID != req.RestaurantID {
		return fmt.Errorf("category does not belong to restaurant")
	}

	if err := domain.ValidateItemName(req.Name); err != nil {
		return err
	}
	if err := domain.ValidatePrice(req.Price); err != nil {
		return err
	}
	if err := domain.ValidateDescription(req.Description); err != nil {
		return err
	}

	item.CategoryID = req.CategoryID
	item.Name = req.Name
	item.Price = req.Price
	item.SetAvailable(req.Available)
	if err := item.SetDescription(req.Description); err != nil {
		return err
	}

	return uc.itemRepo.Update(item)
}
