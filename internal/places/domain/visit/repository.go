package visit

// Repository defines operations for SessionRestaurantVisit persistence.
type Repository interface {
	Upsert(visit *SessionRestaurantVisit) error
	FindBySessionID(sessionID string) ([]*SessionRestaurantVisit, error)
}
