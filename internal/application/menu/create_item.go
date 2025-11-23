package menu

import (
	"bitmerchant/internal/domain"
	"context"
	"fmt"
	"time"
)

type CreateMenuItemRequest struct {
	RestaurantID domain.RestaurantID
	CategoryID   domain.CategoryID
	Name         string
	Description  string
	Price        float64
	Available    bool
}

type CreateMenuItemUseCase struct {
	repo domain.MenuItemRepository
}

func NewCreateMenuItemUseCase(repo domain.MenuItemRepository) *CreateMenuItemUseCase {
	return &CreateMenuItemUseCase{repo: repo}
}

func (uc *CreateMenuItemUseCase) Execute(ctx context.Context, req CreateMenuItemRequest) (*domain.MenuItem, error) {
	id := domain.ItemID(fmt.Sprintf("item_%d", time.Now().UnixNano()))
	
	item, err := domain.NewMenuItem(id, req.CategoryID, req.RestaurantID, req.Name, req.Price)
	if err != nil {
		return nil, err
	}
	
	item.SetDescription(req.Description)
	item.SetAvailable(req.Available)

	if err := uc.repo.Save(item); err != nil {
		return nil, err
	}

	return item, nil
}
