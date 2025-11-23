package restaurant

import (
	"bitmerchant/internal/domain"
	"context"
	"fmt"
	"time"
)

type CreateRestaurantRequest struct {
	Name string
}

type CreateRestaurantUseCase struct {
	repo domain.RestaurantRepository
}

func NewCreateRestaurantUseCase(repo domain.RestaurantRepository) *CreateRestaurantUseCase {
	return &CreateRestaurantUseCase{repo: repo}
}

func (uc *CreateRestaurantUseCase) Execute(ctx context.Context, req CreateRestaurantRequest) (*domain.Restaurant, error) {
	// Generate ID
	id := domain.RestaurantID(fmt.Sprintf("rest_%d", time.Now().UnixNano()))
	
	restaurant, err := domain.NewRestaurant(id, req.Name)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(restaurant); err != nil {
		return nil, err
	}

	return restaurant, nil
}

