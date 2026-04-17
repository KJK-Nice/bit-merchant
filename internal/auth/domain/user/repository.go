package user

import (
	"bitmerchant/internal/common"

	"github.com/go-webauthn/webauthn/webauthn"
)

// Repository defines operations for User persistence.
type Repository interface {
	Save(user *User) error
	FindByID(id common.UserID) (*User, error)
	FindByCredentialID(credentialID []byte) (*User, *webauthn.Credential, error)
	FindByEmail(email string) (*User, error)
	Update(user *User) error
}
