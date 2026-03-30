package memory

import (
	"errors"
	"sync"

	"bitmerchant/internal/domain"
)

// MemorySessionRepository implements SessionRepository with in-memory storage.
type MemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

// NewMemorySessionRepository creates a new in-memory session repository.
func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{
		sessions: make(map[string]*domain.Session),
	}
}

// Save saves a session.
func (r *MemorySessionRepository) Save(session *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.ID] = session
	return nil
}

// Get retrieves a session by ID.
func (r *MemorySessionRepository) Get(id string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, exists := r.sessions[id]
	if !exists {
		return nil, errors.New("session not found")
	}
	return session, nil
}

// Delete removes a session by ID.
func (r *MemorySessionRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, id)
	return nil
}

// DeleteByUserID removes all sessions associated with a user.
func (r *MemorySessionRepository) DeleteByUserID(userID domain.UserID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, session := range r.sessions {
		if session.UserID != nil && *session.UserID == userID {
			delete(r.sessions, id)
		}
	}
	return nil
}
