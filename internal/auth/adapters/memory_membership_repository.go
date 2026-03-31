package adapters

import (
	"errors"
	"sync"

	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"
)

type MemoryMembershipRepository struct {
	mu          sync.RWMutex
	memberships map[common.MembershipID]*membership.Membership
}

func NewMemoryMembershipRepository() *MemoryMembershipRepository {
	return &MemoryMembershipRepository{memberships: make(map[common.MembershipID]*membership.Membership)}
}

func (r *MemoryMembershipRepository) Save(m *membership.Membership) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.memberships[m.ID] = m
	return nil
}

func (r *MemoryMembershipRepository) FindByUserID(userID common.UserID) ([]*membership.Membership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*membership.Membership
	for _, m := range r.memberships {
		if m.UserID == userID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *MemoryMembershipRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*membership.Membership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*membership.Membership
	for _, m := range r.memberships {
		if m.RestaurantID == restaurantID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *MemoryMembershipRepository) FindByUserAndRestaurant(userID common.UserID, restaurantID common.RestaurantID) (*membership.Membership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, m := range r.memberships {
		if m.UserID == userID && m.RestaurantID == restaurantID {
			return m, nil
		}
	}
	return nil, errors.New("membership not found")
}

func (r *MemoryMembershipRepository) Delete(id common.MembershipID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.memberships[id]; !exists {
		return errors.New("membership not found")
	}
	delete(r.memberships, id)
	return nil
}
