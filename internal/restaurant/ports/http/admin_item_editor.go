package http

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/interfaces/templates/admin"
	menuCmd "bitmerchant/internal/menu/app/command"
	"bitmerchant/internal/menu/app/dto"
	menuQuery "bitmerchant/internal/menu/app/query"
	"bitmerchant/internal/menu/domain/menu"

	"github.com/labstack/echo/v4"
)

const (
	adminFlashItemSaved      = "item_saved"
	adminFlashItemSaveFailed = "item_save_failed"
)

// editorItemPath returns the canonical editor URL for the given item, with an
// optional flash code.
func editorItemPath(itemID common.ItemID, flashCode string) string {
	path := "/admin/items/" + url.PathEscape(string(itemID)) + "/edit"
	if flashCode != "" {
		path += "?flash=" + url.QueryEscape(flashCode)
	}
	return path
}

// itemEditorFlash returns the user-facing message for a known flash code.
func itemEditorFlash(flashCode string) (msg string, success bool) {
	switch flashCode {
	case adminFlashItemSaved:
		return "Item saved.", true
	case adminFlashItemSaveFailed:
		return "We could not save the item. Please check the fields and try again.", false
	}
	return "", false
}

// loadEditorCategories pulls the category list via the existing admin query so
// the editor doesn't need a second repository wired in.
func (h *AdminHandler) loadEditorCategories(c echo.Context, restaurantID common.RestaurantID) ([]admin.ItemEditorCategoryOption, error) {
	menuData, err := h.getMenuAdminUC.Handle(c.Request().Context(), menuQuery.MenuForAdmin{RestaurantID: restaurantID})
	if err != nil {
		return nil, err
	}
	out := make([]admin.ItemEditorCategoryOption, 0, len(menuData.Categories))
	for _, cat := range menuData.Categories {
		out = append(out, admin.ItemEditorCategoryOption{
			ID:   string(cat.Category.ID),
			Name: cat.Category.Name,
		})
	}
	return out, nil
}

// GetItemEditor handles GET /admin/items/:itemID/edit.
func (h *AdminHandler) GetItemEditor(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	itemID := common.ItemID(c.Param("itemID"))
	item, err := h.itemRepo.FindByID(itemID)
	if err != nil {
		return c.Redirect(http.StatusFound, adminMenuRedirect(adminFlashMenuItemNotFound))
	}
	if item.RestaurantID != restaurantID {
		return c.Redirect(http.StatusFound, adminMenuRedirect(adminFlashMenuItemNotFound))
	}
	item, err = menuQuery.ItemWithPresignedPhoto(c.Request().Context(), item, h.photos, h.photoSignerCfg)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load item photo")
	}

	cats, err := h.loadEditorCategories(c, restaurantID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load categories")
	}

	dn, st, ini := commonhttp.LayoutUserStringsFromContext(c)
	switchOpts, activeRole, canCreate, sErr := commonhttp.RestaurantSwitcherData(c, h.membershipRepo, h.restaurantRepo)
	if sErr != nil {
		return c.String(http.StatusInternalServerError, "Failed to load navigation")
	}
	activeLabel := commonhttp.ActiveRestaurantLabel(c.Request().Context(), restaurantID, h.restaurantRepo)

	msg, success := itemEditorFlash(c.QueryParam("flash"))
	page := admin.ItemEditorData{
		Item:            item,
		Categories:      cats,
		CSRFToken:       commonhttp.CSRFToken(c),
		FlashMessage:    msg,
		FlashIsSuccess:  success,
		OptionGroupsDTO: dto.FromDomain(item.OptionGroups),
		AllergenKeys:    menu.AllergenKeys,
	}
	return admin.ItemEditorPage(page, activeLabel, dn, st, ini, switchOpts, activeRole, canCreate).Render(c.Request().Context(), c.Response())
}

// PostItemEditor handles POST /admin/items/:itemID/edit.
func (h *AdminHandler) PostItemEditor(c echo.Context) error {
	restaurantID, err := h.restaurantID(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	itemID := common.ItemID(c.Param("itemID"))

	if err := c.Request().ParseForm(); err != nil {
		return c.Redirect(http.StatusFound, editorItemPath(itemID, adminFlashItemSaveFailed))
	}
	form := c.Request().PostForm

	price, _ := strconv.ParseFloat(form.Get("price"), 64)

	allergens := splitAllergens(form["allergens"])
	badges := splitBadgesCSV(form.Get("badges_csv"))
	groups, err := parseOptionGroups(form.Get("option_groups_json"))
	if err != nil {
		return c.Redirect(http.StatusFound, editorItemPath(itemID, adminFlashItemSaveFailed))
	}
	translations, err := parseTranslations(form.Get("translations_json"))
	if err != nil {
		return c.Redirect(http.StatusFound, editorItemPath(itemID, adminFlashItemSaveFailed))
	}

	spice := strings.ToUpper(strings.TrimSpace(form.Get("spice_level")))
	schedule := strings.ToUpper(strings.TrimSpace(form.Get("schedule")))
	if schedule == "" {
		schedule = menu.ScheduleAllDay
	}
	sku := strings.TrimSpace(form.Get("sku"))
	allow := form.Get("allow_special_instructions") == "on"
	available := form.Get("available") == "on"

	cmd := menuCmd.UpdateMenuItem{
		RestaurantID: restaurantID,
		ItemID:       itemID,
		CategoryID:   common.CategoryID(form.Get("categoryID")),
		Name:         strings.TrimSpace(form.Get("name")),
		Description:  form.Get("description"),
		Price:        price,
		Available:    available,
		IsVegetarian: form.Get("is_vegetarian") == "on",
		IsGlutenFree: form.Get("is_gluten_free") == "on",
		IsSpicy:      spice != "", // legacy boolean: any explicit spice level marks the item as spicy
	}
	cmd.SpiceLevel = &spice
	cmd.Schedule = &schedule
	cmd.SKU = &sku
	vegan := form.Get("is_vegan") == "on"
	cmd.IsVegan = &vegan
	dairyFree := form.Get("is_dairy_free") == "on"
	cmd.IsDairyFree = &dairyFree
	halal := form.Get("is_halal") == "on"
	cmd.IsHalal = &halal
	nutFree := form.Get("is_nut_free") == "on"
	cmd.IsNutFree = &nutFree
	cmd.Allergens = &allergens
	cmd.Badges = &badges
	cmd.AllowSpecialInstructions = &allow
	cmd.OptionGroups = &groups
	cmd.Translations = &translations

	if err := h.updateItemUC.Handle(c.Request().Context(), cmd); err != nil {
		return c.Redirect(http.StatusFound, editorItemPath(itemID, adminFlashItemSaveFailed))
	}
	return c.Redirect(http.StatusFound, editorItemPath(itemID, adminFlashItemSaved))
}

// splitAllergens filters the multi-value allergens form input to the canonical
// set; the domain validator rejects anything unrecognised.
func splitAllergens(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

// splitBadgesCSV parses a comma-separated badges string. NormalizeBadges in the
// domain handles trimming, deduping, and length validation.
func splitBadgesCSV(csv string) []string {
	if csv == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

// parseOptionGroups unmarshals the editor's hidden option_groups_json field
// into domain types. An empty payload means "clear all groups".
func parseOptionGroups(raw string) ([]menu.OptionGroup, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil, nil
	}
	var dtos []dto.OptionGroupDTO
	if err := json.Unmarshal([]byte(raw), &dtos); err != nil {
		return nil, err
	}
	return dto.ToDomain(dtos), nil
}

// parseTranslations decodes the editor's translations JSON textarea into a
// locale→{name,description} map. Empty input clears translations.
func parseTranslations(raw string) (map[string]menu.ItemTranslation, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return nil, nil
	}
	var t map[string]menu.ItemTranslation
	if err := json.Unmarshal([]byte(raw), &t); err != nil {
		return nil, err
	}
	return t, nil
}
