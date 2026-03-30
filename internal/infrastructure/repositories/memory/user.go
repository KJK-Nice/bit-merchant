package memory

import (
	"bytes"
	"errors"
	"sync"

	"bitmerchant/internal/domain"

	"github.com/go-webauthn/webauthn/webauthn"
)

// MemoryUserRepository implements UserRepository with in-memory storage.
type MemoryUserRepository struct {
	mu    sync.RWMutex
	users map[domain.UserID]*domain.User
}

// NewMemoryUserRepository creates a new in-memory user repository.
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users: make(map[domain.UserID]*domain.User),
	}
}

// Save saves a user.
func (r *MemoryUserRepository) Save(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

// FindByID finds a user by ID.
func (r *MemoryUserRepository) FindByID(id domain.UserID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// FindByCredentialID finds a user by credential ID.
func (r *MemoryUserRepository) FindByCredentialID(credentialID []byte) (*domain.User, *webauthn.Credential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, user := range r.users {
		for i := range user.Credentials {
			if bytes.Equal(user.Credentials[i].ID, credentialID) {
				return user, &user.Credentials[i], nil
			}
		}
	}
	return nil, nil, errors.New("credential not found")
}

// Update updates a user.
func (r *MemoryUserRepository) Update(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; !exists {
		return errors.New("user not found")
	}
	r.users[user.ID] = user
	return nil
}
