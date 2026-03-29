package menu

import (
	"context"
	"fmt"

	"bitmerchant/internal/domain"
)

// UpdateMenuCategoryRequest updates category metadata.
type UpdateMenuCategoryRequest struct {
	RestaurantID domain.RestaurantID
	CategoryID   domain.CategoryID
	Name         string
	DisplayOrder int
	IsActive     bool
}

// UpdateMenuCategoryUseCase persists category edits after tenancy checks.
type UpdateMenuCategoryUseCase struct {
	repo domain.MenuCategoryRepository
}

// NewUpdateMenuCategoryUseCase constructs the use case.
func NewUpdateMenuCategoryUseCase(repo domain.MenuCategoryRepository) *UpdateMenuCategoryUseCase {
	return &UpdateMenuCategoryUseCase{repo: repo}
}

// Execute validates and updates the category.
func (uc *UpdateMenuCategoryUseCase) Execute(ctx context.Context, req UpdateMenuCategoryRequest) error {
	cat, err := uc.repo.FindByID(req.CategoryID)
	if err != nil {
		return err
	}
	if cat.RestaurantID != req.RestaurantID {
		return fmt.Errorf("category does not belong to restaurant")
	}
	if err := domain.ValidateCategoryName(req.Name); err != nil {
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
