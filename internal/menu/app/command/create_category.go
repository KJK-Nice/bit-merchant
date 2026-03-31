package command

import (
	"context"
	"fmt"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

type CreateMenuCategoryRequest struct {
	RestaurantID common.RestaurantID
	Name         string
	DisplayOrder int
}

type CreateMenuCategoryUseCase struct {
	repo menu.CategoryRepository
}

func NewCreateMenuCategoryUseCase(repo menu.CategoryRepository) *CreateMenuCategoryUseCase {
	return &CreateMenuCategoryUseCase{repo: repo}
}

func (uc *CreateMenuCategoryUseCase) Execute(ctx context.Context, req CreateMenuCategoryRequest) (*menu.MenuCategory, error) {
	id := common.CategoryID(fmt.Sprintf("cat_%d", time.Now().UnixNano()))

	category, err := menu.NewMenuCategory(id, req.RestaurantID, req.Name, req.DisplayOrder)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(category); err != nil {
		return nil, err
	}

	return category, nil
}
