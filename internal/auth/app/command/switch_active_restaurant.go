package command

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
)

// ErrMembershipNotFound is returned when the user cannot access the target restaurant.
var ErrMembershipNotFound = errors.New("membership not found")

// SwitchActiveRestaurant updates the session to point at a restaurant the user belongs to.
type SwitchActiveRestaurant struct {
	SessionID    string
	UserID       common.UserID
	RestaurantID common.RestaurantID
	SessionTTL   time.Duration
}

type SwitchActiveRestaurantHandler decorator.CommandResultHandler[SwitchActiveRestaurant, *session.Session]

type switchActiveRestaurantHandler struct {
	memRepo  membership.Repository
	sessRepo session.Repository
}

func NewSwitchActiveRestaurantHandler(memRepo membership.Repository, sessRepo session.Repository, log *slog.Logger, metrics decorator.MetricsClient) SwitchActiveRestaurantHandler {
	if memRepo == nil || sessRepo == nil {
		panic("nil repository")
	}
	h := switchActiveRestaurantHandler{memRepo: memRepo, sessRepo: sessRepo}
	return decorator.ApplyCommandResultDecorators[SwitchActiveRestaurant, *session.Session](h, log, metrics)
}

func (h switchActiveRestaurantHandler) Handle(ctx context.Context, cmd SwitchActiveRestaurant) (*session.Session, error) {
	_ = ctx
	if _, err := h.memRepo.FindByUserAndRestaurant(cmd.UserID, cmd.RestaurantID); err != nil {
		return nil, ErrMembershipNotFound
	}

	currentSession, err := h.sessRepo.Get(cmd.SessionID)
	if err != nil || currentSession == nil {
		currentSession = &session.Session{
			ID:        cmd.SessionID,
			CreatedAt: time.Now(),
		}
	}
	currentSession.UserID = &cmd.UserID
	currentSession.RestaurantID = &cmd.RestaurantID
	ttl := cmd.SessionTTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	currentSession.ExpiresAt = time.Now().Add(ttl)
	if err := h.sessRepo.Save(currentSession); err != nil {
		return nil, err
	}
	return currentSession, nil
}
