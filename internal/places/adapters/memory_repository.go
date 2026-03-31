package adapters

import (
	"sort"
	"sync"
	"time"

	"bitmerchant/internal/places/domain/visit"
)

type MemoryVisitRepository struct {
	mu     sync.RWMutex
	visits map[string]map[string]*visit.SessionRestaurantVisit
}

func NewMemoryVisitRepository() *MemoryVisitRepository {
	return &MemoryVisitRepository{
		visits: make(map[string]map[string]*visit.SessionRestaurantVisit),
	}
}

func (r *MemoryVisitRepository) Upsert(v *visit.SessionRestaurantVisit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if v == nil || v.SessionID == "" || v.RestaurantID == "" {
		return nil
	}
	key := string(v.RestaurantID)
	now := time.Now()
	if r.visits[v.SessionID] == nil {
		r.visits[v.SessionID] = make(map[string]*visit.SessionRestaurantVisit)
	}
	existing := r.visits[v.SessionID][key]
	if existing == nil {
		cp := *v
		if cp.FirstVisitedAt.IsZero() {
			cp.FirstVisitedAt = now
		}
		if cp.LastVisitedAt.IsZero() {
			cp.LastVisitedAt = now
		}
		r.visits[v.SessionID][key] = &cp
		return nil
	}
	existing.LastVisitedAt = now
	return nil
}

func (r *MemoryVisitRepository) FindBySessionID(sessionID string) ([]*visit.SessionRestaurantVisit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m := r.visits[sessionID]
	if len(m) == 0 {
		return nil, nil
	}
	out := make([]*visit.SessionRestaurantVisit, 0, len(m))
	for _, v := range m {
		cp := *v
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].LastVisitedAt.After(out[j].LastVisitedAt)
	})
	return out, nil
}
