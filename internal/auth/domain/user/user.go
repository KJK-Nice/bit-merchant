package user

import (
	"errors"
	"time"

	"bitmerchant/internal/common"

	"github.com/go-webauthn/webauthn/webauthn"
)

// User represents an authenticated actor in the system.
type User struct {
	ID          common.UserID
	DisplayName string
	Credentials []webauthn.Credential
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewUser(id common.UserID, displayName string) (*User, error) {
	if displayName == "" {
		return nil, errors.New("display name is required")
	}
	now := time.Now()
	return &User{
		ID: id, DisplayName: displayName,
		Credentials: []webauthn.Credential{},
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (u *User) AddCredential(credential webauthn.Credential) {
	u.Credentials = append(u.Credentials, credential)
	u.UpdatedAt = time.Now()
}

func (u *User) UpdateCredential(updated webauthn.Credential) {
	for i := range u.Credentials {
		if string(u.Credentials[i].ID) == string(updated.ID) {
			u.Credentials[i] = updated
			u.UpdatedAt = time.Now()
			return
		}
	}
}

func (u *User) WebAuthnID() []byte                         { return []byte(u.ID) }
func (u *User) WebAuthnName() string                       { return u.DisplayName }
func (u *User) WebAuthnDisplayName() string                { return u.DisplayName }
func (u *User) WebAuthnCredentials() []webauthn.Credential { return u.Credentials }
