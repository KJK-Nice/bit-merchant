package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/notification/webpush"

	"github.com/labstack/echo/v4"
)

// PushHandler handles push subscription registration from browser clients.
type PushHandler struct {
	repo   webpush.Repository
	logger *slog.Logger
}

func NewPushHandler(repo webpush.Repository, logger *slog.Logger) *PushHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &PushHandler{repo: repo, logger: logger}
}

type subscribeRequest struct {
	Endpoint    string          `json:"endpoint"`
	Keys        json.RawMessage `json:"keys"`
	OrderNumber string          `json:"orderNumber"` // customer only
}

type pushKeys struct {
	Auth   string `json:"auth"`
	P256dh string `json:"p256dh"`
}

// SubscribeCustomer stores a customer push subscription linked to an order number.
func (h *PushHandler) SubscribeCustomer(c echo.Context) error {
	var req subscribeRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("push subscribe: invalid request", "role", "customer", "error", err)
		return c.String(http.StatusBadRequest, "invalid request")
	}
	if req.Endpoint == "" || req.OrderNumber == "" {
		h.logger.Warn("push subscribe: missing endpoint or orderNumber", "role", "customer")
		return c.String(http.StatusBadRequest, "endpoint and orderNumber required")
	}

	var keys pushKeys
	if err := json.Unmarshal(req.Keys, &keys); err != nil || keys.Auth == "" || keys.P256dh == "" {
		h.logger.Warn("push subscribe: missing or malformed keys", "role", "customer")
		return c.String(http.StatusBadRequest, "keys.auth and keys.p256dh required")
	}

	sub := &webpush.Subscription{
		Role:      "customer",
		Endpoint:  req.Endpoint,
		AuthKey:   keys.Auth,
		P256DHKey: keys.P256dh,
	}
	if err := h.repo.Upsert(sub); err != nil {
		h.logger.Error("push subscribe: repo upsert failed", "role", "customer", "error", err)
		return c.String(http.StatusInternalServerError, "failed to save subscription")
	}
	if err := h.repo.AddScope(sub.ID, webpush.ScopeTypeOrder, req.OrderNumber); err != nil {
		h.logger.Error("push subscribe: add scope failed", "role", "customer", "order_number", req.OrderNumber, "error", err)
		return c.String(http.StatusInternalServerError, "failed to save subscription scope")
	}
	h.logger.Info("push subscribe stored",
		"role", "customer",
		"order_number", req.OrderNumber,
		"endpoint", req.Endpoint,
	)
	return c.NoContent(http.StatusCreated)
}

// SubscribeKitchen stores a kitchen push subscription linked to the staff's restaurant.
func (h *PushHandler) SubscribeKitchen(c echo.Context) error {
	var req subscribeRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("push subscribe: invalid request", "role", "kitchen", "error", err)
		return c.String(http.StatusBadRequest, "invalid request")
	}
	if req.Endpoint == "" {
		h.logger.Warn("push subscribe: missing endpoint", "role", "kitchen")
		return c.String(http.StatusBadRequest, "endpoint required")
	}

	var keys pushKeys
	if err := json.Unmarshal(req.Keys, &keys); err != nil || keys.Auth == "" || keys.P256dh == "" {
		h.logger.Warn("push subscribe: missing or malformed keys", "role", "kitchen")
		return c.String(http.StatusBadRequest, "keys.auth and keys.p256dh required")
	}

	restaurantID, err := commonhttp.RestaurantIDFromContext(c)
	if err != nil {
		h.logger.Warn("push subscribe: no restaurant context", "role", "kitchen", "error", err)
		return c.String(http.StatusUnauthorized, "restaurant context required")
	}

	sub := &webpush.Subscription{
		Role:      "kitchen",
		Endpoint:  req.Endpoint,
		AuthKey:   keys.Auth,
		P256DHKey: keys.P256dh,
	}
	if err := h.repo.Upsert(sub); err != nil {
		h.logger.Error("push subscribe: repo upsert failed", "role", "kitchen", "error", err)
		return c.String(http.StatusInternalServerError, "failed to save subscription")
	}
	if err := h.repo.AddScope(sub.ID, webpush.ScopeTypeRestaurant, string(restaurantID)); err != nil {
		h.logger.Error("push subscribe: add scope failed", "role", "kitchen", "restaurant_id", string(restaurantID), "error", err)
		return c.String(http.StatusInternalServerError, "failed to save subscription scope")
	}
	h.logger.Info("push subscribe stored",
		"role", "kitchen",
		"restaurant_id", string(restaurantID),
		"endpoint", req.Endpoint,
	)
	return c.NoContent(http.StatusCreated)
}
