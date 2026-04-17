package command

import (
	"context"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/places/domain/visit"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"log/slog"
)

// RecordMenuVisit records that a session viewed a restaurant menu.
type RecordMenuVisit struct {
	SessionID    string
	RestaurantID common.RestaurantID
}

type RecordMenuVisitHandler decorator.CommandHandler[RecordMenuVisit]

type recordMenuVisitHandler struct {
	restaurants restaurant.Repository
	visits      visit.Repository
}

func NewRecordMenuVisitHandler(restaurants restaurant.Repository, visits visit.Repository, log *slog.Logger, metrics decorator.MetricsClient) RecordMenuVisitHandler {
	if restaurants == nil {
		panic("nil restaurant.Repository")
	}
	if visits == nil {
		panic("nil visit.Repository")
	}
	h := recordMenuVisitHandler{restaurants: restaurants, visits: visits}
	return decorator.ApplyCommandDecorators[RecordMenuVisit](h, log, metrics)
}

func (h recordMenuVisitHandler) Handle(ctx context.Context, cmd RecordMenuVisit) error {
	if cmd.SessionID == "" || cmd.RestaurantID == "" {
		return nil
	}
	if _, err := h.restaurants.FindByID(cmd.RestaurantID); err != nil {
		return err
	}
	now := time.Now()
	return h.visits.Upsert(ctx, visit.NewSessionRestaurantVisit(cmd.SessionID, cmd.RestaurantID, now, now))
}
