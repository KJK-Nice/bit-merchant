package adapters

import (
	"database/sql"
	"errors"

	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"
)

type PostgresMembershipRepository struct {
	db *sql.DB
}

func NewPostgresMembershipRepository(db *sql.DB) *PostgresMembershipRepository {
	return &PostgresMembershipRepository{db: db}
}

func (r *PostgresMembershipRepository) Save(m *membership.Membership) error {
	_, err := r.db.Exec(
		`INSERT INTO auth_memberships (id, user_id, restaurant_id, role, created_at)
		 VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (id) DO UPDATE SET user_id=EXCLUDED.user_id, restaurant_id=EXCLUDED.restaurant_id, role=EXCLUDED.role`,
		string(m.ID), string(m.UserID), string(m.RestaurantID), string(m.Role), m.CreatedAt)
	return err
}

func (r *PostgresMembershipRepository) FindByUserID(userID common.UserID) ([]*membership.Membership, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, restaurant_id, role, created_at FROM auth_memberships WHERE user_id = $1`, string(userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemberships(rows)
}

func (r *PostgresMembershipRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*membership.Membership, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, restaurant_id, role, created_at FROM auth_memberships WHERE restaurant_id = $1`, string(restaurantID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemberships(rows)
}

func (r *PostgresMembershipRepository) FindByUserAndRestaurant(userID common.UserID, restaurantID common.RestaurantID) (*membership.Membership, error) {
	var id, uid, rid, role string
	var createdAt sql.NullTime
	err := r.db.QueryRow(
		`SELECT id, user_id, restaurant_id, role, created_at FROM auth_memberships WHERE user_id = $1 AND restaurant_id = $2 LIMIT 1`,
		string(userID), string(restaurantID)).Scan(&id, &uid, &rid, &role, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("membership not found")
		}
		return nil, err
	}
	return &membership.Membership{
		ID: common.MembershipID(id), UserID: common.UserID(uid),
		RestaurantID: common.RestaurantID(rid), Role: common.MemberRole(role),
		CreatedAt: createdAt.Time,
	}, nil
}

func (r *PostgresMembershipRepository) Delete(id common.MembershipID) error {
	result, err := r.db.Exec(`DELETE FROM auth_memberships WHERE id = $1`, string(id))
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("membership not found")
	}
	return nil
}

func scanMemberships(rows *sql.Rows) ([]*membership.Membership, error) {
	var result []*membership.Membership
	for rows.Next() {
		var id, uid, rid, role string
		var createdAt sql.NullTime
		if err := rows.Scan(&id, &uid, &rid, &role, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, &membership.Membership{
			ID: common.MembershipID(id), UserID: common.UserID(uid),
			RestaurantID: common.RestaurantID(rid), Role: common.MemberRole(role),
			CreatedAt: createdAt.Time,
		})
	}
	return result, rows.Err()
}
