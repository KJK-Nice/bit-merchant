package postgres

import (
	"database/sql"
	"errors"

	"bitmerchant/internal/domain"
)

// MembershipRepository implements domain.MembershipRepository for PostgreSQL.
type MembershipRepository struct {
	db *sql.DB
}

func NewMembershipRepository(db *sql.DB) *MembershipRepository {
	return &MembershipRepository{db: db}
}

func (r *MembershipRepository) Save(membership *domain.Membership) error {
	_, err := r.db.Exec(
		`INSERT INTO auth_memberships (id, user_id, restaurant_id, role, created_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (id) DO UPDATE
		 SET user_id = EXCLUDED.user_id,
		     restaurant_id = EXCLUDED.restaurant_id,
		     role = EXCLUDED.role`,
		string(membership.ID),
		string(membership.UserID),
		string(membership.RestaurantID),
		string(membership.Role),
		membership.CreatedAt,
	)
	return err
}

func (r *MembershipRepository) FindByUserID(userID domain.UserID) ([]*domain.Membership, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, restaurant_id, role, created_at
		   FROM auth_memberships
		  WHERE user_id = $1`,
		string(userID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []*domain.Membership
	for rows.Next() {
		membership, err := scanMembership(rows)
		if err != nil {
			return nil, err
		}
		memberships = append(memberships, membership)
	}
	return memberships, rows.Err()
}

func (r *MembershipRepository) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Membership, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, restaurant_id, role, created_at
		   FROM auth_memberships
		  WHERE restaurant_id = $1`,
		string(restaurantID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []*domain.Membership
	for rows.Next() {
		membership, err := scanMembership(rows)
		if err != nil {
			return nil, err
		}
		memberships = append(memberships, membership)
	}
	return memberships, rows.Err()
}

func (r *MembershipRepository) FindByUserAndRestaurant(userID domain.UserID, restaurantID domain.RestaurantID) (*domain.Membership, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, restaurant_id, role, created_at
		   FROM auth_memberships
		  WHERE user_id = $1 AND restaurant_id = $2
		  LIMIT 1`,
		string(userID),
		string(restaurantID),
	)

	return scanMembershipRow(row)
}

func (r *MembershipRepository) Delete(id domain.MembershipID) error {
	result, err := r.db.Exec(`DELETE FROM auth_memberships WHERE id = $1`, string(id))
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("membership not found")
	}
	return nil
}

func scanMembership(rows *sql.Rows) (*domain.Membership, error) {
	var (
		id           string
		userID       string
		restaurantID string
		role         string
		createdAt    sql.NullTime
	)
	if err := rows.Scan(&id, &userID, &restaurantID, &role, &createdAt); err != nil {
		return nil, err
	}

	return &domain.Membership{
		ID:           domain.MembershipID(id),
		UserID:       domain.UserID(userID),
		RestaurantID: domain.RestaurantID(restaurantID),
		Role:         domain.MemberRole(role),
		CreatedAt:    createdAt.Time,
	}, nil
}

func scanMembershipRow(row *sql.Row) (*domain.Membership, error) {
	var (
		id           string
		userID       string
		restaurantID string
		role         string
		createdAt    sql.NullTime
	)
	if err := row.Scan(&id, &userID, &restaurantID, &role, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("membership not found")
		}
		return nil, err
	}

	return &domain.Membership{
		ID:           domain.MembershipID(id),
		UserID:       domain.UserID(userID),
		RestaurantID: domain.RestaurantID(restaurantID),
		Role:         domain.MemberRole(role),
		CreatedAt:    createdAt.Time,
	}, nil
}
