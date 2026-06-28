package adapters

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/money"
	"bitmerchant/internal/menu/domain/menu"
)

type PostgresItemRepository struct {
	db *sql.DB
}

func NewPostgresItemRepository(db *sql.DB) *PostgresItemRepository {
	return &PostgresItemRepository{db: db}
}

const itemSelectCols = `id, category_id, restaurant_id, name, description, price, COALESCE(currency, 'USD'), COALESCE(price_minor, 0), photo_url, photo_original_url, is_available, display_order, created_at, updated_at, COALESCE(is_vegetarian, false), COALESCE(is_gluten_free, false), COALESCE(is_spicy, false), COALESCE(option_groups, '[]'::jsonb), COALESCE(spice_level, ''), COALESCE(sku, ''), COALESCE(schedule, 'ALL_DAY'), COALESCE(is_vegan, false), COALESCE(is_dairy_free, false), COALESCE(is_halal, false), COALESCE(is_nut_free, false), COALESCE(allergens, '[]'::jsonb), COALESCE(badges, '[]'::jsonb), COALESCE(allow_special_instructions, true), COALESCE(translations, '{}'::jsonb)`

func (r *PostgresItemRepository) Save(item *menu.MenuItem) error {
	currency, priceMinor := itemCurrencyAndMinor(item)
	optionGroupsJSON, err := marshalOptionGroups(item.OptionGroups)
	if err != nil {
		return err
	}
	translationsJSON, err := marshalTranslations(item.Translations)
	if err != nil {
		return err
	}
	allergensJSON, err := marshalStringList(item.Allergens)
	if err != nil {
		return err
	}
	badgesJSON, err := marshalStringList(item.Badges)
	if err != nil {
		return err
	}
	schedule := item.Schedule
	if schedule == "" {
		schedule = menu.ScheduleAllDay
	}
	_, err = r.db.Exec(
		`INSERT INTO menu_items (id, category_id, restaurant_id, name, description, price, currency, price_minor, photo_url, photo_original_url, is_available, display_order, created_at, updated_at, is_vegetarian, is_gluten_free, is_spicy, option_groups, spice_level, sku, schedule, is_vegan, is_dairy_free, is_halal, is_nut_free, allergens, badges, allow_special_instructions, translations)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, NULLIF($19, ''), $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)
		 ON CONFLICT (id) DO UPDATE
		 SET category_id = EXCLUDED.category_id, name = EXCLUDED.name,
		     description = EXCLUDED.description, price = EXCLUDED.price,
		     currency = EXCLUDED.currency, price_minor = EXCLUDED.price_minor,
		     photo_url = EXCLUDED.photo_url, photo_original_url = EXCLUDED.photo_original_url,
		     is_available = EXCLUDED.is_available, display_order = EXCLUDED.display_order,
		     is_vegetarian = EXCLUDED.is_vegetarian, is_gluten_free = EXCLUDED.is_gluten_free,
		     is_spicy = EXCLUDED.is_spicy, option_groups = EXCLUDED.option_groups,
		     spice_level = EXCLUDED.spice_level, sku = EXCLUDED.sku, schedule = EXCLUDED.schedule,
		     is_vegan = EXCLUDED.is_vegan, is_dairy_free = EXCLUDED.is_dairy_free,
		     is_halal = EXCLUDED.is_halal, is_nut_free = EXCLUDED.is_nut_free,
		     allergens = EXCLUDED.allergens, badges = EXCLUDED.badges,
		     allow_special_instructions = EXCLUDED.allow_special_instructions,
		     translations = EXCLUDED.translations,
		     updated_at = EXCLUDED.updated_at`,
		string(item.ID), string(item.CategoryID), string(item.RestaurantID),
		item.Name, item.Description, item.Price,
		currency.Code, priceMinor,
		item.PhotoURL, item.PhotoOriginalURL, item.IsAvailable, item.DisplayOrder,
		item.CreatedAt, item.UpdatedAt,
		item.IsVegetarian, item.IsGlutenFree, item.IsSpicy, optionGroupsJSON,
		item.SpiceLevel, item.SKU, schedule,
		item.IsVegan, item.IsDairyFree, item.IsHalal, item.IsNutFree,
		allergensJSON, badgesJSON, item.AllowSpecialInstructions, translationsJSON)
	return err
}

func (r *PostgresItemRepository) FindByID(id common.ItemID) (*menu.MenuItem, error) {
	row := r.db.QueryRow(
		`SELECT `+itemSelectCols+`
		 FROM menu_items WHERE id = $1`, string(id))
	return scanItem(row)
}

func (r *PostgresItemRepository) FindByCategoryID(categoryID common.CategoryID) ([]*menu.MenuItem, error) {
	return r.queryItems(`SELECT `+itemSelectCols+` FROM menu_items WHERE category_id = $1`, string(categoryID))
}

func (r *PostgresItemRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*menu.MenuItem, error) {
	return r.queryItems(`SELECT `+itemSelectCols+` FROM menu_items WHERE restaurant_id = $1`, string(restaurantID))
}

func (r *PostgresItemRepository) FindAvailableByRestaurantID(restaurantID common.RestaurantID) ([]*menu.MenuItem, error) {
	return r.queryItems(`SELECT `+itemSelectCols+` FROM menu_items WHERE restaurant_id = $1 AND is_available = true`, string(restaurantID))
}

func (r *PostgresItemRepository) Update(item *menu.MenuItem) error {
	currency, priceMinor := itemCurrencyAndMinor(item)
	optionGroupsJSON, err := marshalOptionGroups(item.OptionGroups)
	if err != nil {
		return err
	}
	translationsJSON, err := marshalTranslations(item.Translations)
	if err != nil {
		return err
	}
	allergensJSON, err := marshalStringList(item.Allergens)
	if err != nil {
		return err
	}
	badgesJSON, err := marshalStringList(item.Badges)
	if err != nil {
		return err
	}
	schedule := item.Schedule
	if schedule == "" {
		schedule = menu.ScheduleAllDay
	}
	result, err := r.db.Exec(
		`UPDATE menu_items SET category_id=$2, name=$3, description=$4, price=$5, currency=$6, price_minor=$7, photo_url=$8, photo_original_url=$9, is_available=$10, display_order=$11, is_vegetarian=$12, is_gluten_free=$13, is_spicy=$14, option_groups=$15, updated_at=$16,
		     spice_level=NULLIF($17, ''), sku=$18, schedule=$19,
		     is_vegan=$20, is_dairy_free=$21, is_halal=$22, is_nut_free=$23,
		     allergens=$24, badges=$25, allow_special_instructions=$26, translations=$27
		 WHERE id=$1`,
		string(item.ID), string(item.CategoryID), item.Name, item.Description, item.Price,
		currency.Code, priceMinor,
		item.PhotoURL, item.PhotoOriginalURL, item.IsAvailable, item.DisplayOrder,
		item.IsVegetarian, item.IsGlutenFree, item.IsSpicy, optionGroupsJSON, item.UpdatedAt,
		item.SpiceLevel, item.SKU, schedule,
		item.IsVegan, item.IsDairyFree, item.IsHalal, item.IsNutFree,
		allergensJSON, badgesJSON, item.AllowSpecialInstructions, translationsJSON)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("menu item not found")
	}
	return nil
}

func (r *PostgresItemRepository) Delete(id common.ItemID) error {
	result, err := r.db.Exec(`DELETE FROM menu_items WHERE id = $1`, string(id))
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("menu item not found")
	}
	return nil
}

func (r *PostgresItemRepository) CountByRestaurantID(restaurantID common.RestaurantID) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM menu_items WHERE restaurant_id = $1 AND photo_url != ''`, string(restaurantID)).Scan(&count)
	return count, err
}

func (r *PostgresItemRepository) ReorderItemsInCategory(restaurantID common.RestaurantID, categoryID common.CategoryID, orderedItemIDs []common.ItemID) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	existing, err := r.loadCategoryItemIDs(tx, restaurantID, categoryID)
	if err != nil {
		return err
	}

	if err := validateReorderItemList(existing, orderedItemIDs); err != nil {
		return err
	}

	if err := applyCategoryItemOrder(tx, restaurantID, categoryID, orderedItemIDs); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostgresItemRepository) loadCategoryItemIDs(tx *sql.Tx, restaurantID common.RestaurantID, categoryID common.CategoryID) ([]common.ItemID, error) {
	rows, err := tx.Query(
		`SELECT id FROM menu_items WHERE restaurant_id = $1 AND category_id = $2`,
		string(restaurantID), string(categoryID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []common.ItemID
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, common.ItemID(id))
	}
	return ids, rows.Err()
}

func validateReorderItemList(existing []common.ItemID, orderedItemIDs []common.ItemID) error {
	if len(orderedItemIDs) != len(existing) {
		return errors.New("item list does not match category")
	}

	want := make(map[common.ItemID]struct{}, len(existing))
	for _, id := range existing {
		want[id] = struct{}{}
	}
	for _, id := range orderedItemIDs {
		if _, ok := want[id]; !ok {
			return errors.New("invalid item in reorder list")
		}
		delete(want, id)
	}
	if len(want) != 0 {
		return errors.New("item list does not match category")
	}
	return nil
}

func applyCategoryItemOrder(tx *sql.Tx, restaurantID common.RestaurantID, categoryID common.CategoryID, orderedItemIDs []common.ItemID) error {
	for i, id := range orderedItemIDs {
		res, err := tx.Exec(
			`UPDATE menu_items SET display_order = $2, updated_at = $3 WHERE id = $1 AND restaurant_id = $4 AND category_id = $5`,
			string(id), i, time.Now(), string(restaurantID), string(categoryID))
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return errors.New("menu item not found")
		}
	}
	return nil
}

func (r *PostgresItemRepository) queryItems(query string, args ...interface{}) ([]*menu.MenuItem, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*menu.MenuItem
	for rows.Next() {
		item, err := scanItemRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

type itemRowFields struct {
	id, catID, restID, name  string
	description              sql.NullString
	price                    float64
	currencyCode             string
	priceMinor               int64
	photoURL, photoOrigURL   sql.NullString
	isAvailable              bool
	displayOrder             int
	createdAt, updatedAt     time.Time
	isVegetarian             bool
	isGlutenFree             bool
	isSpicy                  bool
	optionGroupsJSON         []byte
	spiceLevel               string
	sku                      string
	schedule                 string
	isVegan                  bool
	isDairyFree              bool
	isHalal                  bool
	isNutFree                bool
	allergensJSON            []byte
	badgesJSON               []byte
	allowSpecialInstructions bool
	translationsJSON         []byte
}

func (f *itemRowFields) toMenuItem() *menu.MenuItem {
	currency, err := money.Parse(f.currencyCode)
	if err != nil {
		currency = money.USD
	}
	groups := unmarshalOptionGroups(f.optionGroupsJSON)
	allergens := unmarshalStringList(f.allergensJSON)
	badges := unmarshalStringList(f.badgesJSON)
	translations := unmarshalTranslations(f.translationsJSON)
	return &menu.MenuItem{
		ID: common.ItemID(f.id), CategoryID: common.CategoryID(f.catID),
		RestaurantID: common.RestaurantID(f.restID), Name: f.name,
		Description: f.description.String, Price: f.price, Currency: currency,
		PhotoURL: f.photoURL.String, PhotoOriginalURL: f.photoOrigURL.String,
		IsAvailable: f.isAvailable, DisplayOrder: f.displayOrder,
		IsVegetarian: f.isVegetarian, IsGlutenFree: f.isGlutenFree, IsSpicy: f.isSpicy,
		IsVegan: f.isVegan, IsDairyFree: f.isDairyFree, IsHalal: f.isHalal, IsNutFree: f.isNutFree,
		SpiceLevel: f.spiceLevel, SKU: f.sku, Schedule: f.schedule,
		Allergens: allergens, Badges: badges,
		AllowSpecialInstructions: f.allowSpecialInstructions,
		OptionGroups:             groups,
		Translations:             translations,
		CreatedAt:                f.createdAt, UpdatedAt: f.updatedAt,
	}
}

func (f *itemRowFields) scanTargets() []any {
	return []any{
		&f.id, &f.catID, &f.restID, &f.name, &f.description, &f.price, &f.currencyCode, &f.priceMinor,
		&f.photoURL, &f.photoOrigURL, &f.isAvailable, &f.displayOrder, &f.createdAt, &f.updatedAt,
		&f.isVegetarian, &f.isGlutenFree, &f.isSpicy, &f.optionGroupsJSON,
		&f.spiceLevel, &f.sku, &f.schedule,
		&f.isVegan, &f.isDairyFree, &f.isHalal, &f.isNutFree,
		&f.allergensJSON, &f.badgesJSON, &f.allowSpecialInstructions, &f.translationsJSON,
	}
}

func scanItem(row *sql.Row) (*menu.MenuItem, error) {
	var f itemRowFields
	if err := row.Scan(f.scanTargets()...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("menu item not found")
		}
		return nil, err
	}
	return f.toMenuItem(), nil
}

func scanItemRows(rows *sql.Rows) (*menu.MenuItem, error) {
	var f itemRowFields
	if err := rows.Scan(f.scanTargets()...); err != nil {
		return nil, err
	}
	return f.toMenuItem(), nil
}

// jsonOptionGroup mirrors menu.OptionGroup for JSON serialisation (snake_case keys).
type jsonOptionGroup struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Required        bool         `json:"required"`
	MinSelections   int          `json:"min_selections"`
	MaxSelections   int          `json:"max_selections"`
	DefaultOptionID *string      `json:"default_option_id,omitempty"`
	Options         []jsonOption `json:"options"`
}

type jsonOption struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	PriceDelta float64 `json:"price_delta"`
}

func marshalOptionGroups(groups []menu.OptionGroup) ([]byte, error) {
	if len(groups) == 0 {
		return []byte("[]"), nil
	}
	jgs := make([]jsonOptionGroup, len(groups))
	for i, g := range groups {
		jos := make([]jsonOption, len(g.Options))
		for j, o := range g.Options {
			jos[j] = jsonOption{ID: o.ID, Name: o.Name, PriceDelta: o.PriceDelta}
		}
		jgs[i] = jsonOptionGroup{
			ID:              g.ID,
			Name:            g.Name,
			Required:        g.Required,
			MinSelections:   g.MinSelections,
			MaxSelections:   g.MaxSelections,
			DefaultOptionID: g.DefaultOptionID,
			Options:         jos,
		}
	}
	return json.Marshal(jgs)
}

func marshalTranslations(t map[string]menu.ItemTranslation) ([]byte, error) {
	if len(t) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(t)
}

func unmarshalTranslations(data []byte) map[string]menu.ItemTranslation {
	if len(data) == 0 {
		return nil
	}
	var t map[string]menu.ItemTranslation
	if err := json.Unmarshal(data, &t); err != nil {
		return nil
	}
	if len(t) == 0 {
		return nil
	}
	return t
}

func unmarshalOptionGroups(data []byte) []menu.OptionGroup {
	if len(data) == 0 {
		return nil
	}
	var jgs []jsonOptionGroup
	if err := json.Unmarshal(data, &jgs); err != nil {
		return nil
	}
	groups := make([]menu.OptionGroup, len(jgs))
	for i, jg := range jgs {
		opts := make([]menu.Option, len(jg.Options))
		for j, jo := range jg.Options {
			opts[j] = menu.Option{ID: jo.ID, Name: jo.Name, PriceDelta: jo.PriceDelta}
		}
		groups[i] = menu.OptionGroup{
			ID:              jg.ID,
			Name:            jg.Name,
			Required:        jg.Required,
			MinSelections:   jg.MinSelections,
			MaxSelections:   jg.MaxSelections,
			DefaultOptionID: jg.DefaultOptionID,
			Options:         opts,
		}
	}
	return groups
}

// marshalStringList serialises a []string into a JSONB array. Empty/nil slices
// become "[]" so the JSONB column never holds a NULL or malformed payload.
func marshalStringList(in []string) ([]byte, error) {
	if len(in) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(in)
}

func unmarshalStringList(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	var out []string
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return out
}

func itemCurrencyAndMinor(item *menu.MenuItem) (money.Currency, int64) {
	currency := item.Currency
	if currency.IsZero() {
		currency = money.USD
	}
	priceMinor := money.FromMajor(item.Price, currency).Amount
	return currency, priceMinor
}
