package adapters

import (
	"database/sql"
	"errors"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

type PostgresItemRepository struct {
	db *sql.DB
}

func NewPostgresItemRepository(db *sql.DB) *PostgresItemRepository {
	return &PostgresItemRepository{db: db}
}

func (r *PostgresItemRepository) Save(item *menu.MenuItem) error {
	_, err := r.db.Exec(
		`INSERT INTO menu_items (id, category_id, restaurant_id, name, description, price, photo_url, photo_original_url, is_available, display_order, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 ON CONFLICT (id) DO UPDATE
		 SET category_id = EXCLUDED.category_id, name = EXCLUDED.name,
		     description = EXCLUDED.description, price = EXCLUDED.price,
		     photo_url = EXCLUDED.photo_url, photo_original_url = EXCLUDED.photo_original_url,
		     is_available = EXCLUDED.is_available, display_order = EXCLUDED.display_order, updated_at = EXCLUDED.updated_at`,
		string(item.ID), string(item.CategoryID), string(item.RestaurantID),
		item.Name, item.Description, item.Price,
		item.PhotoURL, item.PhotoOriginalURL, item.IsAvailable, item.DisplayOrder,
		item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *PostgresItemRepository) FindByID(id common.ItemID) (*menu.MenuItem, error) {
	row := r.db.QueryRow(
		`SELECT id, category_id, restaurant_id, name, description, price, photo_url, photo_original_url, is_available, display_order, created_at, updated_at
		 FROM menu_items WHERE id = $1`, string(id))
	return scanItem(row)
}

func (r *PostgresItemRepository) FindByCategoryID(categoryID common.CategoryID) ([]*menu.MenuItem, error) {
	return r.queryItems(`SELECT id, category_id, restaurant_id, name, description, price, photo_url, photo_original_url, is_available, display_order, created_at, updated_at FROM menu_items WHERE category_id = $1`, string(categoryID))
}

func (r *PostgresItemRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*menu.MenuItem, error) {
	return r.queryItems(`SELECT id, category_id, restaurant_id, name, description, price, photo_url, photo_original_url, is_available, display_order, created_at, updated_at FROM menu_items WHERE restaurant_id = $1`, string(restaurantID))
}

func (r *PostgresItemRepository) FindAvailableByRestaurantID(restaurantID common.RestaurantID) ([]*menu.MenuItem, error) {
	return r.queryItems(`SELECT id, category_id, restaurant_id, name, description, price, photo_url, photo_original_url, is_available, display_order, created_at, updated_at FROM menu_items WHERE restaurant_id = $1 AND is_available = true`, string(restaurantID))
}

func (r *PostgresItemRepository) Update(item *menu.MenuItem) error {
	result, err := r.db.Exec(
		`UPDATE menu_items SET category_id=$2, name=$3, description=$4, price=$5, photo_url=$6, photo_original_url=$7, is_available=$8, display_order=$9, updated_at=$10 WHERE id=$1`,
		string(item.ID), string(item.CategoryID), item.Name, item.Description, item.Price,
		item.PhotoURL, item.PhotoOriginalURL, item.IsAvailable, item.DisplayOrder, item.UpdatedAt)
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

	rows, err := tx.Query(
		`SELECT id FROM menu_items WHERE restaurant_id = $1 AND category_id = $2`,
		string(restaurantID), string(categoryID))
	if err != nil {
		return err
	}
	var existing []common.ItemID
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		existing = append(existing, common.ItemID(id))
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
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
	return tx.Commit()
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

func scanItem(row *sql.Row) (*menu.MenuItem, error) {
	var (
		id, catID, restID, name string
		description             sql.NullString
		price                   float64
		photoURL, photoOrigURL  sql.NullString
		isAvailable             bool
		displayOrder            int
		createdAt, updatedAt    time.Time
	)
	if err := row.Scan(&id, &catID, &restID, &name, &description, &price, &photoURL, &photoOrigURL, &isAvailable, &displayOrder, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("menu item not found")
		}
		return nil, err
	}
	return &menu.MenuItem{
		ID: common.ItemID(id), CategoryID: common.CategoryID(catID),
		RestaurantID: common.RestaurantID(restID), Name: name,
		Description: description.String, Price: price,
		PhotoURL: photoURL.String, PhotoOriginalURL: photoOrigURL.String,
		IsAvailable: isAvailable, DisplayOrder: displayOrder, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

func scanItemRows(rows *sql.Rows) (*menu.MenuItem, error) {
	var (
		id, catID, restID, name string
		description             sql.NullString
		price                   float64
		photoURL, photoOrigURL  sql.NullString
		isAvailable             bool
		displayOrder            int
		createdAt, updatedAt    time.Time
	)
	if err := rows.Scan(&id, &catID, &restID, &name, &description, &price, &photoURL, &photoOrigURL, &isAvailable, &displayOrder, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	return &menu.MenuItem{
		ID: common.ItemID(id), CategoryID: common.CategoryID(catID),
		RestaurantID: common.RestaurantID(restID), Name: name,
		Description: description.String, Price: price,
		PhotoURL: photoURL.String, PhotoOriginalURL: photoOrigURL.String,
		IsAvailable: isAvailable, DisplayOrder: displayOrder, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}
