package places

import (
	"context"
	"time"

	"bitmerchant/internal/domain"
)

// RecordMenuVisitUseCase records that the current session opened a restaurant menu.
type RecordMenuVisitUseCase struct {
	restaurants domain.RestaurantRepository
	visits      domain.SessionRestaurantVisitRepository
}

// NewRecordMenuVisitUseCase constructs the use case.
func NewRecordMenuVisitUseCase(restaurants domain.RestaurantRepository, visits domain.SessionRestaurantVisitRepository) *RecordMenuVisitUseCase {
	return &RecordMenuVisitUseCase{restaurants: restaurants, visits: visits}
}

// Execute validates the restaurant exists and upserts a visit.
func (uc *RecordMenuVisitUseCase) Execute(ctx context.Context, sessionID string, restaurantID domain.RestaurantID) error {
	if sessionID == "" || restaurantID == "" {
		return nil
	}
	if _, err := uc.restaurants.FindByID(restaurantID); err != nil {
		return err
	}
	now := time.Now()
	return uc.visits.Upsert(&domain.SessionRestaurantVisit{
		SessionID:      sessionID,
		RestaurantID:   restaurantID,
		FirstVisitedAt: now,
		LastVisitedAt:  now,
	})
}
