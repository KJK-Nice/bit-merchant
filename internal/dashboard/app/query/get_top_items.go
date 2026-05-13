package query

import (
	"context"
	"sort"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/menu/domain/menu"
	"log/slog"
)

// TopItem now carries a menu item handle so the dashboard can render a
// thumbnail and link the row to the item editor. MenuItemID and PhotoURL
// are best-effort — they stay zero-valued when the originating menu item
// has been deleted since the order was placed.
type TopItem struct {
	MenuItemID common.ItemID
	Name       string
	Quantity   int
	Revenue    float64
	PhotoURL   string
	// RevenueShare is this item's revenue divided by the leader's revenue,
	// in the range [0..1]. The renderer multiplies by the bar width.
	RevenueShare float64
}

// TopSellingMenuItems returns up to five best-selling items for a restaurant.
type TopSellingMenuItems struct {
	RestaurantID common.RestaurantID
}

type TopSellingMenuItemsHandler decorator.QueryHandler[TopSellingMenuItems, []TopItem]

type topSellingMenuItemsHandler struct {
	orders OrderReadModel
	items  menu.ItemRepository
}

func NewTopSellingMenuItemsHandler(orders OrderReadModel, items menu.ItemRepository, log *slog.Logger, metrics decorator.MetricsClient) TopSellingMenuItemsHandler {
	if orders == nil {
		panic("nil OrderReadModel")
	}
	h := topSellingMenuItemsHandler{orders: orders, items: items}
	return decorator.ApplyQueryDecorators[TopSellingMenuItems, []TopItem](h, log, metrics)
}

func (h topSellingMenuItemsHandler) Handle(ctx context.Context, q TopSellingMenuItems) ([]TopItem, error) {
	_ = ctx
	orders, err := h.orders.FindByRestaurantID(q.RestaurantID)
	if err != nil {
		return nil, err
	}
	type bucket struct {
		menuItemID common.ItemID
		name       string
		quantity   int
		revenue    float64
	}
	buckets := make(map[string]*bucket)
	for _, o := range orders {
		if o.PaymentStatus != common.PaymentStatusPaid {
			continue
		}
		for _, item := range o.Items {
			key := string(item.MenuItemID)
			if key == "" {
				key = "name:" + item.Name
			}
			b, ok := buckets[key]
			if !ok {
				b = &bucket{menuItemID: item.MenuItemID, name: item.Name}
				buckets[key] = b
			}
			b.quantity += item.Quantity
			b.revenue += item.Subtotal
		}
	}
	result := make([]TopItem, 0, len(buckets))
	for _, b := range buckets {
		result = append(result, TopItem{
			MenuItemID: b.menuItemID,
			Name:       b.name,
			Quantity:   b.quantity,
			Revenue:    b.revenue,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Quantity != result[j].Quantity {
			return result[i].Quantity > result[j].Quantity
		}
		return result[i].Revenue > result[j].Revenue
	})
	if len(result) > 5 {
		result = result[:5]
	}
	// Annotate revenue share against the leader (first row).
	var leader float64
	if len(result) > 0 {
		leader = result[0].Revenue
	}
	for i := range result {
		if leader > 0 {
			result[i].RevenueShare = result[i].Revenue / leader
		}
		if h.items != nil && result[i].MenuItemID != "" {
			if mi, err := h.items.FindByID(result[i].MenuItemID); err == nil && mi != nil {
				result[i].PhotoURL = mi.PhotoURL
			}
		}
	}
	return result, nil
}
