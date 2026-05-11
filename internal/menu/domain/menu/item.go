package menu

import (
	"errors"
	"strings"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"
)

// Option is a single selectable choice within an OptionGroup.
type Option struct {
	ID         string
	Name       string
	PriceDelta float64 // additional cost; 0 means no surcharge
}

// OptionGroup is a set of choices attached to a menu item.
// Required groups are single-select (radio); optional groups are multi-select (checkbox).
type OptionGroup struct {
	ID       string
	Name     string
	Required bool
	Options  []Option
}

// MenuItem represents a food/drink item.
type MenuItem struct {
	ID               common.ItemID
	CategoryID       common.CategoryID
	RestaurantID     common.RestaurantID
	Name             string
	Description      string
	Price            float64
	Currency         money.Currency
	PhotoURL         string
	PhotoOriginalURL string
	IsAvailable      bool
	DisplayOrder     int
	IsVegetarian     bool
	IsGlutenFree     bool
	IsSpicy          bool
	OptionGroups     []OptionGroup
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// HasOptionGroups reports whether the item has any modifier groups.
func (m *MenuItem) HasOptionGroups() bool {
	return len(m.OptionGroups) > 0
}

// Money returns the price as a money.Money value, falling back to USD when
// the item was loaded from a row that predates currency support.
func (m *MenuItem) Money() money.Money {
	c := m.Currency
	if c.IsZero() {
		c = money.USD
	}
	return money.FromMajor(m.Price, c)
}

func NewMenuItem(id common.ItemID, categoryID common.CategoryID, restaurantID common.RestaurantID, name string, price float64) (*MenuItem, error) {
	return NewMenuItemWithCurrency(id, categoryID, restaurantID, name, price, money.USD)
}

// NewMenuItemWithCurrency creates a menu item priced in the given currency.
// For SAT, price is the whole-sat count (no decimals) — pass e.g. 5000 for
// 5,000 sats.
func NewMenuItemWithCurrency(id common.ItemID, categoryID common.CategoryID, restaurantID common.RestaurantID, name string, price float64, currency money.Currency) (*MenuItem, error) {
	if err := ValidateItemName(name); err != nil {
		return nil, err
	}
	if currency.IsZero() {
		currency = money.USD
	}
	if err := ValidatePriceForCurrency(price, currency); err != nil {
		return nil, err
	}

	now := time.Now()
	return &MenuItem{
		ID:           id,
		CategoryID:   categoryID,
		RestaurantID: restaurantID,
		Name:         name,
		Price:        price,
		Currency:     currency,
		IsAvailable:  true,
		DisplayOrder: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func ValidateDisplayOrder(order int) error {
	if order < 0 {
		return errors.New("display order must be >= 0")
	}
	return nil
}

// SetDisplayOrder updates sort position within the category.
func (m *MenuItem) SetDisplayOrder(order int) error {
	if err := ValidateDisplayOrder(order); err != nil {
		return err
	}
	m.DisplayOrder = order
	m.UpdatedAt = time.Now()
	return nil
}

func ValidateItemName(name string) error {
	if len(name) == 0 || len(name) > 100 {
		return errors.New("item name must be between 1 and 100 characters")
	}
	return nil
}

func ValidatePrice(price float64) error {
	return ValidatePriceForCurrency(price, money.USD)
}

// ValidatePriceForCurrency enforces sane bounds per currency. SAT prices
// must be whole numbers (no fractional sats); fiat prices allow decimals
// down to one minor unit.
func ValidatePriceForCurrency(price float64, currency money.Currency) error {
	if price <= 0 {
		return errors.New("price must be greater than 0")
	}
	if currency.Code == money.SAT.Code {
		if price != float64(int64(price)) {
			return errors.New("sat price must be a whole number")
		}
		// 21M BTC = 2.1e15 sats — well within int64 but cap at 1B sats per
		// item as a sanity bound (~ $1M USD-equivalent at $100k/BTC).
		if price > 1_000_000_000 {
			return errors.New("sat price must be less than 1,000,000,000")
		}
		return nil
	}
	if price > 100_000_000 {
		return errors.New("price must be less than 100,000,000")
	}
	return nil
}

func ValidateDescription(description string) error {
	if len(description) > 500 {
		return errors.New("description must be <= 500 characters")
	}
	return nil
}

func (m *MenuItem) SetDescription(description string) error {
	if err := ValidateDescription(description); err != nil {
		return err
	}
	m.Description = description
	m.UpdatedAt = time.Now()
	return nil
}

func (m *MenuItem) SetPhotoURLs(photoURL, photoOriginalURL string) {
	m.PhotoURL = photoURL
	m.PhotoOriginalURL = photoOriginalURL
	m.UpdatedAt = time.Now()
}

// MakeAvailable marks item as available.
func (m *MenuItem) MakeAvailable() {
	m.IsAvailable = true
	m.UpdatedAt = time.Now()
}

// MakeUnavailable marks item as unavailable.
func (m *MenuItem) MakeUnavailable() {
	m.IsAvailable = false
	m.UpdatedAt = time.Now()
}

// SetAvailable updates item availability.
func (m *MenuItem) SetAvailable(isAvailable bool) {
	m.IsAvailable = isAvailable
	m.UpdatedAt = time.Now()
}

func (m *MenuItem) SetDietaryTags(isVegetarian, isGlutenFree, isSpicy bool) {
	m.IsVegetarian = isVegetarian
	m.IsGlutenFree = isGlutenFree
	m.IsSpicy = isSpicy
	m.UpdatedAt = time.Now()
}

// DietaryTagsString returns space-separated tag keys for use as a data attribute.
func (m *MenuItem) DietaryTagsString() string {
	var tags []string
	if m.IsVegetarian {
		tags = append(tags, "vegetarian")
	}
	if m.IsGlutenFree {
		tags = append(tags, "gluten_free")
	}
	if m.IsSpicy {
		tags = append(tags, "spicy")
	}
	return strings.Join(tags, " ")
}
