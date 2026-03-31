package adapters

import (
	"bytes"
	"errors"
	"sync"

	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"

	"github.com/go-webauthn/webauthn/webauthn"
)

type MemoryUserRepository struct {
	mu    sync.RWMutex
	users map[common.UserID]*user.User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{users: make(map[common.UserID]*user.User)}
}

func (r *MemoryUserRepository) Save(u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	return nil
}

func (r *MemoryUserRepository) FindByID(id common.UserID) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, exists := r.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (r *MemoryUserRepository) FindByCredentialID(credentialID []byte) (*user.User, *webauthn.Credential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		for i := range u.Credentials {
			if bytes.Equal(u.Credentials[i].ID, credentialID) {
				return u, &u.Credentials[i], nil
			}
		}
	}
	return nil, nil, errors.New("credential not found")
}

func (r *MemoryUserRepository) Update(u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[u.ID]; !exists {
		return errors.New("user not found")
	}
	r.users[u.ID] = u
	return nil
}
