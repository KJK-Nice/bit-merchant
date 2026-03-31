package command

import (
	"context"
	"fmt"
	"strings"

	"bitmerchant/internal/common"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

type ToggleRestaurantOpenUseCase struct {
	repo restaurant.Repository
}

func NewToggleRestaurantOpenUseCase(repo restaurant.Repository) *ToggleRestaurantOpenUseCase {
	return &ToggleRestaurantOpenUseCase{repo: repo}
}

func (uc *ToggleRestaurantOpenUseCase) Execute(ctx context.Context, id common.RestaurantID, closedMessage, reopeningHours string) (bool, error) {
	r, err := uc.repo.FindByID(id)
	if err != nil {
		return false, err
	}

	closedMessage = strings.TrimSpace(closedMessage)
	reopeningHours = strings.TrimSpace(reopeningHours)
	if err := restaurant.ValidateDescription(closedMessage); err != nil {
		return false, fmt.Errorf("closed message: %w", err)
	}
	if err := restaurant.ValidateDescription(reopeningHours); err != nil {
		return false, fmt.Errorf("reopening hours: %w", err)
	}

	if r.IsOpen {
		r.Close(closedMessage, reopeningHours)
	} else {
		r.Open()
	}

	if err := uc.repo.Save(r); err != nil {
		return false, err
	}

	return r.IsOpen, nil
}
