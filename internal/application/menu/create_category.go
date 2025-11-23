package menu

import (
	"bitmerchant/internal/domain"
	"context"
	"fmt"
	"time"
)

type CreateMenuCategoryRequest struct {
	RestaurantID domain.RestaurantID
	Name         string
	DisplayOrder int
}

type CreateMenuCategoryUseCase struct {
	repo domain.MenuCategoryRepository
}

func NewCreateMenuCategoryUseCase(repo domain.MenuCategoryRepository) *CreateMenuCategoryUseCase {
	return &CreateMenuCategoryUseCase{repo: repo}
}

func (uc *CreateMenuCategoryUseCase) Execute(ctx context.Context, req CreateMenuCategoryRequest) (*domain.MenuCategory, error) {
	id := domain.CategoryID(fmt.Sprintf("cat_%d", time.Now().UnixNano()))
	
	category, err := domain.NewMenuCategory(id, req.RestaurantID, req.Name, req.DisplayOrder)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(category); err != nil {
		return nil, err
	}

	return category, nil
}
