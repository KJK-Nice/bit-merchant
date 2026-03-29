package memory

import (
	"sort"
	"sync"
	"time"

	"bitmerchant/internal/domain"
)

// MemorySessionRestaurantVisitRepository stores session→restaurant visits in memory.
type MemorySessionRestaurantVisitRepository struct {
	mu      sync.RWMutex
	visits  map[string]map[string]*domain.SessionRestaurantVisit // sessionID -> restaurantID string -> visit
}

// NewMemorySessionRestaurantVisitRepository constructs the repository.
func NewMemorySessionRestaurantVisitRepository() *MemorySessionRestaurantVisitRepository {
	return &MemorySessionRestaurantVisitRepository{
		visits: make(map[string]map[string]*domain.SessionRestaurantVisit),
	}
}

// Upsert creates or updates last_visited_at for a session/restaurant pair.
func (r *MemorySessionRestaurantVisitRepository) Upsert(visit *domain.SessionRestaurantVisit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if visit == nil || visit.SessionID == "" || visit.RestaurantID == "" {
		return nil
	}
	key := string(visit.RestaurantID)
	now := time.Now()
	if r.visits[visit.SessionID] == nil {
		r.visits[visit.SessionID] = make(map[string]*domain.SessionRestaurantVisit)
	}
	existing := r.visits[visit.SessionID][key]
	if existing == nil {
		v := *visit
		if v.FirstVisitedAt.IsZero() {
			v.FirstVisitedAt = now
		}
		if v.LastVisitedAt.IsZero() {
			v.LastVisitedAt = now
		}
		cp := v
		r.visits[visit.SessionID][key] = &cp
		return nil
	}
	existing.LastVisitedAt = now
	return nil
}

// FindBySessionID returns visits for a session, newest last_visited_at first.
func (r *MemorySessionRestaurantVisitRepository) FindBySessionID(sessionID string) ([]*domain.SessionRestaurantVisit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m := r.visits[sessionID]
	if len(m) == 0 {
		return nil, nil
	}
	out := make([]*domain.SessionRestaurantVisit, 0, len(m))
	for _, v := range m {
		cp := *v
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].LastVisitedAt.After(out[j].LastVisitedAt)
	})
	return out, nil
}
