package adapters

import (
	"database/sql"
	"errors"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

type PostgresCategoryRepository struct {
	db *sql.DB
}

func NewPostgresCategoryRepository(db *sql.DB) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{db: db}
}

func (r *PostgresCategoryRepository) Save(category *menu.MenuCategory) error {
	_, err := r.db.Exec(
		`INSERT INTO menu_categories (id, restaurant_id, name, display_order, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (id) DO UPDATE
		 SET name = EXCLUDED.name, display_order = EXCLUDED.display_order,
		     is_active = EXCLUDED.is_active, updated_at = EXCLUDED.updated_at`,
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

func (r *PostgresCategoryRepository) FindByID(id common.CategoryID) (*menu.MenuCategory, error) {
	row := r.db.QueryRow(
		`SELECT id, restaurant_id, name, display_order, is_active, created_at, updated_at
		 FROM menu_categories WHERE id = $1`, string(id))
	return scanCategory(row)
}

func (r *PostgresCategoryRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*menu.MenuCategory, error) {
	rows, err := r.db.Query(
		`SELECT id, restaurant_id, name, display_order, is_active, created_at, updated_at
		 FROM menu_categories WHERE restaurant_id = $1`, string(restaurantID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*menu.MenuCategory
	for rows.Next() {
		cat, err := scanCategoryRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, cat)
	}
	return result, rows.Err()
}

func (r *PostgresCategoryRepository) Update(category *menu.MenuCategory) error {
	result, err := r.db.Exec(
		`UPDATE menu_categories SET name=$2, display_order=$3, is_active=$4, updated_at=$5
		 WHERE id=$1`,
		string(category.ID), category.Name, category.DisplayOrder, category.IsActive, category.UpdatedAt)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("menu category not found")
	}
	return nil
}

func (r *PostgresCategoryRepository) Delete(id common.CategoryID) error {
	result, err := r.db.Exec(`DELETE FROM menu_categories WHERE id = $1`, string(id))
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("menu category not found")
	}
	return nil
}

func scanCategory(row *sql.Row) (*menu.MenuCategory, error) {
	var (
		id, restID, name string
		displayOrder     int
		isActive         bool
		createdAt, updatedAt time.Time
	)
	if err := row.Scan(&id, &restID, &name, &displayOrder, &isActive, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("menu category not found")
		}
		return nil, err
	}
	return &menu.MenuCategory{
		ID: common.CategoryID(id), RestaurantID: common.RestaurantID(restID),
		Name: name, DisplayOrder: displayOrder, IsActive: isActive,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

func scanCategoryRows(rows *sql.Rows) (*menu.MenuCategory, error) {
	var (
		id, restID, name string
		displayOrder     int
		isActive         bool
		createdAt, updatedAt time.Time
	)
	if err := rows.Scan(&id, &restID, &name, &displayOrder, &isActive, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	return &menu.MenuCategory{
		ID: common.CategoryID(id), RestaurantID: common.RestaurantID(restID),
		Name: name, DisplayOrder: displayOrder, IsActive: isActive,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}
