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

// CreateMenuCategory creates a new menu category.
type CreateMenuCategory struct {
	RestaurantID common.RestaurantID
	Name         string
	DisplayOrder int
}

type CreateMenuCategoryHandler decorator.CommandResultHandler[CreateMenuCategory, *menu.MenuCategory]

type createMenuCategoryHandler struct {
	repo menu.CategoryRepository
}

func NewCreateMenuCategoryHandler(repo menu.CategoryRepository, log *slog.Logger, metrics decorator.MetricsClient) CreateMenuCategoryHandler {
	if repo == nil {
		panic("nil menu.CategoryRepository")
	}
	h := createMenuCategoryHandler{repo: repo}
	return decorator.ApplyCommandResultDecorators[CreateMenuCategory, *menu.MenuCategory](h, log, metrics)
}

func (h createMenuCategoryHandler) Handle(ctx context.Context, cmd CreateMenuCategory) (*menu.MenuCategory, error) {
	_ = ctx
	id := common.CategoryID(fmt.Sprintf("cat_%d", time.Now().UnixNano()))

	category, err := menu.NewMenuCategory(id, cmd.RestaurantID, cmd.Name, cmd.DisplayOrder)
	if err != nil {
		return nil, err
	}

	if err := h.repo.Save(category); err != nil {
		return nil, err
	}

	return category, nil
}
