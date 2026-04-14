package adapters

import (
	"context"
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

func (r *MemoryVisitRepository) Upsert(_ context.Context, v *visit.SessionRestaurantVisit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if v == nil || v.SessionID() == "" || v.RestaurantID() == "" {
		return nil
	}
	key := string(v.RestaurantID())
	now := time.Now()
	sessionID := v.SessionID()
	if r.visits[sessionID] == nil {
		r.visits[sessionID] = make(map[string]*visit.SessionRestaurantVisit)
	}
	existing := r.visits[sessionID][key]
	if existing == nil {
		cp := v.Clone()
		if cp.FirstVisitedAt().IsZero() {
			cp.Touch(now)
		}
		if cp.LastVisitedAt().IsZero() {
			cp.Touch(now)
		}
		r.visits[sessionID][key] = cp
		return nil
	}
	existing.Touch(now)
	return nil
}

func (r *MemoryVisitRepository) FindBySessionID(_ context.Context, sessionID string) ([]*visit.SessionRestaurantVisit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m := r.visits[sessionID]
	if len(m) == 0 {
		return nil, nil
	}
	out := make([]*visit.SessionRestaurantVisit, 0, len(m))
	for _, v := range m {
		out = append(out, v.Clone())
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].LastVisitedAt().After(out[j].LastVisitedAt())
	})
	return out, nil
}
