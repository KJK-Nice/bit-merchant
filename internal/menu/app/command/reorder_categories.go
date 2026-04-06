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
	seen := make(map[common.CategoryID]struct{})
	for _, id := range orderedCategoryIDs {
		if _, dup := seen[id]; dup {
			return fmt.Errorf("duplicate category id in order")
		}
		seen[id] = struct{}{}
	}

	cats, err := uc.catRepo.FindByRestaurantID(restaurantID)
	if err != nil {
		return err
	}
	if len(orderedCategoryIDs) != len(cats) {
		return fmt.Errorf("category count mismatch")
	}
	byID := make(map[common.CategoryID]*menu.MenuCategory)
	for _, c := range cats {
		byID[c.ID] = c
	}
	for _, id := range orderedCategoryIDs {
		if _, ok := byID[id]; !ok {
			return fmt.Errorf("category does not belong to restaurant")
		}
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
