package memory

import (
	"errors"
	"sync"

	"bitmerchant/internal/domain"
)

// MemoryInvitationRepository implements InvitationRepository with in-memory storage.
type MemoryInvitationRepository struct {
	mu          sync.RWMutex
	invitations map[domain.InvitationID]*domain.Invitation
}

// NewMemoryInvitationRepository creates a new in-memory invitation repository.
func NewMemoryInvitationRepository() *MemoryInvitationRepository {
	return &MemoryInvitationRepository{
		invitations: make(map[domain.InvitationID]*domain.Invitation),
	}
}

// Save saves an invitation.
func (r *MemoryInvitationRepository) Save(invitation *domain.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invitations[invitation.ID] = invitation
	return nil
}

// FindByToken finds an invitation by token.
func (r *MemoryInvitationRepository) FindByToken(token string) (*domain.Invitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, invitation := range r.invitations {
		if invitation.Token == token {
			return invitation, nil
		}
	}
	return nil, errors.New("invitation not found")
}

// FindByRestaurantID finds invitations by restaurant ID.
func (r *MemoryInvitationRepository) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.Invitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Invitation
	for _, invitation := range r.invitations {
		if invitation.RestaurantID == restaurantID {
			result = append(result, invitation)
		}
	}
	return result, nil
}

// Update updates an invitation.
func (r *MemoryInvitationRepository) Update(invitation *domain.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.invitations[invitation.ID]; !exists {
		return errors.New("invitation not found")
	}
	r.invitations[invitation.ID] = invitation
	return nil
}
