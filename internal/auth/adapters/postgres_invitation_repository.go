package adapters

import (
	"database/sql"
	"errors"

	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/common"
)

type PostgresInvitationRepository struct {
	db *sql.DB
}

func NewPostgresInvitationRepository(db *sql.DB) *PostgresInvitationRepository {
	return &PostgresInvitationRepository{db: db}
}

func (r *PostgresInvitationRepository) Save(inv *invitation.Invitation) error {
	var usedByStr *string
	if inv.UsedByUserID != nil {
		s := string(*inv.UsedByUserID)
		usedByStr = &s
	}
	_, err := r.db.Exec(
		`INSERT INTO auth_invitations (id, restaurant_id, role, token, expires_at, used_at, used_by_user_id, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 ON CONFLICT (id) DO UPDATE SET restaurant_id=EXCLUDED.restaurant_id, role=EXCLUDED.role,
		   token=EXCLUDED.token, expires_at=EXCLUDED.expires_at, used_at=EXCLUDED.used_at, used_by_user_id=EXCLUDED.used_by_user_id`,
		string(inv.ID), string(inv.RestaurantID), string(inv.Role), inv.Token,
		inv.ExpiresAt, inv.UsedAt, usedByStr, inv.CreatedAt)
	return err
}

func (r *PostgresInvitationRepository) FindByToken(token string) (*invitation.Invitation, error) {
	row := r.db.QueryRow(
		`SELECT id, restaurant_id, role, token, expires_at, used_at, used_by_user_id, created_at
		 FROM auth_invitations WHERE token = $1 LIMIT 1`, token)
	return scanInvitation(row)
}

func (r *PostgresInvitationRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*invitation.Invitation, error) {
	rows, err := r.db.Query(
		`SELECT id, restaurant_id, role, token, expires_at, used_at, used_by_user_id, created_at
		 FROM auth_invitations WHERE restaurant_id = $1`, string(restaurantID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*invitation.Invitation
	for rows.Next() {
		var id, rid, role, token string
		var expiresAt, createdAt sql.NullTime
		var usedAt sql.NullTime
		var usedBy sql.NullString
		if err := rows.Scan(&id, &rid, &role, &token, &expiresAt, &usedAt, &usedBy, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, buildInvitation(id, rid, role, token, expiresAt, usedAt, usedBy, createdAt))
	}
	return result, rows.Err()
}

func (r *PostgresInvitationRepository) Update(inv *invitation.Invitation) error {
	var usedByStr *string
	if inv.UsedByUserID != nil {
		s := string(*inv.UsedByUserID)
		usedByStr = &s
	}
	result, err := r.db.Exec(
		`UPDATE auth_invitations SET restaurant_id=$2, role=$3, token=$4, expires_at=$5, used_at=$6, used_by_user_id=$7 WHERE id=$1`,
		string(inv.ID), string(inv.RestaurantID), string(inv.Role), inv.Token,
		inv.ExpiresAt, inv.UsedAt, usedByStr)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("invitation not found")
	}
	return nil
}

func scanInvitation(row *sql.Row) (*invitation.Invitation, error) {
	var id, rid, role, token string
	var expiresAt, createdAt sql.NullTime
	var usedAt sql.NullTime
	var usedBy sql.NullString
	if err := row.Scan(&id, &rid, &role, &token, &expiresAt, &usedAt, &usedBy, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invitation not found")
		}
		return nil, err
	}
	return buildInvitation(id, rid, role, token, expiresAt, usedAt, usedBy, createdAt), nil
}

func buildInvitation(id, rid, role, token string, expiresAt, usedAt sql.NullTime, usedBy sql.NullString, createdAt sql.NullTime) *invitation.Invitation {
	inv := &invitation.Invitation{
		ID: common.InvitationID(id), RestaurantID: common.RestaurantID(rid),
		Role: common.MemberRole(role), Token: token,
		ExpiresAt: expiresAt.Time, CreatedAt: createdAt.Time,
	}
	if usedAt.Valid {
		t := usedAt.Time
		inv.UsedAt = &t
	}
	if usedBy.Valid && usedBy.String != "" {
		uid := common.UserID(usedBy.String)
		inv.UsedByUserID = &uid
	}
	return inv
}
