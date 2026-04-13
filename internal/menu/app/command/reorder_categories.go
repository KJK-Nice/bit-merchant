package command

import (
	"context"
	"fmt"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

type ReorderMenuCategoriesUseCase struct {
	catRepo menu.CategoryRepository
}

func NewReorderMenuCategoriesUseCase(catRepo menu.CategoryRepository) *ReorderMenuCategoriesUseCase {
	return &ReorderMenuCategoriesUseCase{catRepo: catRepo}
}

func (uc *ReorderMenuCategoriesUseCase) Execute(ctx context.Context, restaurantID common.RestaurantID, orderedCategoryIDs []common.CategoryID) error {
	if len(orderedCategoryIDs) == 0 {
		return nil
	}
	if err := validateUniqueCategoryOrder(orderedCategoryIDs); err != nil {
		return err
	}

	cats, err := uc.catRepo.FindByRestaurantID(restaurantID)
	if err != nil {
		return err
	}
	byID, err := validateAndMapCategories(cats, orderedCategoryIDs)
	if err != nil {
		return err
	}

	for i, id := range orderedCategoryIDs {
		cat := byID[id]
		cat.DisplayOrder = i
		cat.UpdatedAt = time.Now()
		if err := uc.catRepo.Update(cat); err != nil {
			return err
		}
	}
	return nil
}

func validateUniqueCategoryOrder(orderedCategoryIDs []common.CategoryID) error {
	seen := make(map[common.CategoryID]struct{}, len(orderedCategoryIDs))
	for _, id := range orderedCategoryIDs {
		if _, dup := seen[id]; dup {
			return fmt.Errorf("duplicate category id in order")
		}
		seen[id] = struct{}{}
	}
	return nil
}

func validateAndMapCategories(cats []*menu.MenuCategory, orderedCategoryIDs []common.CategoryID) (map[common.CategoryID]*menu.MenuCategory, error) {
	if len(orderedCategoryIDs) != len(cats) {
		return nil, fmt.Errorf("category count mismatch")
	}

	byID := make(map[common.CategoryID]*menu.MenuCategory, len(cats))
	for _, c := range cats {
		byID[c.ID] = c
	}
	for _, id := range orderedCategoryIDs {
		if _, ok := byID[id]; !ok {
			return nil, fmt.Errorf("category does not belong to restaurant")
		}
	}
	return byID, nil
}
