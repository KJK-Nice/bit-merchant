package sse

import (
	"bytes"
	"context"
	"fmt"

	"bitmerchant/internal/common"
	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/interfaces/templates"
	orderQuery "bitmerchant/internal/ordering/app/query"
	"bitmerchant/internal/ordering/domain/order"
)

// broadcastCustomerStatus renders OrderStatus for a single order and pushes the
// result on the order's SSE topic. It also rebroadcasts every active order in
// the same restaurant created after `o`, because their queue position drops
// by one when `o` advances.
func broadcastCustomerStatus(ctx context.Context, logger *logging.Logger, sse *commonhttp.SSEHandler, repo order.Repository, o *order.Order) {
	if o == nil {
		return
	}
	pushView(ctx, logger, sse, repo, o)

	if !shouldRebroadcastQueue(o) {
		return
	}
	cohort, err := repo.FindActiveByRestaurantID(o.RestaurantID)
	if err != nil {
		logger.Error("queue rebroadcast: lookup failed", "error", err)
		return
	}
	for _, peer := range cohort {
		if peer == nil || peer.ID == o.ID {
			continue
		}
		if peer.CreatedAt.After(o.CreatedAt) && peer.FulfillmentStatus == common.FulfillmentStatusPaid {
			pushView(ctx, logger, sse, repo, peer)
		}
	}
}

func pushView(ctx context.Context, logger *logging.Logger, sse *commonhttp.SSEHandler, repo order.Repository, o *order.Order) {
	view, err := orderQuery.BuildOrderStatusView(repo, o, orderQuery.DefaultPrepTarget)
	if err != nil {
		logger.Error("status view: build failed", "error", err, "orderNumber", o.OrderNumber)
		return
	}
	var buf bytes.Buffer
	if err := templates.OrderStatus(view).Render(ctx, &buf); err != nil {
		logger.Error("status view: render failed", "error", err, "orderNumber", o.OrderNumber)
		return
	}
	msg := commonhttp.FormatDatastarEvent(buf.String())
	sse.Broadcast(fmt.Sprintf(commonhttp.TopicOrder, o.OrderNumber), msg)
}

// shouldRebroadcastQueue returns true when an order's transition shifts the
// queue (i.e. when it leaves the "paid, waiting to start" position). A paid
// order moving to preparing/ready/completed bumps everyone behind it forward.
func shouldRebroadcastQueue(o *order.Order) bool {
	switch o.FulfillmentStatus {
	case common.FulfillmentStatusPreparing,
		common.FulfillmentStatusReady,
		common.FulfillmentStatusCompleted:
		return true
	}
	return false
}
