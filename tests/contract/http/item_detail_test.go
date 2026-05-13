package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bitmerchant/internal/infrastructure/repositories/memory"
	menuQuery "bitmerchant/internal/menu/app/query"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/ordering/app/cart"
	orderinghttp "bitmerchant/internal/ordering/ports/http"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newItemDetailContext wires the bits an item-detail test needs.
func newItemDetailContext(t *testing.T, item *menu.MenuItem) (*orderinghttp.CartHandler, *echo.Echo, *httptest.ResponseRecorder, echo.Context) {
	t.Helper()
	cartService := cart.NewCartService()
	itemRepo := memory.NewMemoryMenuItemRepository()
	require.NoError(t, itemRepo.Save(item))

	h := orderinghttp.NewCartHandler(cartService, itemRepo, nil, menuQuery.PhotoSignerConfig{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/menu/item/"+string(item.ID)+"?restaurantID="+string(item.RestaurantID), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/menu/item/:itemID")
	c.SetParamNames("itemID")
	c.SetParamValues(string(item.ID))
	c.Set("sessionID", "sess_test")
	return h, e, rec, c
}

func TestItemDetail_RendersBadgesAllergensSpice(t *testing.T) {
	item, _ := menu.NewMenuItem("i_detail_1", "c1", "r1", "Pork Belly Bao", 6.50)
	require.NoError(t, item.SetBadges([]string{"Popular"}))
	require.NoError(t, item.SetAllergens([]string{"Gluten", "Soy"}))
	require.NoError(t, item.SetSpiceLevel(menu.SpiceLevelMedium))
	item.SetDietaryFlags(false, false, true, true, false, false) // GF + DF

	h, _, rec, c := newItemDetailContext(t, item)

	assert.NoError(t, h.GetItemDetail(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "Popular", "badge text must render")
	assert.Contains(t, body, "Contains: gluten, soy", "allergens line renders lowercased")
	assert.Contains(t, body, "Medium", "spice level label renders")
	assert.Contains(t, body, "Dairy-free", "expanded dietary set renders")
}

func TestItemDetail_HidesSpiceChipWhenEmpty(t *testing.T) {
	item, _ := menu.NewMenuItem("i_detail_2", "c1", "r1", "Plain Rice", 2.50)
	// No SpiceLevel set
	h, _, rec, c := newItemDetailContext(t, item)
	assert.NoError(t, h.GetItemDetail(c))
	body := rec.Body.String()
	assert.NotContains(t, body, ">Mild<", "no Mild chip")
	assert.NotContains(t, body, ">Medium<", "no Medium chip")
	assert.NotContains(t, body, ">Hot<", "no Hot chip")
}

func TestItemDetail_GatesSpecialInstructionsTextarea(t *testing.T) {
	item, _ := menu.NewMenuItem("i_detail_3", "c1", "r1", "No-notes item", 4.00)
	item.SetAllowSpecialInstructions(false)

	h, _, rec, c := newItemDetailContext(t, item)
	assert.NoError(t, h.GetItemDetail(c))
	body := rec.Body.String()
	assert.NotContains(t, body, `name="specialInstructions"`, "textarea should be omitted when allow=false")
	assert.NotContains(t, body, "Special instructions", "label should not render")
}

func TestItemDetail_AllowsSpecialInstructionsByDefault(t *testing.T) {
	item, _ := menu.NewMenuItem("i_detail_4", "c1", "r1", "With-notes item", 4.00)
	// NewMenuItem defaults AllowSpecialInstructions = true.

	h, _, rec, c := newItemDetailContext(t, item)
	assert.NoError(t, h.GetItemDetail(c))
	body := rec.Body.String()
	assert.Contains(t, body, `name="specialInstructions"`, "textarea present by default")
}
