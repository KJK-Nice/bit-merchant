package postgres

import (
	"database/sql"
	"errors"

	"bitmerchant/internal/domain"
)

// MenuCategoryRepository implements domain.MenuCategoryRepository for PostgreSQL.
type MenuCategoryRepository struct {
	db *sql.DB
}

func NewMenuCategoryRepository(db *sql.DB) *MenuCategoryRepository {
	return &MenuCategoryRepository{db: db}
}

func (r *MenuCategoryRepository) Save(category *domain.MenuCategory) error {
	_, err := r.db.Exec(
		`INSERT INTO menu_categories (id, restaurant_id, name, display_order, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (id) DO UPDATE
		 SET restaurant_id = EXCLUDED.restaurant_id,
		     name = EXCLUDED.name,
		     display_order = EXCLUDED.display_order,
		     is_active = EXCLUDED.is_active,
		     updated_at = EXCLUDED.updated_at`,
		string(category.ID),
		string(category.RestaurantID),
		category.Name,
		category.DisplayOrder,
		category.IsActive,
		category.CreatedAt,
		category.UpdatedAt,
	)
	return err
}

func (r *MenuCategoryRepository) FindByID(id domain.CategoryID) (*domain.MenuCategory, error) {
	row := r.db.QueryRow(
		`SELECT id, restaurant_id, name, display_order, is_active, created_at, updated_at
		   FROM menu_categories
		  WHERE id = $1`,
		string(id),
	)
	return scanMenuCategoryRow(row)
}

func (r *MenuCategoryRepository) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.MenuCategory, error) {
	rows, err := r.db.Query(
		`SELECT id, restaurant_id, name, display_order, is_active, created_at, updated_at
		   FROM menu_categories
		  WHERE restaurant_id = $1
		  ORDER BY display_order ASC, created_at ASC`,
		string(restaurantID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*domain.MenuCategory
	for rows.Next() {
		category, scanErr := scanMenuCategory(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (r *MenuCategoryRepository) Update(category *domain.MenuCategory) error {
	result, err := r.db.Exec(
		`UPDATE menu_categories
		    SET restaurant_id = $2,
		        name = $3,
		        display_order = $4,
		        is_active = $5,
		        updated_at = $6
		  WHERE id = $1`,
		string(category.ID),
		string(category.RestaurantID),
		category.Name,
		category.DisplayOrder,
		category.IsActive,
		category.UpdatedAt,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("menu category not found")
	}
	return nil
}

func (r *MenuCategoryRepository) Delete(id domain.CategoryID) error {
	result, err := r.db.Exec(`DELETE FROM menu_categories WHERE id = $1`, string(id))
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("menu category not found")
	}
	return nil
}

func scanMenuCategory(rows *sql.Rows) (*domain.MenuCategory, error) {
	var (
		id           string
		restaurantID string
		name         string
		displayOrder int
		isActive     bool
		createdAt    sql.NullTime
		updatedAt    sql.NullTime
	)
	if err := rows.Scan(&id, &restaurantID, &name, &displayOrder, &isActive, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	return &domain.MenuCategory{
		ID:           domain.CategoryID(id),
		RestaurantID: domain.RestaurantID(restaurantID),
		Name:         name,
		DisplayOrder: displayOrder,
		IsActive:     isActive,
		CreatedAt:    createdAt.Time,
		UpdatedAt:    updatedAt.Time,
	}, nil
}

func scanMenuCategoryRow(row *sql.Row) (*domain.MenuCategory, error) {
	var (
		id           string
		restaurantID string
		name         string
		displayOrder int
		isActive     bool
		createdAt    sql.NullTime
		updatedAt    sql.NullTime
	)
	if err := row.Scan(&id, &restaurantID, &name, &displayOrder, &isActive, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("menu category not found")
		}
		return nil, err
	}

	return &domain.MenuCategory{
		ID:           domain.CategoryID(id),
		RestaurantID: domain.RestaurantID(restaurantID),
		Name:         name,
		DisplayOrder: displayOrder,
		IsActive:     isActive,
		CreatedAt:    createdAt.Time,
		UpdatedAt:    updatedAt.Time,
	}, nil
}
