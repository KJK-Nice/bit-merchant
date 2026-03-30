package postgres

import (
	"database/sql"
	"errors"

	"bitmerchant/internal/domain"
)

// MenuItemRepository implements domain.MenuItemRepository for PostgreSQL.
type MenuItemRepository struct {
	db *sql.DB
}

func NewMenuItemRepository(db *sql.DB) *MenuItemRepository {
	return &MenuItemRepository{db: db}
}

func (r *MenuItemRepository) Save(item *domain.MenuItem) error {
	_, err := r.db.Exec(
		`INSERT INTO menu_items (id, category_id, restaurant_id, name, description, price, photo_url, photo_original_url, is_available, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (id) DO UPDATE
		 SET category_id = EXCLUDED.category_id,
		     restaurant_id = EXCLUDED.restaurant_id,
		     name = EXCLUDED.name,
		     description = EXCLUDED.description,
		     price = EXCLUDED.price,
		     photo_url = EXCLUDED.photo_url,
		     photo_original_url = EXCLUDED.photo_original_url,
		     is_available = EXCLUDED.is_available,
		     updated_at = EXCLUDED.updated_at`,
		string(item.ID),
		string(item.CategoryID),
		string(item.RestaurantID),
		item.Name,
		item.Description,
		item.Price,
		item.PhotoURL,
		item.PhotoOriginalURL,
		item.IsAvailable,
		item.CreatedAt,
		item.UpdatedAt,
	)
	return err
}

func (r *MenuItemRepository) FindByID(id domain.ItemID) (*domain.MenuItem, error) {
	row := r.db.QueryRow(
		`SELECT id, category_id, restaurant_id, name, description, price, photo_url, photo_original_url, is_available, created_at, updated_at
		   FROM menu_items
		  WHERE id = $1`,
		string(id),
	)
	return scanMenuItemRow(row)
}

func (r *MenuItemRepository) FindByCategoryID(categoryID domain.CategoryID) ([]*domain.MenuItem, error) {
	rows, err := r.db.Query(
		`SELECT id, category_id, restaurant_id, name, description, price, photo_url, photo_original_url, is_available, created_at, updated_at
		   FROM menu_items
		  WHERE category_id = $1
		  ORDER BY created_at ASC`,
		string(categoryID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMenuItemRows(rows)
}

func (r *MenuItemRepository) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.MenuItem, error) {
	rows, err := r.db.Query(
		`SELECT id, category_id, restaurant_id, name, description, price, photo_url, photo_original_url, is_available, created_at, updated_at
		   FROM menu_items
		  WHERE restaurant_id = $1
		  ORDER BY created_at ASC`,
		string(restaurantID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMenuItemRows(rows)
}

func (r *MenuItemRepository) FindAvailableByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.MenuItem, error) {
	rows, err := r.db.Query(
		`SELECT id, category_id, restaurant_id, name, description, price, photo_url, photo_original_url, is_available, created_at, updated_at
		   FROM menu_items
		  WHERE restaurant_id = $1 AND is_available = TRUE
		  ORDER BY created_at ASC`,
		string(restaurantID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMenuItemRows(rows)
}

func (r *MenuItemRepository) Update(item *domain.MenuItem) error {
	result, err := r.db.Exec(
		`UPDATE menu_items
		    SET category_id = $2,
		        restaurant_id = $3,
		        name = $4,
		        description = $5,
		        price = $6,
		        photo_url = $7,
		        photo_original_url = $8,
		        is_available = $9,
		        updated_at = $10
		  WHERE id = $1`,
		string(item.ID),
		string(item.CategoryID),
		string(item.RestaurantID),
		item.Name,
		item.Description,
		item.Price,
		item.PhotoURL,
		item.PhotoOriginalURL,
		item.IsAvailable,
		item.UpdatedAt,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("menu item not found")
	}
	return nil
}

func (r *MenuItemRepository) Delete(id domain.ItemID) error {
	result, err := r.db.Exec(`DELETE FROM menu_items WHERE id = $1`, string(id))
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("menu item not found")
	}
	return nil
}

func (r *MenuItemRepository) CountByRestaurantID(restaurantID domain.RestaurantID) (int, error) {
	row := r.db.QueryRow(
		`SELECT COUNT(*)
		   FROM menu_items
		  WHERE restaurant_id = $1 AND photo_url <> ''`,
		string(restaurantID),
	)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func scanMenuItemRows(rows *sql.Rows) ([]*domain.MenuItem, error) {
	var items []*domain.MenuItem
	for rows.Next() {
		item, err := scanMenuItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanMenuItem(rows *sql.Rows) (*domain.MenuItem, error) {
	var (
		id               string
		categoryID       string
		restaurantID     string
		name             string
		description      string
		price            float64
		photoURL         string
		photoOriginalURL string
		isAvailable      bool
		createdAt        sql.NullTime
		updatedAt        sql.NullTime
	)
	if err := rows.Scan(
		&id,
		&categoryID,
		&restaurantID,
		&name,
		&description,
		&price,
		&photoURL,
		&photoOriginalURL,
		&isAvailable,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, err
	}

	return &domain.MenuItem{
		ID:               domain.ItemID(id),
		CategoryID:       domain.CategoryID(categoryID),
		RestaurantID:     domain.RestaurantID(restaurantID),
		Name:             name,
		Description:      description,
		Price:            price,
		PhotoURL:         photoURL,
		PhotoOriginalURL: photoOriginalURL,
		IsAvailable:      isAvailable,
		CreatedAt:        createdAt.Time,
		UpdatedAt:        updatedAt.Time,
	}, nil
}

func scanMenuItemRow(row *sql.Row) (*domain.MenuItem, error) {
	var (
		id               string
		categoryID       string
		restaurantID     string
		name             string
		description      string
		price            float64
		photoURL         string
		photoOriginalURL string
		isAvailable      bool
		createdAt        sql.NullTime
		updatedAt        sql.NullTime
	)
	if err := row.Scan(
		&id,
		&categoryID,
		&restaurantID,
		&name,
		&description,
		&price,
		&photoURL,
		&photoOriginalURL,
		&isAvailable,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("menu item not found")
		}
		return nil, err
	}

	return &domain.MenuItem{
		ID:               domain.ItemID(id),
		CategoryID:       domain.CategoryID(categoryID),
		RestaurantID:     domain.RestaurantID(restaurantID),
		Name:             name,
		Description:      description,
		Price:            price,
		PhotoURL:         photoURL,
		PhotoOriginalURL: photoOriginalURL,
		IsAvailable:      isAvailable,
		CreatedAt:        createdAt.Time,
		UpdatedAt:        updatedAt.Time,
	}, nil
}
