package adapters

import (
	"errors"
	"sync"

	"bitmerchant/internal/auth/domain/passwordreset"
)

type MemoryPasswordResetTokenRepository struct {
	mu     sync.RWMutex
	byHash map[string]*passwordreset.Token
}

func NewMemoryPasswordResetTokenRepository() *MemoryPasswordResetTokenRepository {
	return &MemoryPasswordResetTokenRepository{byHash: make(map[string]*passwordreset.Token)}
}

func (r *MemoryPasswordResetTokenRepository) Save(token *passwordreset.Token) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *token
	r.byHash[token.TokenHash] = &cp
	return nil
}

func (r *MemoryPasswordResetTokenRepository) FindByHash(tokenHash string) (*passwordreset.Token, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.byHash[tokenHash]
	if !ok {
		return nil, errors.New("reset token not found")
	}
	cp := *t
	return &cp, nil
}

func (r *MemoryPasswordResetTokenRepository) Update(token *passwordreset.Token) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byHash[token.TokenHash]; !ok {
		return errors.New("reset token not found")
	}
	cp := *token
	r.byHash[token.TokenHash] = &cp
	return nil
}
