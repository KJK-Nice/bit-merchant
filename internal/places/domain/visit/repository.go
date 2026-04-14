package visit

import "context"

// Repository defines operations for SessionRestaurantVisit persistence.
type Repository interface {
	Upsert(ctx context.Context, visit *SessionRestaurantVisit) error
	FindBySessionID(ctx context.Context, sessionID string) ([]*SessionRestaurantVisit, error)
}
