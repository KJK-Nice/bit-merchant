package query

import (
	"context"
	"sort"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/menu/domain/menu"
	"bitmerchant/internal/ordering/domain/order"
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

const topItemsLimit = 5

type topItemBucket struct {
	menuItemID common.ItemID
	name       string
	quantity   int
	revenue    float64
}

func (h topSellingMenuItemsHandler) Handle(ctx context.Context, q TopSellingMenuItems) ([]TopItem, error) {
	_ = ctx
	orders, err := h.orders.FindByRestaurantID(q.RestaurantID)
	if err != nil {
		return nil, err
	}
	result := topItemsFromBuckets(aggregatePaidItems(orders))
	annotateTopItems(result, h.items)
	return result, nil
}

// aggregatePaidItems folds line items across paid orders into per-item
// buckets keyed by menu-item ID (or name when the ID is missing).
func aggregatePaidItems(orders []*order.Order) map[string]*topItemBucket {
	buckets := make(map[string]*topItemBucket)
	for _, o := range orders {
		if o.PaymentStatus != common.PaymentStatusPaid {
			continue
		}
		for _, item := range o.Items {
			key := bucketKey(item.MenuItemID, item.Name)
			b, ok := buckets[key]
			if !ok {
				b = &topItemBucket{menuItemID: item.MenuItemID, name: item.Name}
				buckets[key] = b
			}
			b.quantity += item.Quantity
			b.revenue += item.Subtotal
		}
	}
	return buckets
}

func bucketKey(id common.ItemID, name string) string {
	if id != "" {
		return string(id)
	}
	return "name:" + name
}

// topItemsFromBuckets sorts buckets, trims to topItemsLimit, and computes
// each row's RevenueShare against the leader.
func topItemsFromBuckets(buckets map[string]*topItemBucket) []TopItem {
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
	if len(result) > topItemsLimit {
		result = result[:topItemsLimit]
	}
	if len(result) > 0 && result[0].Revenue > 0 {
		leader := result[0].Revenue
		for i := range result {
			result[i].RevenueShare = result[i].Revenue / leader
		}
	}
	return result
}

// annotateTopItems fills PhotoURL from the menu item repo. Best-effort —
// the row is left thumb-less when the originating menu item has been
// deleted or the repo is nil (tests).
func annotateTopItems(items []TopItem, repo menu.ItemRepository) {
	if repo == nil {
		return
	}
	for i := range items {
		if items[i].MenuItemID == "" {
			continue
		}
		mi, err := repo.FindByID(items[i].MenuItemID)
		if err != nil || mi == nil {
			continue
		}
		items[i].PhotoURL = mi.PhotoURL
	}
}
