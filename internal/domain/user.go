package domain

import (
	"errors"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// UserID represents a unique user identifier.
type UserID string

// User represents an authenticated actor in the system.
type User struct {
	ID          UserID
	DisplayName string
	Credentials []webauthn.Credential
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewUser creates a new user aggregate.
func NewUser(id UserID, displayName string) (*User, error) {
	if displayName == "" {
		return nil, errors.New("display name is required")
	}

	now := time.Now()
	return &User{
		ID:          id,
		DisplayName: displayName,
		Credentials: []webauthn.Credential{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// AddCredential appends a new credential to the user.
func (u *User) AddCredential(credential webauthn.Credential) {
	u.Credentials = append(u.Credentials, credential)
	u.UpdatedAt = time.Now()
}

// UpdateCredential updates an existing credential in-place.
func (u *User) UpdateCredential(updated webauthn.Credential) {
	for i := range u.Credentials {
		if string(u.Credentials[i].ID) == string(updated.ID) {
			u.Credentials[i] = updated
			u.UpdatedAt = time.Now()
			return
		}
	}
}

// WebAuthnID implements webauthn.User.
func (u *User) WebAuthnID() []byte {
	return []byte(u.ID)
}

// WebAuthnName implements webauthn.User.
func (u *User) WebAuthnName() string {
	return u.DisplayName
}

// WebAuthnDisplayName implements webauthn.User.
func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnCredentials implements webauthn.User.
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}
