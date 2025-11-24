package restaurant

import (
	"context"

	"bitmerchant/internal/domain"
)

type ToggleRestaurantOpenUseCase struct {
	repo domain.RestaurantRepository
}

func NewToggleRestaurantOpenUseCase(repo domain.RestaurantRepository) *ToggleRestaurantOpenUseCase {
	return &ToggleRestaurantOpenUseCase{
		repo: repo,
	}
}

func (uc *ToggleRestaurantOpenUseCase) Execute(ctx context.Context, id domain.RestaurantID) (bool, error) {
	restaurant, err := uc.repo.FindByID(id)
	if err != nil {
		return false, err
	}

	restaurant.IsOpen = !restaurant.IsOpen
	
	if err := uc.repo.Save(restaurant); err != nil {
		return false, err
	}

	return restaurant.IsOpen, nil
}

