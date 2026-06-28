package menu

import (
	"errors"
	"sort"
	"strings"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"
)

// ItemTranslation holds a localized name/description for a menu item.
type ItemTranslation struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Option is a single selectable choice within an OptionGroup.
type Option struct {
	ID         string
	Name       string
	PriceDelta float64 // additional cost; 0 means no surcharge
}

// OptionGroup is a set of choices attached to a menu item.
// Required groups must have MinSelections >= 1; MaxSelections caps multi-select groups.
// DefaultOptionID, when non-nil, must reference an option present in Options.
type OptionGroup struct {
	ID              string
	Name            string
	Required        bool
	MinSelections   int
	MaxSelections   int
	DefaultOptionID *string
	Options         []Option
}

// Spice level values. Empty string ("") means unspecified — render no chip.
const (
	SpiceLevelMild   = "MILD"
	SpiceLevelMedium = "MEDIUM"
	SpiceLevelHot    = "HOT"
)

// Schedule values.
const (
	ScheduleAllDay  = "ALL_DAY"
	ScheduleLunch   = "LUNCH"
	ScheduleDinner  = "DINNER"
	ScheduleWeekend = "WEEKEND"
)

// AllergenKeys is the canonical set of allergens a menu item may declare.
var AllergenKeys = []string{"Gluten", "Soy", "Sesame", "Peanut", "Egg", "Dairy", "Shellfish"}

// MaxBadges caps how many free-text badges an item may carry.
const MaxBadges = 3

// MaxBadgeLength caps the length of any single badge.
const MaxBadgeLength = 24

// MaxSKULength caps the length of an SKU.
const MaxSKULength = 32

// MenuItem represents a food/drink item.
type MenuItem struct {
	ID                       common.ItemID
	CategoryID               common.CategoryID
	RestaurantID             common.RestaurantID
	Name                     string
	Description              string
	Price                    float64
	Currency                 money.Currency
	PhotoURL                 string
	PhotoOriginalURL         string
	IsAvailable              bool
	DisplayOrder             int
	IsVegetarian             bool
	IsGlutenFree             bool
	IsSpicy                  bool // deprecated; use SpiceLevel. Kept for one release.
	IsVegan                  bool
	IsDairyFree              bool
	IsHalal                  bool
	IsNutFree                bool
	SpiceLevel               string
	Schedule                 string
	SKU                      string
	Allergens                []string
	Badges                   []string
	AllowSpecialInstructions bool
	OptionGroups             []OptionGroup
	// Translations maps a locale code (e.g. "es", "th") to a localized
	// name/description. The base Name/Description are the default (English)
	// and the fallback when a locale has no entry.
	Translations map[string]ItemTranslation
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NameFor returns the item name in the given locale, falling back to the base
// name when there is no (non-empty) translation.
func (m *MenuItem) NameFor(locale string) string {
	if t, ok := m.Translations[locale]; ok && t.Name != "" {
		return t.Name
	}
	return m.Name
}

// DescriptionFor returns the item description in the given locale, falling back
// to the base description when there is no (non-empty) translation.
func (m *MenuItem) DescriptionFor(locale string) string {
	if t, ok := m.Translations[locale]; ok && t.Description != "" {
		return t.Description
	}
	return m.Description
}

// Locales returns the sorted locale codes that have a translation entry.
func (m *MenuItem) Locales() []string {
	if len(m.Translations) == 0 {
		return nil
	}
	out := make([]string, 0, len(m.Translations))
	for code := range m.Translations {
		out = append(out, code)
	}
	sort.Strings(out)
	return out
}

// SetTranslations validates and replaces the translation map. Locale codes are
// trimmed/lowercased; blank codes and fully-empty entries are dropped.
func (m *MenuItem) SetTranslations(translations map[string]ItemTranslation) error {
	cleaned := make(map[string]ItemTranslation, len(translations))
	for code, t := range translations {
		code = strings.ToLower(strings.TrimSpace(code))
		if code == "" || len(code) > 8 {
			continue
		}
		t.Name = strings.TrimSpace(t.Name)
		t.Description = strings.TrimSpace(t.Description)
		if t.Name == "" && t.Description == "" {
			continue
		}
		cleaned[code] = t
	}
	if len(cleaned) == 0 {
		cleaned = nil
	}
	m.Translations = cleaned
	m.UpdatedAt = time.Now()
	return nil
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
		ID:                       id,
		CategoryID:               categoryID,
		RestaurantID:             restaurantID,
		Name:                     name,
		Price:                    price,
		Currency:                 currency,
		IsAvailable:              true,
		DisplayOrder:             0,
		Schedule:                 ScheduleAllDay,
		AllowSpecialInstructions: true,
		CreatedAt:                now,
		UpdatedAt:                now,
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

// SetDietaryFlags assigns the full dietary set in one call.
func (m *MenuItem) SetDietaryFlags(isVegetarian, isVegan, isGlutenFree, isDairyFree, isHalal, isNutFree bool) {
	m.IsVegetarian = isVegetarian
	m.IsVegan = isVegan
	m.IsGlutenFree = isGlutenFree
	m.IsDairyFree = isDairyFree
	m.IsHalal = isHalal
	m.IsNutFree = isNutFree
	m.UpdatedAt = time.Now()
}

// DietaryTagsString returns space-separated tag keys for use as a data attribute.
func (m *MenuItem) DietaryTagsString() string {
	var tags []string
	if m.IsVegetarian {
		tags = append(tags, "vegetarian")
	}
	if m.IsVegan {
		tags = append(tags, "vegan")
	}
	if m.IsGlutenFree {
		tags = append(tags, "gluten_free")
	}
	if m.IsDairyFree {
		tags = append(tags, "dairy_free")
	}
	if m.IsHalal {
		tags = append(tags, "halal")
	}
	if m.IsNutFree {
		tags = append(tags, "nut_free")
	}
	if m.IsSpicy {
		tags = append(tags, "spicy")
	}
	return strings.Join(tags, " ")
}

// ValidateSpiceLevel accepts "" (unspecified) or one of MILD/MEDIUM/HOT.
func ValidateSpiceLevel(level string) error {
	switch level {
	case "", SpiceLevelMild, SpiceLevelMedium, SpiceLevelHot:
		return nil
	}
	return errors.New("spice level must be one of MILD, MEDIUM, HOT")
}

// ValidateSchedule accepts one of ALL_DAY/LUNCH/DINNER/WEEKEND. Empty maps to ALL_DAY at the caller.
func ValidateSchedule(schedule string) error {
	switch schedule {
	case ScheduleAllDay, ScheduleLunch, ScheduleDinner, ScheduleWeekend:
		return nil
	}
	return errors.New("schedule must be one of ALL_DAY, LUNCH, DINNER, WEEKEND")
}

// ValidateSKU enforces optional, ASCII, length cap.
func ValidateSKU(sku string) error {
	if len(sku) > MaxSKULength {
		return errors.New("sku must be <= 32 characters")
	}
	for _, r := range sku {
		if r > 0x7E || r < 0x20 {
			return errors.New("sku must be printable ASCII")
		}
	}
	return nil
}

// NormalizeBadges trims, drops empties, de-duplicates (case-sensitive), and enforces caps.
func NormalizeBadges(in []string) ([]string, error) {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, b := range in {
		b = strings.TrimSpace(b)
		if b == "" {
			continue
		}
		if len(b) > MaxBadgeLength {
			return nil, errors.New("badge must be <= 24 characters")
		}
		if _, dup := seen[b]; dup {
			continue
		}
		seen[b] = struct{}{}
		out = append(out, b)
	}
	if len(out) > MaxBadges {
		return nil, errors.New("at most 3 badges allowed")
	}
	return out, nil
}

// NormalizeAllergens validates against AllergenKeys and de-duplicates.
func NormalizeAllergens(in []string) ([]string, error) {
	allowed := map[string]struct{}{}
	for _, k := range AllergenKeys {
		allowed[k] = struct{}{}
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, a := range in {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		if _, ok := allowed[a]; !ok {
			return nil, errors.New("unknown allergen: " + a)
		}
		if _, dup := seen[a]; dup {
			continue
		}
		seen[a] = struct{}{}
		out = append(out, a)
	}
	return out, nil
}

// ValidateOptionGroupRules enforces: MinSelections >= 0, MaxSelections >= MinSelections when MaxSelections > 0,
// Required => MinSelections >= 1, DefaultOptionID (if set) references an Option in the group, and option IDs are unique.
func ValidateOptionGroupRules(g OptionGroup) error {
	if err := validateOptionGroupHeader(g); err != nil {
		return err
	}
	ids, err := validateOptionGroupOptions(g.Options)
	if err != nil {
		return err
	}
	if g.DefaultOptionID != nil {
		if _, ok := ids[*g.DefaultOptionID]; !ok {
			return errors.New("default option id must reference an option in the group")
		}
	}
	return nil
}

func validateOptionGroupHeader(g OptionGroup) error {
	if strings.TrimSpace(g.Name) == "" {
		return errors.New("option group name must not be empty")
	}
	if g.MinSelections < 0 {
		return errors.New("option group MinSelections must be >= 0")
	}
	if g.MaxSelections > 0 && g.MaxSelections < g.MinSelections {
		return errors.New("option group MaxSelections must be >= MinSelections")
	}
	if g.Required && g.MinSelections < 1 {
		return errors.New("required option group must have MinSelections >= 1")
	}
	return nil
}

func validateOptionGroupOptions(opts []Option) (map[string]struct{}, error) {
	seen := make(map[string]struct{}, len(opts))
	for _, o := range opts {
		if err := validateOption(o); err != nil {
			return nil, err
		}
		if _, dup := seen[o.ID]; dup {
			return nil, errors.New("option ids must be unique within a group")
		}
		seen[o.ID] = struct{}{}
	}
	return seen, nil
}

func validateOption(o Option) error {
	if strings.TrimSpace(o.Name) == "" {
		return errors.New("option name must not be empty")
	}
	if o.PriceDelta < 0 {
		return errors.New("option price delta must be >= 0")
	}
	if o.ID == "" {
		return errors.New("option id must not be empty")
	}
	return nil
}

// SetSpiceLevel validates and assigns the spice level.
func (m *MenuItem) SetSpiceLevel(level string) error {
	if err := ValidateSpiceLevel(level); err != nil {
		return err
	}
	m.SpiceLevel = level
	m.UpdatedAt = time.Now()
	return nil
}

// SetSchedule validates and assigns the schedule.
func (m *MenuItem) SetSchedule(schedule string) error {
	if err := ValidateSchedule(schedule); err != nil {
		return err
	}
	m.Schedule = schedule
	m.UpdatedAt = time.Now()
	return nil
}

// SetSKU validates and assigns the SKU.
func (m *MenuItem) SetSKU(sku string) error {
	if err := ValidateSKU(sku); err != nil {
		return err
	}
	m.SKU = sku
	m.UpdatedAt = time.Now()
	return nil
}

// SetBadges normalises and assigns the badges.
func (m *MenuItem) SetBadges(badges []string) error {
	normalised, err := NormalizeBadges(badges)
	if err != nil {
		return err
	}
	m.Badges = normalised
	m.UpdatedAt = time.Now()
	return nil
}

// SetAllergens normalises and assigns the allergens.
func (m *MenuItem) SetAllergens(allergens []string) error {
	normalised, err := NormalizeAllergens(allergens)
	if err != nil {
		return err
	}
	m.Allergens = normalised
	m.UpdatedAt = time.Now()
	return nil
}

// SetAllowSpecialInstructions toggles the per-item special-instructions textarea.
func (m *MenuItem) SetAllowSpecialInstructions(allow bool) {
	m.AllowSpecialInstructions = allow
	m.UpdatedAt = time.Now()
}

// SetOptionGroups validates each group and replaces the slice wholesale.
func (m *MenuItem) SetOptionGroups(groups []OptionGroup) error {
	for _, g := range groups {
		if err := ValidateOptionGroupRules(g); err != nil {
			return err
		}
	}
	m.OptionGroups = groups
	m.UpdatedAt = time.Now()
	return nil
}
