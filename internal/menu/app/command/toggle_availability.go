package command

import (
	"context"
	"fmt"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

type ToggleMenuItemAvailabilityUseCase struct {
	repo menu.ItemRepository
}

func NewToggleMenuItemAvailabilityUseCase(repo menu.ItemRepository) *ToggleMenuItemAvailabilityUseCase {
	return &ToggleMenuItemAvailabilityUseCase{repo: repo}
}

func (uc *ToggleMenuItemAvailabilityUseCase) Execute(ctx context.Context, restaurantID common.RestaurantID, itemID common.ItemID) error {
	item, err := uc.repo.FindByID(itemID)
	if err != nil {
		return err
	}
	if item.RestaurantID != restaurantID {
		return fmt.Errorf("item does not belong to restaurant")
	}
	item.SetAvailable(!item.IsAvailable)
	return uc.repo.Update(item)
}
