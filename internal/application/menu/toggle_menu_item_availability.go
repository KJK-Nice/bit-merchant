package menu

import (
	"context"
	"fmt"

	"bitmerchant/internal/domain"
)

// ToggleMenuItemAvailabilityUseCase flips IsAvailable for quick admin actions.
type ToggleMenuItemAvailabilityUseCase struct {
	repo domain.MenuItemRepository
}

// NewToggleMenuItemAvailabilityUseCase constructs the use case.
func NewToggleMenuItemAvailabilityUseCase(repo domain.MenuItemRepository) *ToggleMenuItemAvailabilityUseCase {
	return &ToggleMenuItemAvailabilityUseCase{repo: repo}
}

// Execute toggles availability if the item belongs to the restaurant.
func (uc *ToggleMenuItemAvailabilityUseCase) Execute(ctx context.Context, restaurantID domain.RestaurantID, itemID domain.ItemID) error {
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
