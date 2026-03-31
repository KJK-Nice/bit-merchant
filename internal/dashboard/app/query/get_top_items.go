package query

import (
	"context"
	"sort"

	"bitmerchant/internal/common"
	"bitmerchant/internal/ordering/domain/order"
)

type TopItem struct {
	Name     string
	Quantity int
	Revenue  float64
}

type GetTopSellingItemsUseCase struct {
	orderRepo order.Repository
}

func NewGetTopSellingItemsUseCase(orderRepo order.Repository) *GetTopSellingItemsUseCase {
	return &GetTopSellingItemsUseCase{orderRepo: orderRepo}
}

func (uc *GetTopSellingItemsUseCase) Execute(ctx context.Context, restaurantID common.RestaurantID) ([]TopItem, error) {
	orders, err := uc.orderRepo.FindByRestaurantID(restaurantID)
	if err != nil {
		return nil, err
	}
	itemStats := make(map[string]*TopItem)
	for _, o := range orders {
		if o.PaymentStatus != common.PaymentStatusPaid {
			continue
		}
		for _, item := range o.Items {
			if _, exists := itemStats[item.Name]; !exists {
				itemStats[item.Name] = &TopItem{Name: item.Name}
			}
			stats := itemStats[item.Name]
			stats.Quantity += item.Quantity
			stats.Revenue += item.Subtotal
		}
	}
	var result []TopItem
	for _, stats := range itemStats {
		result = append(result, *stats)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Quantity > result[j].Quantity
	})
	if len(result) > 5 {
		result = result[:5]
	}
	return result, nil
}
