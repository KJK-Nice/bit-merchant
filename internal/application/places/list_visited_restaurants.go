package places

import (
	"context"
	"time"

	"bitmerchant/internal/domain"
)

// VisitedPlace is a restaurant the customer has opened on this session.
type VisitedPlace struct {
	RestaurantID   domain.RestaurantID
	Name           string
	LastVisitedAt  time.Time
	HasOrderedHere bool
}

// ListVisitedRestaurantsUseCase builds the "My places" list for a session.
type ListVisitedRestaurantsUseCase struct {
	visits      domain.SessionRestaurantVisitRepository
	restaurants domain.RestaurantRepository
	orders      domain.OrderRepository
}

// NewListVisitedRestaurantsUseCase constructs the use case.
func NewListVisitedRestaurantsUseCase(
	visits domain.SessionRestaurantVisitRepository,
	restaurants domain.RestaurantRepository,
	orders domain.OrderRepository,
) *ListVisitedRestaurantsUseCase {
	return &ListVisitedRestaurantsUseCase{
		visits: visits, restaurants: restaurants, orders: orders,
	}
}

// Execute returns visited places with display names, newest visit first.
func (uc *ListVisitedRestaurantsUseCase) Execute(ctx context.Context, sessionID string) ([]VisitedPlace, error) {
	if sessionID == "" {
		return nil, nil
	}
	visitRows, err := uc.visits.FindBySessionID(sessionID)
	if err != nil {
		return nil, err
	}
	ordered := make(map[domain.RestaurantID]struct{})
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
			RestaurantID:   v.RestaurantID,
			Name:           rest.Name,
			LastVisitedAt:  v.LastVisitedAt,
			HasOrderedHere: hasOrdered,
		})
	}
	return out, nil
}
