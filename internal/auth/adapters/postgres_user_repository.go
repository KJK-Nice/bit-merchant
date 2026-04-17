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
	emailVal := nullableString(u.Email)
	hashVal := nullableString(u.PasswordHash)
	_, err = r.db.Exec(
		`INSERT INTO auth_users (id, display_name, email, password_hash, credentials_json, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (id) DO UPDATE SET display_name=EXCLUDED.display_name, email=EXCLUDED.email, password_hash=EXCLUDED.password_hash, credentials_json=EXCLUDED.credentials_json, updated_at=EXCLUDED.updated_at`,
		string(u.ID), u.DisplayName, emailVal, hashVal, credJSON, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *PostgresUserRepository) FindByID(id common.UserID) (*user.User, error) {
	var uid, displayName string
	var email, passwordHash sql.NullString
	var credJSON []byte
	var createdAt, updatedAt sql.NullTime
	err := r.db.QueryRow(
		`SELECT id, display_name, email, password_hash, credentials_json, created_at, updated_at FROM auth_users WHERE id = $1`,
		string(id)).Scan(&uid, &displayName, &email, &passwordHash, &credJSON, &createdAt, &updatedAt)
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
		Email: email.String, PasswordHash: passwordHash.String,
		Credentials: creds, CreatedAt: createdAt.Time, UpdatedAt: updatedAt.Time,
	}, nil
}

func (r *PostgresUserRepository) FindByCredentialID(credentialID []byte) (*user.User, *webauthn.Credential, error) {
	rows, err := r.db.Query(`SELECT id, display_name, email, password_hash, credentials_json, created_at, updated_at FROM auth_users`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var uid, displayName string
		var email, passwordHash sql.NullString
		var credJSON []byte
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&uid, &displayName, &email, &passwordHash, &credJSON, &createdAt, &updatedAt); err != nil {
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
					Email: email.String, PasswordHash: passwordHash.String,
					Credentials: creds, CreatedAt: createdAt.Time, UpdatedAt: updatedAt.Time,
				}
				return u, &creds[i], nil
			}
		}
	}
	return nil, nil, errors.New("credential not found")
}

func (r *PostgresUserRepository) FindByEmail(email string) (*user.User, error) {
	var uid, displayName string
	var emailCol, passwordHash sql.NullString
	var credJSON []byte
	var createdAt, updatedAt sql.NullTime
	err := r.db.QueryRow(
		`SELECT id, display_name, email, password_hash, credentials_json, created_at, updated_at FROM auth_users WHERE LOWER(email) = LOWER($1)`,
		email).Scan(&uid, &displayName, &emailCol, &passwordHash, &credJSON, &createdAt, &updatedAt)
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
		Email: emailCol.String, PasswordHash: passwordHash.String,
		Credentials: creds, CreatedAt: createdAt.Time, UpdatedAt: updatedAt.Time,
	}, nil
}

func (r *PostgresUserRepository) Update(u *user.User) error {
	credJSON, err := json.Marshal(u.Credentials)
	if err != nil {
		return err
	}
	emailVal := nullableString(u.Email)
	hashVal := nullableString(u.PasswordHash)
	result, err := r.db.Exec(
		`UPDATE auth_users SET display_name=$2, email=$3, password_hash=$4, credentials_json=$5, updated_at=$6 WHERE id=$1`,
		string(u.ID), u.DisplayName, emailVal, hashVal, credJSON, u.UpdatedAt)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func nullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
