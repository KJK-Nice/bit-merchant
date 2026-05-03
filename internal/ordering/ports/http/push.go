package http

import (
	"encoding/json"
	"net/http"

	commonhttp "bitmerchant/internal/common/http"
	"bitmerchant/internal/notification/webpush"

	"github.com/labstack/echo/v4"
)

// PushHandler handles push subscription registration from browser clients.
type PushHandler struct {
	repo webpush.Repository
}

func NewPushHandler(repo webpush.Repository) *PushHandler {
	return &PushHandler{repo: repo}
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
		return c.String(http.StatusBadRequest, "invalid request")
	}
	if req.Endpoint == "" || req.OrderNumber == "" {
		return c.String(http.StatusBadRequest, "endpoint and orderNumber required")
	}

	var keys pushKeys
	if err := json.Unmarshal(req.Keys, &keys); err != nil || keys.Auth == "" || keys.P256dh == "" {
		return c.String(http.StatusBadRequest, "keys.auth and keys.p256dh required")
	}

	sub := &webpush.Subscription{
		Role:        "customer",
		OrderNumber: req.OrderNumber,
		Endpoint:    req.Endpoint,
		AuthKey:     keys.Auth,
		P256DHKey:   keys.P256dh,
	}
	if err := h.repo.Upsert(sub); err != nil {
		return c.String(http.StatusInternalServerError, "failed to save subscription")
	}
	return c.NoContent(http.StatusCreated)
}

// SubscribeKitchen stores a kitchen push subscription linked to the staff's restaurant.
func (h *PushHandler) SubscribeKitchen(c echo.Context) error {
	var req subscribeRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "invalid request")
	}
	if req.Endpoint == "" {
		return c.String(http.StatusBadRequest, "endpoint required")
	}

	var keys pushKeys
	if err := json.Unmarshal(req.Keys, &keys); err != nil || keys.Auth == "" || keys.P256dh == "" {
		return c.String(http.StatusBadRequest, "keys.auth and keys.p256dh required")
	}

	restaurantID, err := commonhttp.RestaurantIDFromContext(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, "restaurant context required")
	}

	sub := &webpush.Subscription{
		Role:         "kitchen",
		RestaurantID: restaurantID,
		Endpoint:     req.Endpoint,
		AuthKey:      keys.Auth,
		P256DHKey:    keys.P256dh,
	}
	if err := h.repo.Upsert(sub); err != nil {
		return c.String(http.StatusInternalServerError, "failed to save subscription")
	}
	return c.NoContent(http.StatusCreated)
}
