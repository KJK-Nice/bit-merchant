package adapters

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(hash, plain string) error
}

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{cost: bcrypt.DefaultCost}
}

func (h *BcryptPasswordHasher) Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *BcryptPasswordHasher) Verify(hash, plain string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return errors.New("invalid credentials")
	}
	return err
}
