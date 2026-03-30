package memory

import (
	"errors"
	"sync"

	"bitmerchant/internal/domain"
)

// MemoryMembershipRepository implements MembershipRepository with in-memory storage.
type MemoryMembershipRepository struct {
	mu          sync.RWMutex
	memberships map[domain.MembershipID]*domain.Membership
}

// NewMemoryMembershipRepository creates a new in-memory membership repository.
func NewMemoryMembershipRepository() *MemoryMembershipRepository {
	return &MemoryMembershipRepository{
		memberships: make(map[domain.MembershipID]*domain.Membership),
	}
}

// Save saves a membership.
func (r *MemoryMembershipRepository) Save(membership *domain.Membership) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.memberships[membership.ID] = membership
	return nil
}

// FindByUserID finds memberships by user ID.
func (r *MemoryMembershipRepository) FindByUserID(userID domain.UserID) ([]*domain.Membership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Membership
	for _, membership := range r.memberships {
		if membership.UserID == userID {
			result = append(result, membership)
		}
	}
	return result, nil
}

// FindByRestaurantID finds memberships by restaurant ID.
func (r *MemoryMembershipRepository) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Membership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Membership
	for _, membership := range r.memberships {
		if membership.RestaurantID == restaurantID {
			result = append(result, membership)
		}
	}
	return result, nil
}

// FindByUserAndRestaurant finds a membership by user and restaurant.
func (r *MemoryMembershipRepository) FindByUserAndRestaurant(userID domain.UserID, restaurantID domain.RestaurantID) (*domain.Membership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, membership := range r.memberships {
		if membership.UserID == userID && membership.RestaurantID == restaurantID {
			return membership, nil
		}
	}
	return nil, errors.New("membership not found")
}

// Delete removes a membership.
func (r *MemoryMembershipRepository) Delete(id domain.MembershipID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.memberships[id]; !exists {
		return errors.New("membership not found")
	}
	delete(r.memberships, id)
	return nil
}
