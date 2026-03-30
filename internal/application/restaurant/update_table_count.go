package restaurant

import (
	"context"
	"time"

	"bitmerchant/internal/domain"
)

// UpdateRestaurantTableCountUseCase persists the configured number of dining tables for QR codes.
type UpdateRestaurantTableCountUseCase struct {
	repo domain.RestaurantRepository
}

// NewUpdateRestaurantTableCountUseCase constructs the use case.
func NewUpdateRestaurantTableCountUseCase(repo domain.RestaurantRepository) *UpdateRestaurantTableCountUseCase {
	return &UpdateRestaurantTableCountUseCase{repo: repo}
}

// Execute validates count and updates the restaurant.
func (uc *UpdateRestaurantTableCountUseCase) Execute(ctx context.Context, restaurantID domain.RestaurantID, tableCount int) error {
	if err := domain.ValidateTableCount(tableCount); err != nil {
		return err
	}
	rest, err := uc.repo.FindByID(restaurantID)
	if err != nil {
		return err
	}
	rest.TableCount = tableCount
	rest.UpdatedAt = time.Now()
	return uc.repo.Update(rest)
}
