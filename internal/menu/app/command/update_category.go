package command

import (
	"context"
	"fmt"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/menu/domain/menu"
	"log/slog"
)

// UpdateMenuCategory updates category metadata and active flag.
type UpdateMenuCategory struct {
	RestaurantID common.RestaurantID
	CategoryID   common.CategoryID
	Name         string
	DisplayOrder int
	IsActive     bool
}

type UpdateMenuCategoryHandler decorator.CommandHandler[UpdateMenuCategory]

type updateMenuCategoryHandler struct {
	repo menu.CategoryRepository
}

func NewUpdateMenuCategoryHandler(repo menu.CategoryRepository, log *slog.Logger, metrics decorator.MetricsClient) UpdateMenuCategoryHandler {
	if repo == nil {
		panic("nil menu.CategoryRepository")
	}
	h := updateMenuCategoryHandler{repo: repo}
	return decorator.ApplyCommandDecorators[UpdateMenuCategory](h, log, metrics)
}

func (h updateMenuCategoryHandler) Handle(ctx context.Context, cmd UpdateMenuCategory) error {
	_ = ctx
	cat, err := h.repo.FindByID(cmd.CategoryID)
	if err != nil {
		return err
	}
	if cat.RestaurantID != cmd.RestaurantID {
		return fmt.Errorf("category does not belong to restaurant")
	}
	if err := menu.ValidateCategoryName(cmd.Name); err != nil {
		return err
	}
	if cmd.DisplayOrder < 0 {
		return fmt.Errorf("display order must be >= 0")
	}

	cat.Name = cmd.Name
	cat.DisplayOrder = cmd.DisplayOrder
	cat.SetActive(cmd.IsActive)

	return h.repo.Update(cat)
}
