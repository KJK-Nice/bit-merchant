package dashboard

import (
	"context"
	"sort"

	"bitmerchant/internal/domain"
)

type TopItem struct {
	Name     string
	Quantity int
	Revenue  float64
}

type GetTopSellingItemsUseCase struct {
	orderRepo domain.OrderRepository
}

func NewGetTopSellingItemsUseCase(orderRepo domain.OrderRepository) *GetTopSellingItemsUseCase {
	return &GetTopSellingItemsUseCase{
		orderRepo: orderRepo,
	}
}

func (uc *GetTopSellingItemsUseCase) Execute(ctx context.Context, restaurantID domain.RestaurantID) ([]TopItem, error) {
	orders, err := uc.orderRepo.FindByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}

	// Map: ItemName -> Stats
	itemStats := make(map[string]*TopItem)

	for _, o := range orders {
		if o.PaymentStatus != domain.PaymentStatusPaid {
			continue
		}

		for _, item := range o.Items {
			if _, exists := itemStats[item.Name]; !exists {
				itemStats[item.Name] = &TopItem{
					Name: item.Name,
				}
			}
			stats := itemStats[item.Name]
			stats.Quantity += item.Quantity
			stats.Revenue += item.Subtotal
		}
	}

	// Convert to slice
	var result []TopItem
	for _, stats := range itemStats {
		result = append(result, *stats)
	}

	// Sort by Quantity desc
	sort.Slice(result, func(i, j int) bool {
		return result[i].Quantity > result[j].Quantity
	})

	// Limit to top 5? Requirement says "Top Items".
	if len(result) > 5 {
		result = result[:5]
	}

	return result, nil
}

