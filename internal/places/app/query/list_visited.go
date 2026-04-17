package query

import (
	"context"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/ordering/domain/order"
	"bitmerchant/internal/places/domain/visit"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"log/slog"
)

// VisitedPlace is a read model row for the customer "my places" list.
type VisitedPlace struct {
	RestaurantID   common.RestaurantID
	Name           string
	LastVisitedAt  time.Time
	HasOrderedHere bool
	IsOpen         bool
}

// SessionVisitedPlaces lists restaurants visited in a session.
type SessionVisitedPlaces struct {
	SessionID string
}

type SessionVisitedPlacesHandler decorator.QueryHandler[SessionVisitedPlaces, []VisitedPlace]

type sessionVisitedPlacesHandler struct {
	visits      visit.Repository
	restaurants restaurant.Repository
	orders      order.Repository
}

func NewSessionVisitedPlacesHandler(visits visit.Repository, restaurants restaurant.Repository, orders order.Repository, log *slog.Logger, metrics decorator.MetricsClient) SessionVisitedPlacesHandler {
	if visits == nil {
		panic("nil visit.Repository")
	}
	if restaurants == nil {
		panic("nil restaurant.Repository")
	}
	h := sessionVisitedPlacesHandler{visits: visits, restaurants: restaurants, orders: orders}
	return decorator.ApplyQueryDecorators[SessionVisitedPlaces, []VisitedPlace](h, log, metrics)
}

func (h sessionVisitedPlacesHandler) Handle(ctx context.Context, q SessionVisitedPlaces) ([]VisitedPlace, error) {
	if q.SessionID == "" {
		return nil, nil
	}
	visitRows, err := h.visits.FindBySessionID(ctx, q.SessionID)
	if err != nil {
		return nil, err
	}
	ordered := make(map[common.RestaurantID]struct{})
	if h.orders != nil {
		orders, oerr := h.orders.FindBySessionID(q.SessionID)
		if oerr == nil {
			for _, o := range orders {
				ordered[o.RestaurantID] = struct{}{}
			}
		}
	}
	out := make([]VisitedPlace, 0, len(visitRows))
	for _, v := range visitRows {
		rest, err := h.restaurants.FindByID(v.RestaurantID())
		if err != nil || rest == nil {
			continue
		}
		_, hasOrdered := ordered[v.RestaurantID()]
		out = append(out, VisitedPlace{
			RestaurantID: v.RestaurantID(), Name: rest.Name,
			LastVisitedAt: v.LastVisitedAt(), HasOrderedHere: hasOrdered,
			IsOpen: rest.IsOpen,
		})
	}
	return out, nil
}
