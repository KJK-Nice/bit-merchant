package http_test

import (
	"bitmerchant/internal/common"

	httpMiddleware "bitmerchant/internal/common/http/middleware"
	"bitmerchant/internal/infrastructure/qr"
	"bitmerchant/internal/infrastructure/repositories/memory"
	menuCmd "bitmerchant/internal/menu/app/command"
	menuQuery "bitmerchant/internal/menu/app/query"
	menuDomain "bitmerchant/internal/menu/domain/menu"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"
	restauranthttp "bitmerchant/internal/restaurant/ports/http"

	// Mock Use Cases for Admin Handler
	restaurantQuery "bitmerchant/internal/restaurant/app/query"
	"context"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAdminEndpoints(t *testing.T) {
	e := echo.New()

	repoRest := memory.NewMemoryRestaurantRepository()
	repoCat := memory.NewMemoryMenuCategoryRepository()
	repoItem := memory.NewMemoryMenuItemRepository()

	createRestUC := restaurantCmd.NewCreateRestaurantHandler(repoRest, nil, nil)
	createCatUC := menuCmd.NewCreateMenuCategoryHandler(repoCat, nil, nil)
	createItemUC := menuCmd.NewCreateMenuItemHandler(repoItem, nil, nil)

	getMenuAdminUC := menuQuery.NewMenuForAdminHandler(repoCat, repoItem, repoRest, nil, menuQuery.PhotoSignerConfig{}, nil, nil)
	updateItemUC := menuCmd.NewUpdateMenuItemHandler(repoItem, repoCat, nil, nil)
	updateCategoryUC := menuCmd.NewUpdateMenuCategoryHandler(repoCat, nil, nil)
	toggleAvailUC := menuCmd.NewToggleMenuItemAvailabilityHandler(repoItem, nil, nil)
	reorderCatUC := menuCmd.NewReorderMenuCategoriesHandler(repoCat, nil, nil)
	reorderItemUC := menuCmd.NewReorderMenuItemsHandler(repoItem, repoCat, nil, nil)
	updateTableUC := restaurantCmd.NewUpdateRestaurantTableCountHandler(repoRest, nil, nil)
	generateQRUC := restaurantQuery.NewRestaurantTableQRImageHandler(qr.NewQRCodeService(), "http://localhost", repoRest, nil, nil)

	membershipRepo := memory.NewMemoryMembershipRepository()
	adminHandler := restauranthttp.NewAdminHandler(
		createRestUC,
		createCatUC,
		createItemUC,
		getMenuAdminUC,
		updateItemUC,
		updateCategoryUC,
		toggleAvailUC,
		nil, // uploadPhotoUC
		reorderCatUC,
		reorderItemUC,
		repoItem,
		updateTableUC,
		generateQRUC,
		membershipRepo,
		repoRest,
	)

	// Seed a restaurant for the dashboard context
	restID := common.RestaurantID("restaurant_1")
	rest, _ := restaurant.NewRestaurant(restID, "Test Rest")
	require.NoError(t, repoRest.Save(rest))

	t.Run("GET /admin/dashboard returns dashboard", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(httpMiddleware.ContextRestaurantID, restID)

		assert.NoError(t, adminHandler.Dashboard(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Menu Management")
	})

	t.Run("POST /admin/category creates category", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/admin/category", strings.NewReader("name=Starters&displayOrder=1"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(httpMiddleware.ContextRestaurantID, restID)

		assert.NoError(t, adminHandler.CreateCategory(c))
		// Expect redirect or HTML fragment depending on implementation (Datastar vs standard)
		// Assuming standard form post for now as per MVP
		assert.Equal(t, http.StatusFound, rec.Code)

		// Verify in repo
		cats, _ := repoCat.FindByRestaurantID(restID)
		assert.NotEmpty(t, cats)
		assert.Equal(t, "Starters", cats[0].Name)
	})

	t.Run("POST /admin/item creates item", func(t *testing.T) {
		// Need a category first
		cat, _ := createCatUC.Handle(context.Background(), menuCmd.CreateMenuCategory{
			RestaurantID: restID,
			Name:         "Mains",
			DisplayOrder: 1,
		})

		form := "name=Steak&price=25.00&description=Juicy&categoryID=" + string(cat.ID)
		req := httptest.NewRequest(http.MethodPost, "/admin/item", strings.NewReader(form))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(httpMiddleware.ContextRestaurantID, restID)

		assert.NoError(t, adminHandler.CreateItem(c))
		assert.Equal(t, http.StatusFound, rec.Code)

		// Verify in repo
		items, _ := repoItem.FindByCategoryID(cat.ID)
		assert.NotEmpty(t, items)
		assert.Equal(t, "Steak", items[0].Name)
	})

	t.Run("GET /admin/items/:itemID/edit renders editor", func(t *testing.T) {
		cat, _ := createCatUC.Handle(context.Background(), menuCmd.CreateMenuCategory{
			RestaurantID: restID, Name: "Editor cat",
		})
		item, _ := createItemUC.Handle(context.Background(), menuCmd.CreateMenuItem{
			RestaurantID: restID, CategoryID: cat.ID, Name: "Pork Belly Bao", Price: 6.50,
		})
		req := httptest.NewRequest(http.MethodGet, "/admin/items/"+string(item.ID)+"/edit", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/items/:itemID/edit")
		c.SetParamNames("itemID")
		c.SetParamValues(string(item.ID))
		c.Set(httpMiddleware.ContextRestaurantID, restID)

		assert.NoError(t, adminHandler.GetItemEditor(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Pork Belly Bao")
		assert.Contains(t, body, "option_groups_json")
		assert.Contains(t, body, "Modifier groups")
	})

	t.Run("POST /admin/items/:itemID/edit roundtrips new fields", func(t *testing.T) {
		cat, _ := createCatUC.Handle(context.Background(), menuCmd.CreateMenuCategory{
			RestaurantID: restID, Name: "Save cat",
		})
		item, _ := createItemUC.Handle(context.Background(), menuCmd.CreateMenuItem{
			RestaurantID: restID, CategoryID: cat.ID, Name: "Save target", Price: 5.00,
		})

		ogJSON := `[{"id":"g1","name":"Sauce","required":true,"min_selections":1,"max_selections":1,"default_option_id":"o1","options":[{"id":"o1","name":"Hoisin","price_delta":0},{"id":"o2","name":"Mayo","price_delta":0.5}]}]`
		form := url.Values{}
		form.Set("name", "Save target")
		form.Set("price", "5.00")
		form.Set("description", "desc")
		form.Set("categoryID", string(cat.ID))
		form.Set("available", "on")
		form.Set("spice_level", "MEDIUM")
		form.Set("schedule", "LUNCH")
		form.Set("sku", "BAO-001")
		form.Set("badges_csv", "Popular, New")
		form.Set("is_vegan", "on")
		form.Set("allow_special_instructions", "on")
		form.Set("option_groups_json", ogJSON)
		form["allergens"] = []string{"Gluten", "Soy"}

		req := httptest.NewRequest(http.MethodPost, "/admin/items/"+string(item.ID)+"/edit", strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/items/:itemID/edit")
		c.SetParamNames("itemID")
		c.SetParamValues(string(item.ID))
		c.Set(httpMiddleware.ContextRestaurantID, restID)

		assert.NoError(t, adminHandler.PostItemEditor(c))
		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Contains(t, rec.Header().Get("Location"), "flash=item_saved")

		saved, err := repoItem.FindByID(item.ID)
		require.NoError(t, err)
		assert.Equal(t, "MEDIUM", saved.SpiceLevel)
		assert.Equal(t, "LUNCH", saved.Schedule)
		assert.Equal(t, "BAO-001", saved.SKU)
		assert.True(t, saved.IsVegan)
		assert.True(t, saved.AllowSpecialInstructions)
		assert.Equal(t, []string{"Popular", "New"}, saved.Badges)
		assert.Equal(t, []string{"Gluten", "Soy"}, saved.Allergens)
		require.Len(t, saved.OptionGroups, 1)
		assert.Equal(t, "Sauce", saved.OptionGroups[0].Name)
		require.NotNil(t, saved.OptionGroups[0].DefaultOptionID)
		assert.Equal(t, "o1", *saved.OptionGroups[0].DefaultOptionID)
	})

	t.Run("POST /admin/items/:itemID/edit rejects invalid spice level", func(t *testing.T) {
		cat, _ := createCatUC.Handle(context.Background(), menuCmd.CreateMenuCategory{
			RestaurantID: restID, Name: "Invalid spice cat",
		})
		item, _ := createItemUC.Handle(context.Background(), menuCmd.CreateMenuItem{
			RestaurantID: restID, CategoryID: cat.ID, Name: "Spice target", Price: 5.00,
		})
		form := url.Values{}
		form.Set("name", "Spice target")
		form.Set("price", "5.00")
		form.Set("categoryID", string(cat.ID))
		form.Set("spice_level", "EXTREME") // invalid

		req := httptest.NewRequest(http.MethodPost, "/admin/items/"+string(item.ID)+"/edit", strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/items/:itemID/edit")
		c.SetParamNames("itemID")
		c.SetParamValues(string(item.ID))
		c.Set(httpMiddleware.ContextRestaurantID, restID)

		assert.NoError(t, adminHandler.PostItemEditor(c))
		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Contains(t, rec.Header().Get("Location"), "flash=item_save_failed")
	})

	t.Run("POST /admin/items/:itemID/edit rejects default option not in group", func(t *testing.T) {
		cat, _ := createCatUC.Handle(context.Background(), menuCmd.CreateMenuCategory{
			RestaurantID: restID, Name: "Default cat",
		})
		item, _ := createItemUC.Handle(context.Background(), menuCmd.CreateMenuItem{
			RestaurantID: restID, CategoryID: cat.ID, Name: "Default target", Price: 5.00,
		})
		ogJSON := `[{"id":"g1","name":"Sauce","required":true,"min_selections":1,"max_selections":1,"default_option_id":"o99","options":[{"id":"o1","name":"Only","price_delta":0}]}]`
		form := url.Values{}
		form.Set("name", "Default target")
		form.Set("price", "5.00")
		form.Set("categoryID", string(cat.ID))
		form.Set("option_groups_json", ogJSON)

		req := httptest.NewRequest(http.MethodPost, "/admin/items/"+string(item.ID)+"/edit", strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/items/:itemID/edit")
		c.SetParamNames("itemID")
		c.SetParamValues(string(item.ID))
		c.Set(httpMiddleware.ContextRestaurantID, restID)

		assert.NoError(t, adminHandler.PostItemEditor(c))
		assert.Contains(t, rec.Header().Get("Location"), "flash=item_save_failed")
	})

	t.Run("POST /admin/items/:itemID/edit with empty option_groups_json clears groups", func(t *testing.T) {
		cat, _ := createCatUC.Handle(context.Background(), menuCmd.CreateMenuCategory{
			RestaurantID: restID, Name: "Empty groups cat",
		})
		item, _ := createItemUC.Handle(context.Background(), menuCmd.CreateMenuItem{
			RestaurantID: restID, CategoryID: cat.ID, Name: "Empty target", Price: 5.00,
		})

		// First, set some groups.
		seedDef := "o1"
		require.NoError(t, item.SetOptionGroups([]menuDomain.OptionGroup{{
			ID: "g1", Name: "Sauce", Required: true, MinSelections: 1, MaxSelections: 1,
			DefaultOptionID: &seedDef,
			Options:         []menuDomain.Option{{ID: "o1", Name: "Hoisin"}},
		}}))
		require.NoError(t, repoItem.Update(item))

		form := url.Values{}
		form.Set("name", "Empty target")
		form.Set("price", "5.00")
		form.Set("categoryID", string(cat.ID))
		form.Set("option_groups_json", "[]")

		req := httptest.NewRequest(http.MethodPost, "/admin/items/"+string(item.ID)+"/edit", strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/items/:itemID/edit")
		c.SetParamNames("itemID")
		c.SetParamValues(string(item.ID))
		c.Set(httpMiddleware.ContextRestaurantID, restID)

		assert.NoError(t, adminHandler.PostItemEditor(c))
		assert.Contains(t, rec.Header().Get("Location"), "flash=item_saved")

		saved, err := repoItem.FindByID(item.ID)
		require.NoError(t, err)
		assert.Empty(t, saved.OptionGroups, "groups should be cleared")
	})
}
