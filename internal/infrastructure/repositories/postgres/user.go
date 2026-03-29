package postgres

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"

	"bitmerchant/internal/domain"

	"github.com/go-webauthn/webauthn/webauthn"
)

// UserRepository implements domain.UserRepository for PostgreSQL.
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(user *domain.User) error {
	credentialsJSON, err := json.Marshal(user.Credentials)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO auth_users (id, display_name, credentials_json, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (id) DO UPDATE
		 SET display_name = EXCLUDED.display_name,
		     credentials_json = EXCLUDED.credentials_json,
		     updated_at = EXCLUDED.updated_at`,
		string(user.ID),
		user.DisplayName,
		credentialsJSON,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *UserRepository) FindByID(id domain.UserID) (*domain.User, error) {
	row := r.db.QueryRow(
		`SELECT id, display_name, credentials_json, created_at, updated_at
		   FROM auth_users
		  WHERE id = $1`,
		string(id),
	)

	var (
		userID          string
		displayName     string
		credentialsJSON []byte
		createdAt       sql.NullTime
		updatedAt       sql.NullTime
	)
	if err := row.Scan(&userID, &displayName, &credentialsJSON, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	var credentials []webauthn.Credential
	if len(credentialsJSON) > 0 {
		if err := json.Unmarshal(credentialsJSON, &credentials); err != nil {
			return nil, err
		}
	}

	return &domain.User{
		ID:          domain.UserID(userID),
		DisplayName: displayName,
		Credentials: credentials,
		CreatedAt:   createdAt.Time,
		UpdatedAt:   updatedAt.Time,
	}, nil
}

func (r *UserRepository) FindByCredentialID(credentialID []byte) (*domain.User, *webauthn.Credential, error) {
	rows, err := r.db.Query(
		`SELECT id, display_name, credentials_json, created_at, updated_at
		   FROM auth_users`,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			userID          string
			displayName     string
			credentialsJSON []byte
			createdAt       sql.NullTime
			updatedAt       sql.NullTime
		)
		if err := rows.Scan(&userID, &displayName, &credentialsJSON, &createdAt, &updatedAt); err != nil {
			return nil, nil, err
		}

		var credentials []webauthn.Credential
		if len(credentialsJSON) > 0 {
			if err := json.Unmarshal(credentialsJSON, &credentials); err != nil {
				return nil, nil, err
			}
		}

		for i := range credentials {
			if bytes.Equal(credentials[i].ID, credentialID) {
				user := &domain.User{
					ID:          domain.UserID(userID),
					DisplayName: displayName,
					Credentials: credentials,
					CreatedAt:   createdAt.Time,
					UpdatedAt:   updatedAt.Time,
				}
				return user, &credentials[i], nil
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return nil, nil, errors.New("credential not found")
}

func (r *UserRepository) Update(user *domain.User) error {
	credentialsJSON, err := json.Marshal(user.Credentials)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(
		`UPDATE auth_users
		    SET display_name = $2,
		        credentials_json = $3,
		        updated_at = $4
		  WHERE id = $1`,
		string(user.ID),
		user.DisplayName,
		credentialsJSON,
		user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("user not found")
	}

	return nil
}
