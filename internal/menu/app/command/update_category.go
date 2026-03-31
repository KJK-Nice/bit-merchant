package command

import (
	"context"
	"fmt"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

type UpdateMenuCategoryRequest struct {
	RestaurantID common.RestaurantID
	CategoryID   common.CategoryID
	Name         string
	DisplayOrder int
	IsActive     bool
}

type UpdateMenuCategoryUseCase struct {
	repo menu.CategoryRepository
}

func NewUpdateMenuCategoryUseCase(repo menu.CategoryRepository) *UpdateMenuCategoryUseCase {
	return &UpdateMenuCategoryUseCase{repo: repo}
}

func (uc *UpdateMenuCategoryUseCase) Execute(ctx context.Context, req UpdateMenuCategoryRequest) error {
	cat, err := uc.repo.FindByID(req.CategoryID)
	if err != nil {
		return err
	}
	if cat.RestaurantID != req.RestaurantID {
		return fmt.Errorf("category does not belong to restaurant")
	}
	if err := menu.ValidateCategoryName(req.Name); err != nil {
		return err
	}
	if req.DisplayOrder < 0 {
		return fmt.Errorf("display order must be >= 0")
	}

	cat.Name = req.Name
	cat.DisplayOrder = req.DisplayOrder
	cat.SetActive(req.IsActive)

	return uc.repo.Update(cat)
}
