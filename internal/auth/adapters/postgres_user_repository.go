package adapters

import (
	"database/sql"
	"encoding/json"
	"errors"

	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"

	"github.com/go-webauthn/webauthn/webauthn"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Save(u *user.User) error {
	credJSON, err := json.Marshal(u.Credentials)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(
		`INSERT INTO auth_users (id, display_name, credentials_json, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (id) DO UPDATE SET display_name=EXCLUDED.display_name, credentials_json=EXCLUDED.credentials_json, updated_at=EXCLUDED.updated_at`,
		string(u.ID), u.DisplayName, credJSON, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *PostgresUserRepository) FindByID(id common.UserID) (*user.User, error) {
	var uid, displayName string
	var credJSON []byte
	var createdAt, updatedAt sql.NullTime
	err := r.db.QueryRow(
		`SELECT id, display_name, credentials_json, created_at, updated_at FROM auth_users WHERE id = $1`,
		string(id)).Scan(&uid, &displayName, &credJSON, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	var creds []webauthn.Credential
	if len(credJSON) > 0 {
		_ = json.Unmarshal(credJSON, &creds)
	}
	return &user.User{
		ID: common.UserID(uid), DisplayName: displayName,
		Credentials: creds, CreatedAt: createdAt.Time, UpdatedAt: updatedAt.Time,
	}, nil
}

func (r *PostgresUserRepository) FindByCredentialID(credentialID []byte) (*user.User, *webauthn.Credential, error) {
	rows, err := r.db.Query(`SELECT id, display_name, credentials_json, created_at, updated_at FROM auth_users`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var uid, displayName string
		var credJSON []byte
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&uid, &displayName, &credJSON, &createdAt, &updatedAt); err != nil {
			return nil, nil, err
		}
		var creds []webauthn.Credential
		if len(credJSON) > 0 {
			_ = json.Unmarshal(credJSON, &creds)
		}
		for i := range creds {
			if string(creds[i].ID) == string(credentialID) {
				u := &user.User{
					ID: common.UserID(uid), DisplayName: displayName,
					Credentials: creds, CreatedAt: createdAt.Time, UpdatedAt: updatedAt.Time,
				}
				return u, &creds[i], nil
			}
		}
	}
	return nil, nil, errors.New("credential not found")
}

func (r *PostgresUserRepository) Update(u *user.User) error {
	credJSON, err := json.Marshal(u.Credentials)
	if err != nil {
		return err
	}
	result, err := r.db.Exec(
		`UPDATE auth_users SET display_name=$2, credentials_json=$3, updated_at=$4 WHERE id=$1`,
		string(u.ID), u.DisplayName, credJSON, u.UpdatedAt)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("user not found")
	}
	return nil
}
