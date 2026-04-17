package command

import (
	"context"
	"fmt"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/menu/domain/menu"
	"log/slog"
)

// ReorderMenuCategories persists a new display order for all categories of a restaurant.
type ReorderMenuCategories struct {
	RestaurantID       common.RestaurantID
	OrderedCategoryIDs []common.CategoryID
}

type ReorderMenuCategoriesHandler decorator.CommandHandler[ReorderMenuCategories]

type reorderMenuCategoriesHandler struct {
	catRepo menu.CategoryRepository
}

func NewReorderMenuCategoriesHandler(catRepo menu.CategoryRepository, log *slog.Logger, metrics decorator.MetricsClient) ReorderMenuCategoriesHandler {
	if catRepo == nil {
		panic("nil menu.CategoryRepository")
	}
	h := reorderMenuCategoriesHandler{catRepo: catRepo}
	return decorator.ApplyCommandDecorators[ReorderMenuCategories](h, log, metrics)
}

func (h reorderMenuCategoriesHandler) Handle(ctx context.Context, cmd ReorderMenuCategories) error {
	_ = ctx
	if len(cmd.OrderedCategoryIDs) == 0 {
		return nil
	}
	if err := validateUniqueCategoryOrder(cmd.OrderedCategoryIDs); err != nil {
		return err
	}

	cats, err := h.catRepo.FindByRestaurantID(cmd.RestaurantID)
	if err != nil {
		return err
	}
	byID, err := validateAndMapCategories(cats, cmd.OrderedCategoryIDs)
	if err != nil {
		return err
	}

	for i, id := range cmd.OrderedCategoryIDs {
		cat := byID[id]
		cat.DisplayOrder = i
		cat.UpdatedAt = time.Now()
		if err := h.catRepo.Update(cat); err != nil {
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
