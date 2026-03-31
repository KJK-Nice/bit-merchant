package command

import (
	"context"
	"fmt"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

type CreateRestaurantRequest struct {
	Name string
}

type CreateRestaurantUseCase struct {
	repo restaurant.Repository
}

func NewCreateRestaurantUseCase(repo restaurant.Repository) *CreateRestaurantUseCase {
	return &CreateRestaurantUseCase{repo: repo}
}

func (uc *CreateRestaurantUseCase) Execute(ctx context.Context, req CreateRestaurantRequest) (*restaurant.Restaurant, error) {
	id := common.RestaurantID(fmt.Sprintf("rest_%d", time.Now().UnixNano()))

	r, err := restaurant.NewRestaurant(id, req.Name)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(r); err != nil {
		return nil, err
	}

	return r, nil
}
