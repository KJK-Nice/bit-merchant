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

	if err := menu.ValidateItemName(req.Name); err != nil {
		return err
	}
	if err := menu.ValidatePrice(req.Price); err != nil {
		return err
	}
	if err := menu.ValidateDescription(req.Description); err != nil {
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
