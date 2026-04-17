package query

import (
	"context"
	"sort"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"log/slog"
)

type TopItem struct {
	Name     string
	Quantity int
	Revenue  float64
}

// TopSellingMenuItems returns up to five best-selling items for a restaurant.
type TopSellingMenuItems struct {
	RestaurantID common.RestaurantID
}

type TopSellingMenuItemsHandler decorator.QueryHandler[TopSellingMenuItems, []TopItem]

type topSellingMenuItemsHandler struct {
	orders OrderReadModel
}

func NewTopSellingMenuItemsHandler(orders OrderReadModel, log *slog.Logger, metrics decorator.MetricsClient) TopSellingMenuItemsHandler {
	if orders == nil {
		panic("nil OrderReadModel")
	}
	h := topSellingMenuItemsHandler{orders: orders}
	return decorator.ApplyQueryDecorators[TopSellingMenuItems, []TopItem](h, log, metrics)
}

func (h topSellingMenuItemsHandler) Handle(ctx context.Context, q TopSellingMenuItems) ([]TopItem, error) {
	_ = ctx
	orders, err := h.orders.FindByRestaurantID(q.RestaurantID)
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
