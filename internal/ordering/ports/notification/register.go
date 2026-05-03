package notification

import (
	"encoding/json"
	"fmt"

	"bitmerchant/internal/common"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/notification"
	orderevent "bitmerchant/internal/ordering/app/event"

	"github.com/ThreeDotsLabs/watermill/message"
)

// RegisterOrderNotificationHandlers wires push notification handlers into the Watermill router.
// Handler names use the "notif_order_*" prefix — distinct from the "sse_order_*" SSE handlers,
// allowing Watermill to fan-out each event to both handler sets independently.
func RegisterOrderNotificationHandlers(
	router *message.Router,
	subscriber message.Subscriber,
	logger *logging.Logger,
	svc *notification.Service,
) {
	router.AddConsumerHandler("notif_order_created", common.EventOrderCreated, subscriber,
		func(msg *message.Message) error {
			var ev orderevent.OrderCreated
			if err := json.Unmarshal(msg.Payload, &ev); err != nil {
				logger.Warn("skipping malformed order created event (notification)", "error", err)
				return nil
			}
			svc.Send(msg.Context(), notification.Notification{
				Title: "New Order Received",
				Body:  fmt.Sprintf("Order #%s received", ev.OrderNumber),
				URL:   "/kitchen",
				Metadata: map[string]string{
					"role":          "kitchen",
					"restaurant_id": string(ev.RestaurantID),
				},
			})
			return nil
		},
	)

	router.AddConsumerHandler("notif_order_preparing", common.EventOrderPreparing, subscriber,
		func(msg *message.Message) error {
			var ev orderevent.OrderPreparing
			if err := json.Unmarshal(msg.Payload, &ev); err != nil {
				logger.Warn("skipping malformed order preparing event (notification)", "error", err)
				return nil
			}
			svc.Send(msg.Context(), notification.Notification{
				Title: "Order Update",
				Body:  "Your order is being prepared",
				URL:   fmt.Sprintf("/order/%s", ev.OrderNumber),
				Metadata: map[string]string{
					"role":         "customer",
					"order_number": string(ev.OrderNumber),
				},
			})
			return nil
		},
	)

	router.AddConsumerHandler("notif_order_ready", common.EventOrderReady, subscriber,
		func(msg *message.Message) error {
			var ev orderevent.OrderReady
			if err := json.Unmarshal(msg.Payload, &ev); err != nil {
				logger.Warn("skipping malformed order ready event (notification)", "error", err)
				return nil
			}
			svc.Send(msg.Context(), notification.Notification{
				Title: "Order Ready",
				Body:  "Your order is ready for pickup! 🎉",
				URL:   fmt.Sprintf("/order/%s", ev.OrderNumber),
				Metadata: map[string]string{
					"role":         "customer",
					"order_number": string(ev.OrderNumber),
				},
			})
			return nil
		},
	)

	router.AddConsumerHandler("notif_order_completed", common.EventOrderCompleted, subscriber,
		func(msg *message.Message) error {
			var ev orderevent.OrderCompleted
			if err := json.Unmarshal(msg.Payload, &ev); err != nil {
				logger.Warn("skipping malformed order completed event (notification)", "error", err)
				return nil
			}
			svc.Send(msg.Context(), notification.Notification{
				Title: "Order Complete",
				Body:  "Thank you for your order!",
				URL:   fmt.Sprintf("/order/%s", ev.OrderNumber),
				Metadata: map[string]string{
					"role":         "customer",
					"order_number": string(ev.OrderNumber),
				},
			})
			return nil
		},
	)
}
