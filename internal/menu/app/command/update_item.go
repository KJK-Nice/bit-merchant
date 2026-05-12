package command

import (
	"context"
	"fmt"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/menu/domain/menu"
	"log/slog"
)

// UpdateMenuItem updates an item and may move it to another category.
//
// Optional / pointer fields use nil to mean "leave unchanged" so legacy callers
// (the dashboard inline modal, integration tests) keep working with the smaller
// payload they already send.
type UpdateMenuItem struct {
	RestaurantID common.RestaurantID
	ItemID       common.ItemID
	CategoryID   common.CategoryID
	Name         string
	Description  string
	Price        float64
	Available    bool
	IsVegetarian bool
	IsGlutenFree bool
	IsSpicy      bool

	// Fields below are optional; nil pointers / nil slices mean
	// "do not change". Empty values explicitly clear the field.
	SpiceLevel               *string
	Schedule                 *string
	SKU                      *string
	IsVegan                  *bool
	IsDairyFree              *bool
	IsHalal                  *bool
	IsNutFree                *bool
	Allergens                *[]string
	Badges                   *[]string
	AllowSpecialInstructions *bool
	OptionGroups             *[]menu.OptionGroup
}

type UpdateMenuItemHandler decorator.CommandHandler[UpdateMenuItem]

type updateMenuItemHandler struct {
	itemRepo menu.ItemRepository
	catRepo  menu.CategoryRepository
}

func NewUpdateMenuItemHandler(itemRepo menu.ItemRepository, catRepo menu.CategoryRepository, log *slog.Logger, metrics decorator.MetricsClient) UpdateMenuItemHandler {
	if itemRepo == nil || catRepo == nil {
		panic("nil menu repository")
	}
	h := updateMenuItemHandler{itemRepo: itemRepo, catRepo: catRepo}
	return decorator.ApplyCommandDecorators[UpdateMenuItem](h, log, metrics)
}

func (h updateMenuItemHandler) Handle(ctx context.Context, cmd UpdateMenuItem) error {
	_ = ctx
	item, cat, err := h.loadItemAndCategory(cmd)
	if err != nil {
		return err
	}
	if err := validateUpdateMenuItem(cmd, item, cat); err != nil {
		return err
	}

	oldCat := item.CategoryID
	item.CategoryID = cmd.CategoryID
	item.Name = cmd.Name
	item.Price = cmd.Price
	item.SetAvailable(cmd.Available)
	if err := item.SetDescription(cmd.Description); err != nil {
		return err
	}
	item.SetDietaryTags(cmd.IsVegetarian, cmd.IsGlutenFree, cmd.IsSpicy)

	if err := applyOptionalFields(item, cmd); err != nil {
		return err
	}

	if oldCat != cmd.CategoryID {
		if err := h.moveItemToCategoryEnd(item, cmd.CategoryID); err != nil {
			return err
		}
	}

	return h.itemRepo.Update(item)
}

func (h updateMenuItemHandler) loadItemAndCategory(cmd UpdateMenuItem) (*menu.MenuItem, *menu.MenuCategory, error) {
	item, err := h.itemRepo.FindByID(cmd.ItemID)
	if err != nil {
		return nil, nil, err
	}
	cat, err := h.catRepo.FindByID(cmd.CategoryID)
	if err != nil {
		return nil, nil, err
	}
	return item, cat, nil
}

func validateUpdateMenuItem(cmd UpdateMenuItem, item *menu.MenuItem, cat *menu.MenuCategory) error {
	if err := validateItemAndCategoryOwnership(item, cat, cmd.RestaurantID); err != nil {
		return err
	}
	if err := menu.ValidateItemName(cmd.Name); err != nil {
		return err
	}
	if err := menu.ValidatePriceForCurrency(cmd.Price, item.Currency); err != nil {
		return err
	}
	return menu.ValidateDescription(cmd.Description)
}

// applyOptionalFields applies the editor-only overlay fields. Nil pointers /
// nil slices leave the existing value untouched so callers that don't know
// about the new fields keep working.
func applyOptionalFields(item *menu.MenuItem, cmd UpdateMenuItem) error {
	if err := applyOptionalValidatedSetters(item, cmd); err != nil {
		return err
	}
	applyOptionalFlags(item, cmd)
	return nil
}

func applyOptionalValidatedSetters(item *menu.MenuItem, cmd UpdateMenuItem) error {
	for _, step := range optionalSetters(cmd) {
		if step.set == nil {
			continue
		}
		if err := step.set(item); err != nil {
			return err
		}
	}
	return nil
}

// optionalSetters returns the slice of per-field setters that need to run when
// the cmd carries a non-nil pointer for that field. A nil function means "skip".
func optionalSetters(cmd UpdateMenuItem) []struct {
	set func(*menu.MenuItem) error
} {
	type step = struct{ set func(*menu.MenuItem) error }
	steps := []step{{}, {}, {}, {}, {}, {}}
	if cmd.SpiceLevel != nil {
		v := *cmd.SpiceLevel
		steps[0].set = func(m *menu.MenuItem) error { return m.SetSpiceLevel(v) }
	}
	if cmd.Schedule != nil {
		v := *cmd.Schedule
		steps[1].set = func(m *menu.MenuItem) error { return m.SetSchedule(v) }
	}
	if cmd.SKU != nil {
		v := *cmd.SKU
		steps[2].set = func(m *menu.MenuItem) error { return m.SetSKU(v) }
	}
	if cmd.Allergens != nil {
		v := *cmd.Allergens
		steps[3].set = func(m *menu.MenuItem) error { return m.SetAllergens(v) }
	}
	if cmd.Badges != nil {
		v := *cmd.Badges
		steps[4].set = func(m *menu.MenuItem) error { return m.SetBadges(v) }
	}
	if cmd.OptionGroups != nil {
		v := *cmd.OptionGroups
		steps[5].set = func(m *menu.MenuItem) error { return m.SetOptionGroups(v) }
	}
	return steps
}

func applyOptionalFlags(item *menu.MenuItem, cmd UpdateMenuItem) {
	if cmd.IsVegan != nil {
		item.IsVegan = *cmd.IsVegan
	}
	if cmd.IsDairyFree != nil {
		item.IsDairyFree = *cmd.IsDairyFree
	}
	if cmd.IsHalal != nil {
		item.IsHalal = *cmd.IsHalal
	}
	if cmd.IsNutFree != nil {
		item.IsNutFree = *cmd.IsNutFree
	}
	if cmd.AllowSpecialInstructions != nil {
		item.SetAllowSpecialInstructions(*cmd.AllowSpecialInstructions)
	}
}

func validateItemAndCategoryOwnership(item *menu.MenuItem, cat *menu.MenuCategory, restaurantID common.RestaurantID) error {
	if item.RestaurantID != restaurantID {
		return fmt.Errorf("item does not belong to restaurant")
	}
	if cat.RestaurantID != restaurantID {
		return fmt.Errorf("category does not belong to restaurant")
	}
	return nil
}

func (h updateMenuItemHandler) moveItemToCategoryEnd(item *menu.MenuItem, categoryID common.CategoryID) error {
	maxOrder, err := h.maxDisplayOrderInCategoryExcluding(categoryID, item.ID)
	if err != nil {
		return err
	}
	return item.SetDisplayOrder(maxOrder + 1)
}

func (h updateMenuItemHandler) maxDisplayOrderInCategoryExcluding(categoryID common.CategoryID, excludeItemID common.ItemID) (int, error) {
	maxOrder := -1
	siblings, err := h.itemRepo.FindByCategoryID(categoryID)
	if err != nil {
		return 0, err
	}
	for _, s := range siblings {
		if s.ID != excludeItemID && s.DisplayOrder > maxOrder {
			maxOrder = s.DisplayOrder
		}
	}
	return maxOrder, nil
}
