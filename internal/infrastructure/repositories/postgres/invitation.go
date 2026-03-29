package postgres

import (
	"database/sql"
	"errors"
	"time"

	"bitmerchant/internal/domain"
)

// InvitationRepository implements domain.InvitationRepository for PostgreSQL.
type InvitationRepository struct {
	db *sql.DB
}

func NewInvitationRepository(db *sql.DB) *InvitationRepository {
	return &InvitationRepository{db: db}
}

func (r *InvitationRepository) Save(invitation *domain.Invitation) error {
	_, err := r.db.Exec(
		`INSERT INTO auth_invitations (id, restaurant_id, role, token, expires_at, used_at, used_by_user_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (id) DO UPDATE
		 SET restaurant_id = EXCLUDED.restaurant_id,
		     role = EXCLUDED.role,
		     token = EXCLUDED.token,
		     expires_at = EXCLUDED.expires_at,
		     used_at = EXCLUDED.used_at,
		     used_by_user_id = EXCLUDED.used_by_user_id`,
		string(invitation.ID),
		string(invitation.RestaurantID),
		string(invitation.Role),
		invitation.Token,
		invitation.ExpiresAt,
		invitation.UsedAt,
		stringPtr(invitation.UsedByUserID),
		invitation.CreatedAt,
	)
	return err
}

func (r *InvitationRepository) FindByToken(token string) (*domain.Invitation, error) {
	row := r.db.QueryRow(
		`SELECT id, restaurant_id, role, token, expires_at, used_at, used_by_user_id, created_at
		   FROM auth_invitations
		  WHERE token = $1
		  LIMIT 1`,
		token,
	)
	return scanInvitationRow(row)
}

func (r *InvitationRepository) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Invitation, error) {
	rows, err := r.db.Query(
		`SELECT id, restaurant_id, role, token, expires_at, used_at, used_by_user_id, created_at
		   FROM auth_invitations
		  WHERE restaurant_id = $1`,
		string(restaurantID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*domain.Invitation
	for rows.Next() {
		invitation, err := scanInvitation(rows)
		if err != nil {
			return nil, err
		}
		invitations = append(invitations, invitation)
	}
	return invitations, rows.Err()
}

func (r *InvitationRepository) Update(invitation *domain.Invitation) error {
	result, err := r.db.Exec(
		`UPDATE auth_invitations
		    SET restaurant_id = $2,
		        role = $3,
		        token = $4,
		        expires_at = $5,
		        used_at = $6,
		        used_by_user_id = $7
		  WHERE id = $1`,
		string(invitation.ID),
		string(invitation.RestaurantID),
		string(invitation.Role),
		invitation.Token,
		invitation.ExpiresAt,
		invitation.UsedAt,
		stringPtr(invitation.UsedByUserID),
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("invitation not found")
	}
	return nil
}

func scanInvitation(rows *sql.Rows) (*domain.Invitation, error) {
	var (
		id           string
		restaurantID string
		role         string
		token        string
		expiresAt    sql.NullTime
		usedAt       sql.NullTime
		usedByUserID sql.NullString
		createdAt    sql.NullTime
	)
	if err := rows.Scan(&id, &restaurantID, &role, &token, &expiresAt, &usedAt, &usedByUserID, &createdAt); err != nil {
		return nil, err
	}

	var usedAtPtr *time.Time
	if usedAt.Valid {
		usedAtValue := usedAt.Time
		usedAtPtr = &usedAtValue
	}

	var usedByPtr *domain.UserID
	if usedByUserID.Valid && usedByUserID.String != "" {
		uid := domain.UserID(usedByUserID.String)
		usedByPtr = &uid
	}

	return &domain.Invitation{
		ID:           domain.InvitationID(id),
		RestaurantID: domain.RestaurantID(restaurantID),
		Role:         domain.MemberRole(role),
		Token:        token,
		ExpiresAt:    expiresAt.Time,
		UsedAt:       usedAtPtr,
		UsedByUserID: usedByPtr,
		CreatedAt:    createdAt.Time,
	}, nil
}

func scanInvitationRow(row *sql.Row) (*domain.Invitation, error) {
	var (
		id           string
		restaurantID string
		role         string
		token        string
		expiresAt    sql.NullTime
		usedAt       sql.NullTime
		usedByUserID sql.NullString
		createdAt    sql.NullTime
	)
	if err := row.Scan(&id, &restaurantID, &role, &token, &expiresAt, &usedAt, &usedByUserID, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invitation not found")
		}
		return nil, err
	}

	var usedAtPtr *time.Time
	if usedAt.Valid {
		usedAtValue := usedAt.Time
		usedAtPtr = &usedAtValue
	}

	var usedByPtr *domain.UserID
	if usedByUserID.Valid && usedByUserID.String != "" {
		uid := domain.UserID(usedByUserID.String)
		usedByPtr = &uid
	}

	return &domain.Invitation{
		ID:           domain.InvitationID(id),
		RestaurantID: domain.RestaurantID(restaurantID),
		Role:         domain.MemberRole(role),
		Token:        token,
		ExpiresAt:    expiresAt.Time,
		UsedAt:       usedAtPtr,
		UsedByUserID: usedByPtr,
		CreatedAt:    createdAt.Time,
	}, nil
}

func stringPtr(userID *domain.UserID) *string {
	if userID == nil {
		return nil
	}
	s := string(*userID)
	return &s
}
