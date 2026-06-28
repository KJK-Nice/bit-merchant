package sse

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/interfaces/templates/components"
	"bitmerchant/internal/ordering/app/event"
	"bitmerchant/internal/ordering/domain/order"
)

// serviceRequestSubtext builds the "Table 7 · Maya · 3:04 PM" line for an alert.
func serviceRequestSubtext(tableLabel, customerName string, hourMinute string) string {
	parts := make([]string, 0, 3)
	if tableLabel != "" {
		parts = append(parts, "Table "+tableLabel)
	}
	if customerName != "" {
		parts = append(parts, customerName)
	}
	parts = append(parts, hourMinute)
	return strings.Join(parts, " · ")
}

// broadcastServiceAlert appends an alert tile to the FOH server view's
// #service-requests strip. The domID is unique per request instant so repeated
// (post-throttle) requests stack rather than overwrite.
func broadcastServiceAlert(ctx context.Context, logger *logging.Logger, sse *commonhttp.SSEHandler, domID, heading, subtext, tone string) {
	var buf bytes.Buffer
	if err := components.ServiceRequestAlert(domID, heading, subtext, tone).Render(ctx, &buf); err != nil {
		logger.Error("service alert: render failed", "error", err)
		return
	}
	msg := commonhttp.FormatDatastarPatch(buf.String(), "#service-requests", "append")
	sse.Broadcast(commonhttp.TopicServer, msg)
}

// ServerCalledHandler surfaces a "call server" request on the FOH view and
// reflects the request back to the customer's status stream.
type ServerCalledHandler struct {
	logger *logging.Logger
	sse    *commonhttp.SSEHandler
	repo   order.Repository
}

func NewServerCalledHandler(logger *logging.Logger, sse *commonhttp.SSEHandler, repo order.Repository) *ServerCalledHandler {
	return &ServerCalledHandler{logger: logger, sse: sse, repo: repo}
}

func (h *ServerCalledHandler) Handle(ctx context.Context, ev event.ServerCalled) error {
	h.logger.Info("Server called", "orderID", ev.OrderID)
	domID := fmt.Sprintf("service-req-server-%s-%d", ev.OrderID, ev.CalledAt.Unix())
	subtext := serviceRequestSubtext(ev.TableLabel, ev.CustomerName, ev.CalledAt.Format("3:04 PM"))
	broadcastServiceAlert(ctx, h.logger, h.sse, domID, "🔔 Call server", subtext, "server")

	if o, err := h.repo.FindByID(ev.OrderID); err == nil && o != nil {
		pushView(ctx, h.logger, h.sse, h.repo, o)
	}
	return nil
}

// BillRequestedHandler surfaces a "request bill" request on the FOH view and
// reflects the request back to the customer's status stream.
type BillRequestedHandler struct {
	logger *logging.Logger
	sse    *commonhttp.SSEHandler
	repo   order.Repository
}

func NewBillRequestedHandler(logger *logging.Logger, sse *commonhttp.SSEHandler, repo order.Repository) *BillRequestedHandler {
	return &BillRequestedHandler{logger: logger, sse: sse, repo: repo}
}

func (h *BillRequestedHandler) Handle(ctx context.Context, ev event.BillRequested) error {
	h.logger.Info("Bill requested", "orderID", ev.OrderID)
	domID := fmt.Sprintf("service-req-bill-%s-%d", ev.OrderID, ev.RequestedAt.Unix())
	subtext := serviceRequestSubtext(ev.TableLabel, ev.CustomerName, ev.RequestedAt.Format("3:04 PM"))
	broadcastServiceAlert(ctx, h.logger, h.sse, domID, "🧾 Bill requested", subtext, "bill")

	if o, err := h.repo.FindByID(ev.OrderID); err == nil && o != nil {
		pushView(ctx, h.logger, h.sse, h.repo, o)
	}
	return nil
}
