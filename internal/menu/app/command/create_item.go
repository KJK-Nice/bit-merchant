package command

import (
	"context"
	"fmt"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

type CreateMenuItemRequest struct {
	RestaurantID common.RestaurantID
	CategoryID   common.CategoryID
	Name         string
	Description  string
	Price        float64
	Available    bool
}

type CreateMenuItemUseCase struct {
	repo menu.ItemRepository
}

func NewCreateMenuItemUseCase(repo menu.ItemRepository) *CreateMenuItemUseCase {
	return &CreateMenuItemUseCase{repo: repo}
}

func (uc *CreateMenuItemUseCase) Execute(ctx context.Context, req CreateMenuItemRequest) (*menu.MenuItem, error) {
	id := common.ItemID(fmt.Sprintf("item_%d", time.Now().UnixNano()))

	item, err := menu.NewMenuItem(id, req.CategoryID, req.RestaurantID, req.Name, req.Price)
	if err != nil {
		return nil, err
	}

	if err := item.SetDescription(req.Description); err != nil {
		return nil, err
	}
	item.SetAvailable(req.Available)

	maxOrder := -1
	siblings, err := uc.repo.FindByCategoryID(req.CategoryID)
	if err != nil {
		return nil, err
	}
	for _, s := range siblings {
		if s.DisplayOrder > maxOrder {
			maxOrder = s.DisplayOrder
		}
	}
	_ = item.SetDisplayOrder(maxOrder + 1)

	if err := uc.repo.Save(item); err != nil {
		return nil, err
	}

	return item, nil
}
