package restaurant

import (
	"context"
	"fmt"
	"strings"

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

// Execute flips restaurant availability: when closing, applies closedMessage and reopeningHours;
// when opening, clears those fields. Both optional strings are trimmed; each must pass ValidateDescription (<=500 chars).
func (uc *ToggleRestaurantOpenUseCase) Execute(ctx context.Context, id domain.RestaurantID, closedMessage, reopeningHours string) (bool, error) {
	restaurant, err := uc.repo.FindByID(id)
	if err != nil {
		return false, err
	}

	closedMessage = strings.TrimSpace(closedMessage)
	reopeningHours = strings.TrimSpace(reopeningHours)
	if err := domain.ValidateDescription(closedMessage); err != nil {
		return false, fmt.Errorf("closed message: %w", err)
	}
	if err := domain.ValidateDescription(reopeningHours); err != nil {
		return false, fmt.Errorf("reopening hours: %w", err)
	}

	if restaurant.IsOpen {
		restaurant.UpdateStatus(false, closedMessage, reopeningHours)
	} else {
		restaurant.UpdateStatus(true, "", "")
	}

	if err := uc.repo.Save(restaurant); err != nil {
		return false, err
	}

	return restaurant.IsOpen, nil
}
