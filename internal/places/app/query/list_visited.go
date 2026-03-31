package query

import (
	"context"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
	"bitmerchant/internal/places/domain/visit"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

type VisitedPlace struct {
	RestaurantID   common.RestaurantID
	Name           string
	LastVisitedAt  time.Time
	HasOrderedHere bool
	IsOpen         bool
}

type ListVisitedRestaurantsUseCase struct {
	visits      visit.Repository
	restaurants restaurant.Repository
	orders      order.Repository
}

func NewListVisitedRestaurantsUseCase(visits visit.Repository, restaurants restaurant.Repository, orders order.Repository) *ListVisitedRestaurantsUseCase {
	return &ListVisitedRestaurantsUseCase{visits: visits, restaurants: restaurants, orders: orders}
}

func (uc *ListVisitedRestaurantsUseCase) Execute(ctx context.Context, sessionID string) ([]VisitedPlace, error) {
	if sessionID == "" {
		return nil, nil
	}
	visitRows, err := uc.visits.FindBySessionID(sessionID)
	if err != nil {
		return nil, err
	}
	ordered := make(map[common.RestaurantID]struct{})
	if uc.orders != nil {
		orders, oerr := uc.orders.FindBySessionID(sessionID)
		if oerr == nil {
			for _, o := range orders {
				ordered[o.RestaurantID] = struct{}{}
			}
		}
	}
	out := make([]VisitedPlace, 0, len(visitRows))
	for _, v := range visitRows {
		rest, err := uc.restaurants.FindByID(v.RestaurantID)
		if err != nil || rest == nil {
			continue
		}
		_, hasOrdered := ordered[v.RestaurantID]
		out = append(out, VisitedPlace{
			RestaurantID: v.RestaurantID, Name: rest.Name,
			LastVisitedAt: v.LastVisitedAt, HasOrderedHere: hasOrdered,
			IsOpen: rest.IsOpen,
		})
	}
	return out, nil
}
