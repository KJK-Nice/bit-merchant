package command

import (
	"context"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/places/domain/visit"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

type RecordMenuVisitUseCase struct {
	restaurants restaurant.Repository
	visits      visit.Repository
}

func NewRecordMenuVisitUseCase(restaurants restaurant.Repository, visits visit.Repository) *RecordMenuVisitUseCase {
	return &RecordMenuVisitUseCase{restaurants: restaurants, visits: visits}
}

func (uc *RecordMenuVisitUseCase) Execute(ctx context.Context, sessionID string, restaurantID common.RestaurantID) error {
	if sessionID == "" || restaurantID == "" {
		return nil
	}
	if _, err := uc.restaurants.FindByID(restaurantID); err != nil {
		return err
	}
	now := time.Now()
	return uc.visits.Upsert(&visit.SessionRestaurantVisit{
		SessionID: sessionID, RestaurantID: restaurantID,
		FirstVisitedAt: now, LastVisitedAt: now,
	})
}
