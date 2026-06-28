package domain_test

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"
	"bitmerchant/internal/menu/domain/menu"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewMenuCategory(t *testing.T) {
	t.Run("should create valid category", func(t *testing.T) {
		id := common.CategoryID("cat_1")
		restID := common.RestaurantID("rest_1")
		name := "Appetizers"
		order := 1

		cat, err := menu.NewMenuCategory(id, restID, name, order)

		assert.NoError(t, err)
		assert.NotNil(t, cat)
		assert.Equal(t, id, cat.ID)
		assert.Equal(t, restID, cat.RestaurantID)
		assert.Equal(t, name, cat.Name)
		assert.Equal(t, order, cat.DisplayOrder)
		assert.True(t, cat.IsActive)
		assert.WithinDuration(t, time.Now(), cat.CreatedAt, time.Second)
	})

	t.Run("should fail with empty name", func(t *testing.T) {
		_, err := menu.NewMenuCategory("id", "rid", "", 1)
		assert.Error(t, err)
	})

	t.Run("should fail with negative order", func(t *testing.T) {
		_, err := menu.NewMenuCategory("id", "rid", "name", -1)
		assert.Error(t, err)
	})
}

func TestMenuCategory_SetActive(t *testing.T) {
	cat, _ := menu.NewMenuCategory("id", "rid", "name", 1)

	cat.SetActive(false)
	assert.False(t, cat.IsActive)

	cat.SetActive(true)
	assert.True(t, cat.IsActive)
}

func TestNewMenuItem(t *testing.T) {
	t.Run("should create valid item", func(t *testing.T) {
		id := common.ItemID("item_1")
		catID := common.CategoryID("cat_1")
		restID := common.RestaurantID("rest_1")
		name := "Burger"
		price := 12.50

		item, err := menu.NewMenuItem(id, catID, restID, name, price)

		assert.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, id, item.ID)
		assert.Equal(t, name, item.Name)
		assert.Equal(t, price, item.Price)
		assert.Equal(t, money.USD, item.Currency, "default currency must be USD for legacy callers")
		assert.True(t, item.IsAvailable)
	})

	t.Run("should create satoshi-priced item", func(t *testing.T) {
		item, err := menu.NewMenuItemWithCurrency("item_sat", "cat", "rest", "Espresso", 5_000, money.SAT)
		require.NoError(t, err)
		assert.Equal(t, money.SAT, item.Currency)
		assert.Equal(t, "5,000 sats", item.Money().Format())
	})

	t.Run("should reject fractional sat prices", func(t *testing.T) {
		_, err := menu.NewMenuItemWithCurrency("id", "cat", "rest", "n", 0.5, money.SAT)
		assert.Error(t, err)
	})

	t.Run("should reject sat prices above the cap", func(t *testing.T) {
		_, err := menu.NewMenuItemWithCurrency("id", "cat", "rest", "n", 2_000_000_000, money.SAT)
		assert.Error(t, err)
	})

	t.Run("should fail with invalid price", func(t *testing.T) {
		_, err := menu.NewMenuItem("id", "cid", "rid", "name", 0)
		assert.Error(t, err)
		_, err = menu.NewMenuItem("id", "cid", "rid", "name", -10)
		assert.Error(t, err)
	})

	t.Run("should fail with invalid name", func(t *testing.T) {
		_, err := menu.NewMenuItem("id", "cid", "rid", "", 10)
		assert.Error(t, err)
	})
}

func TestMenuItem_Setters(t *testing.T) {
	item, _ := menu.NewMenuItem("id", "cid", "rid", "name", 10)

	// Description
	err := item.SetDescription("Tasty burger")
	assert.NoError(t, err)
	assert.Equal(t, "Tasty burger", item.Description)

	// Description too long
	longDesc := ""
	for i := 0; i < 501; i++ {
		longDesc += "a"
	}
	err = item.SetDescription(longDesc)
	assert.Error(t, err)

	// Photos
	item.SetPhotoURLs("http://thumb.jpg", "http://orig.jpg")
	assert.Equal(t, "http://thumb.jpg", item.PhotoURL)
	assert.Equal(t, "http://orig.jpg", item.PhotoOriginalURL)

	// Availability
	item.SetAvailable(false)
	assert.False(t, item.IsAvailable)
}

func TestNewMenuItem_Defaults(t *testing.T) {
	item, err := menu.NewMenuItem("id", "cid", "rid", "name", 10)
	require.NoError(t, err)
	assert.Equal(t, menu.ScheduleAllDay, item.Schedule, "new items default to ALL_DAY schedule")
	assert.True(t, item.AllowSpecialInstructions, "new items allow special instructions by default")
	assert.Empty(t, item.SpiceLevel, "new items have no spice level until set")
}

func TestValidateSpiceLevel(t *testing.T) {
	cases := map[string]bool{
		"":        true,
		"MILD":    true,
		"MEDIUM":  true,
		"HOT":     true,
		"mild":    false,
		"EXTREME": false,
	}
	for input, ok := range cases {
		err := menu.ValidateSpiceLevel(input)
		if ok {
			assert.NoError(t, err, "spice level %q should be valid", input)
		} else {
			assert.Error(t, err, "spice level %q should be rejected", input)
		}
	}
}

func TestValidateSchedule(t *testing.T) {
	for _, ok := range []string{"ALL_DAY", "LUNCH", "DINNER", "WEEKEND"} {
		assert.NoError(t, menu.ValidateSchedule(ok))
	}
	for _, bad := range []string{"", "ALL", "all_day", "BREAKFAST"} {
		assert.Error(t, menu.ValidateSchedule(bad))
	}
}

func TestValidateSKU(t *testing.T) {
	assert.NoError(t, menu.ValidateSKU(""))
	assert.NoError(t, menu.ValidateSKU("BAO-001"))
	assert.Error(t, menu.ValidateSKU(string(make([]byte, 33))), "33-byte SKU exceeds cap")
	assert.Error(t, menu.ValidateSKU("BAO\x01"), "non-printable rejected")
}

func TestNormalizeBadges(t *testing.T) {
	t.Run("trims, dedupes, drops empties", func(t *testing.T) {
		got, err := menu.NormalizeBadges([]string{" Popular ", "Popular", "", "New"})
		require.NoError(t, err)
		assert.Equal(t, []string{"Popular", "New"}, got)
	})
	t.Run("rejects too long", func(t *testing.T) {
		_, err := menu.NormalizeBadges([]string{string(make([]byte, 25))})
		assert.Error(t, err)
	})
	t.Run("rejects more than three", func(t *testing.T) {
		_, err := menu.NormalizeBadges([]string{"a", "b", "c", "d"})
		assert.Error(t, err)
	})
}

func TestNormalizeAllergens(t *testing.T) {
	t.Run("accepts known values, dedupes", func(t *testing.T) {
		got, err := menu.NormalizeAllergens([]string{"Gluten", "Soy", "Gluten"})
		require.NoError(t, err)
		assert.Equal(t, []string{"Gluten", "Soy"}, got)
	})
	t.Run("rejects unknown", func(t *testing.T) {
		_, err := menu.NormalizeAllergens([]string{"Plutonium"})
		assert.Error(t, err)
	})
}

func TestValidateOptionGroupRules(t *testing.T) {
	mkOpt := func(id, name string) menu.Option { return menu.Option{ID: id, Name: name, PriceDelta: 0} }
	defaultID := "o1"

	t.Run("happy required pick-1", func(t *testing.T) {
		err := menu.ValidateOptionGroupRules(menu.OptionGroup{
			Name: "Sauce", Required: true, MinSelections: 1, MaxSelections: 1,
			DefaultOptionID: &defaultID,
			Options:         []menu.Option{mkOpt("o1", "Hoisin"), mkOpt("o2", "Mayo")},
		})
		assert.NoError(t, err)
	})

	t.Run("happy optional pick-any", func(t *testing.T) {
		err := menu.ValidateOptionGroupRules(menu.OptionGroup{
			Name: "Extras", MinSelections: 0, MaxSelections: 0,
			Options: []menu.Option{mkOpt("o1", "Pork"), mkOpt("o2", "Egg")},
		})
		assert.NoError(t, err)
	})

	t.Run("required must have min >= 1", func(t *testing.T) {
		err := menu.ValidateOptionGroupRules(menu.OptionGroup{
			Name: "Sauce", Required: true, Options: []menu.Option{mkOpt("o1", "Hoisin")},
		})
		assert.Error(t, err)
	})

	t.Run("max less than min rejected", func(t *testing.T) {
		err := menu.ValidateOptionGroupRules(menu.OptionGroup{
			Name: "X", MinSelections: 2, MaxSelections: 1,
			Options: []menu.Option{mkOpt("o1", "A")},
		})
		assert.Error(t, err)
	})

	t.Run("default option must be in group", func(t *testing.T) {
		missing := "o99"
		err := menu.ValidateOptionGroupRules(menu.OptionGroup{
			Name: "Sauce", Required: true, MinSelections: 1, MaxSelections: 1,
			DefaultOptionID: &missing,
			Options:         []menu.Option{mkOpt("o1", "Hoisin")},
		})
		assert.Error(t, err)
	})

	t.Run("empty name rejected", func(t *testing.T) {
		err := menu.ValidateOptionGroupRules(menu.OptionGroup{Name: "  "})
		assert.Error(t, err)
	})

	t.Run("duplicate option ids rejected", func(t *testing.T) {
		err := menu.ValidateOptionGroupRules(menu.OptionGroup{
			Name: "X", Options: []menu.Option{mkOpt("o1", "A"), mkOpt("o1", "B")},
		})
		assert.Error(t, err)
	})

	t.Run("negative price delta rejected", func(t *testing.T) {
		err := menu.ValidateOptionGroupRules(menu.OptionGroup{
			Name: "X", Options: []menu.Option{{ID: "o1", Name: "A", PriceDelta: -1}},
		})
		assert.Error(t, err)
	})
}

func TestMenuItem_SetOptionGroups_ValidatesEach(t *testing.T) {
	item, _ := menu.NewMenuItem("id", "cid", "rid", "name", 10)

	t.Run("happy path", func(t *testing.T) {
		defaultID := "o1"
		err := item.SetOptionGroups([]menu.OptionGroup{{
			ID: "g1", Name: "Sauce", Required: true, MinSelections: 1, MaxSelections: 1,
			DefaultOptionID: &defaultID,
			Options:         []menu.Option{{ID: "o1", Name: "Hoisin"}},
		}})
		assert.NoError(t, err)
		assert.Len(t, item.OptionGroups, 1)
	})

	t.Run("first invalid group rejects whole set", func(t *testing.T) {
		err := item.SetOptionGroups([]menu.OptionGroup{
			{ID: "g1", Name: "OK", Options: []menu.Option{{ID: "o1", Name: "A"}}},
			{ID: "g2", Name: "", Options: nil}, // invalid
		})
		assert.Error(t, err)
	})
}

func TestMenuItem_SetBadges_SetAllergens_Roundtrip(t *testing.T) {
	item, _ := menu.NewMenuItem("id", "cid", "rid", "name", 10)
	require.NoError(t, item.SetBadges([]string{"Popular", "  ", "Popular"}))
	assert.Equal(t, []string{"Popular"}, item.Badges)

	require.NoError(t, item.SetAllergens([]string{"Gluten", "Soy"}))
	assert.Equal(t, []string{"Gluten", "Soy"}, item.Allergens)
}

func TestMenuItem_DietaryTagsString_IncludesExpandedSet(t *testing.T) {
	item, _ := menu.NewMenuItem("id", "cid", "rid", "name", 10)
	item.SetDietaryFlags(true, true, false, true, false, true)
	tags := item.DietaryTagsString()
	assert.Contains(t, tags, "vegetarian")
	assert.Contains(t, tags, "vegan")
	assert.Contains(t, tags, "dairy_free")
	assert.Contains(t, tags, "nut_free")
}

func TestMenuItem_Translations(t *testing.T) {
	item, err := menu.NewMenuItem("i1", "c1", "r1", "Pork Bao", 6.50)
	require.NoError(t, err)
	item.Description = "Steamed bun with pork"

	// No translations → fall back to base name/description.
	assert.Equal(t, "Pork Bao", item.NameFor("es"))
	assert.Equal(t, "Steamed bun with pork", item.DescriptionFor("es"))
	assert.Nil(t, item.Locales())

	require.NoError(t, item.SetTranslations(map[string]menu.ItemTranslation{
		"ES": {Name: "Bao de cerdo", Description: "Bollo al vapor con cerdo"},
		"th": {Name: "ซาลาเปาหมู"},
		"  ": {Name: "ignored blank locale"},
		"fr": {Name: "", Description: ""}, // fully empty → dropped
	}))

	// Locale codes are lowercased; blank/empty entries dropped.
	assert.Equal(t, []string{"es", "th"}, item.Locales())
	assert.Equal(t, "Bao de cerdo", item.NameFor("es"))
	assert.Equal(t, "Bollo al vapor con cerdo", item.DescriptionFor("es"))
	// th has a name but no description → description falls back to base.
	assert.Equal(t, "ซาลาเปาหมู", item.NameFor("th"))
	assert.Equal(t, "Steamed bun with pork", item.DescriptionFor("th"))
	// Unknown locale → base.
	assert.Equal(t, "Pork Bao", item.NameFor("de"))
}
