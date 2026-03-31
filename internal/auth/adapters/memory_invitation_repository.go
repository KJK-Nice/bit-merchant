package adapters

import (
	"errors"
	"sync"

	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/common"
)

type MemoryInvitationRepository struct {
	mu          sync.RWMutex
	invitations map[common.InvitationID]*invitation.Invitation
}

func NewMemoryInvitationRepository() *MemoryInvitationRepository {
	return &MemoryInvitationRepository{invitations: make(map[common.InvitationID]*invitation.Invitation)}
}

func (r *MemoryInvitationRepository) Save(inv *invitation.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invitations[inv.ID] = inv
	return nil
}

func (r *MemoryInvitationRepository) FindByToken(token string) (*invitation.Invitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, inv := range r.invitations {
		if inv.Token == token {
			return inv, nil
		}
	}
	return nil, errors.New("invitation not found")
}

func (r *MemoryInvitationRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*invitation.Invitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*invitation.Invitation
	for _, inv := range r.invitations {
		if inv.RestaurantID == restaurantID {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (r *MemoryInvitationRepository) Update(inv *invitation.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.invitations[inv.ID]; !exists {
		return errors.New("invitation not found")
	}
	r.invitations[inv.ID] = inv
	return nil
}
