package adapters

import (
	"database/sql"
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

const itemSelectCols = `id, category_id, restaurant_id, name, description, price, COALESCE(currency, 'USD'), COALESCE(price_minor, 0), photo_url, photo_original_url, is_available, display_order, created_at, updated_at, COALESCE(is_vegetarian, false), COALESCE(is_gluten_free, false), COALESCE(is_spicy, false)`

func (r *PostgresItemRepository) Save(item *menu.MenuItem) error {
	currency, priceMinor := itemCurrencyAndMinor(item)
	_, err := r.db.Exec(
		`INSERT INTO menu_items (id, category_id, restaurant_id, name, description, price, currency, price_minor, photo_url, photo_original_url, is_available, display_order, created_at, updated_at, is_vegetarian, is_gluten_free, is_spicy)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		 ON CONFLICT (id) DO UPDATE
		 SET category_id = EXCLUDED.category_id, name = EXCLUDED.name,
		     description = EXCLUDED.description, price = EXCLUDED.price,
		     currency = EXCLUDED.currency, price_minor = EXCLUDED.price_minor,
		     photo_url = EXCLUDED.photo_url, photo_original_url = EXCLUDED.photo_original_url,
		     is_available = EXCLUDED.is_available, display_order = EXCLUDED.display_order,
		     is_vegetarian = EXCLUDED.is_vegetarian, is_gluten_free = EXCLUDED.is_gluten_free,
		     is_spicy = EXCLUDED.is_spicy, updated_at = EXCLUDED.updated_at`,
		string(item.ID), string(item.CategoryID), string(item.RestaurantID),
		item.Name, item.Description, item.Price,
		currency.Code, priceMinor,
		item.PhotoURL, item.PhotoOriginalURL, item.IsAvailable, item.DisplayOrder,
		item.CreatedAt, item.UpdatedAt,
		item.IsVegetarian, item.IsGlutenFree, item.IsSpicy)
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
	result, err := r.db.Exec(
		`UPDATE menu_items SET category_id=$2, name=$3, description=$4, price=$5, currency=$6, price_minor=$7, photo_url=$8, photo_original_url=$9, is_available=$10, display_order=$11, is_vegetarian=$12, is_gluten_free=$13, is_spicy=$14, updated_at=$15 WHERE id=$1`,
		string(item.ID), string(item.CategoryID), item.Name, item.Description, item.Price,
		currency.Code, priceMinor,
		item.PhotoURL, item.PhotoOriginalURL, item.IsAvailable, item.DisplayOrder,
		item.IsVegetarian, item.IsGlutenFree, item.IsSpicy, item.UpdatedAt)
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
	id, catID, restID, name string
	description             sql.NullString
	price                   float64
	currencyCode            string
	priceMinor              int64
	photoURL, photoOrigURL  sql.NullString
	isAvailable             bool
	displayOrder            int
	createdAt, updatedAt    time.Time
	isVegetarian            bool
	isGlutenFree            bool
	isSpicy                 bool
}

func (f *itemRowFields) toMenuItem() *menu.MenuItem {
	currency, err := money.Parse(f.currencyCode)
	if err != nil {
		currency = money.USD
	}
	return &menu.MenuItem{
		ID: common.ItemID(f.id), CategoryID: common.CategoryID(f.catID),
		RestaurantID: common.RestaurantID(f.restID), Name: f.name,
		Description: f.description.String, Price: f.price, Currency: currency,
		PhotoURL: f.photoURL.String, PhotoOriginalURL: f.photoOrigURL.String,
		IsAvailable: f.isAvailable, DisplayOrder: f.displayOrder,
		IsVegetarian: f.isVegetarian, IsGlutenFree: f.isGlutenFree, IsSpicy: f.isSpicy,
		CreatedAt: f.createdAt, UpdatedAt: f.updatedAt,
	}
}

func scanItem(row *sql.Row) (*menu.MenuItem, error) {
	var f itemRowFields
	if err := row.Scan(&f.id, &f.catID, &f.restID, &f.name, &f.description, &f.price, &f.currencyCode, &f.priceMinor, &f.photoURL, &f.photoOrigURL, &f.isAvailable, &f.displayOrder, &f.createdAt, &f.updatedAt, &f.isVegetarian, &f.isGlutenFree, &f.isSpicy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("menu item not found")
		}
		return nil, err
	}
	return f.toMenuItem(), nil
}

func scanItemRows(rows *sql.Rows) (*menu.MenuItem, error) {
	var f itemRowFields
	if err := rows.Scan(&f.id, &f.catID, &f.restID, &f.name, &f.description, &f.price, &f.currencyCode, &f.priceMinor, &f.photoURL, &f.photoOrigURL, &f.isAvailable, &f.displayOrder, &f.createdAt, &f.updatedAt, &f.isVegetarian, &f.isGlutenFree, &f.isSpicy); err != nil {
		return nil, err
	}
	return f.toMenuItem(), nil
}

func itemCurrencyAndMinor(item *menu.MenuItem) (money.Currency, int64) {
	currency := item.Currency
	if currency.IsZero() {
		currency = money.USD
	}
	priceMinor := money.FromMajor(item.Price, currency).Amount
	return currency, priceMinor
}
