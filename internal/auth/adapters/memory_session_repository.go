package adapters

import (
	"errors"
	"sync"

	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/common"
)

type MemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*session.Session
}

func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{sessions: make(map[string]*session.Session)}
}

func (r *MemorySessionRepository) Save(s *session.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.ID] = s
	return nil
}

func (r *MemorySessionRepository) Get(id string) (*session.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, exists := r.sessions[id]
	if !exists {
		return nil, errors.New("session not found")
	}
	return s, nil
}

func (r *MemorySessionRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, id)
	return nil
}

func (r *MemorySessionRepository) DeleteByUserID(userID common.UserID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, s := range r.sessions {
		if s.UserID != nil && *s.UserID == userID {
			delete(r.sessions, id)
		}
	}
	return nil
}
