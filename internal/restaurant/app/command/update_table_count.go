package command

import (
	"context"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

type UpdateRestaurantTableCountUseCase struct {
	repo restaurant.Repository
}

func NewUpdateRestaurantTableCountUseCase(repo restaurant.Repository) *UpdateRestaurantTableCountUseCase {
	return &UpdateRestaurantTableCountUseCase{repo: repo}
}

func (uc *UpdateRestaurantTableCountUseCase) Execute(ctx context.Context, restaurantID common.RestaurantID, tableCount int) error {
	if err := restaurant.ValidateTableCount(tableCount); err != nil {
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
