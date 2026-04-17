package command

import (
	"context"
	"fmt"
	"strings"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"log/slog"
)

// ToggleRestaurantOpen opens or closes a restaurant for service.
type ToggleRestaurantOpen struct {
	RestaurantID   common.RestaurantID
	ClosedMessage  string
	ReopeningHours string
}

type ToggleRestaurantOpenHandler decorator.CommandResultHandler[ToggleRestaurantOpen, bool]

type toggleRestaurantOpenHandler struct {
	repo restaurant.Repository
}

func NewToggleRestaurantOpenHandler(repo restaurant.Repository, log *slog.Logger, metrics decorator.MetricsClient) ToggleRestaurantOpenHandler {
	if repo == nil {
		panic("nil restaurant.Repository")
	}
	h := toggleRestaurantOpenHandler{repo: repo}
	return decorator.ApplyCommandResultDecorators[ToggleRestaurantOpen, bool](h, log, metrics)
}

func (h toggleRestaurantOpenHandler) Handle(ctx context.Context, cmd ToggleRestaurantOpen) (bool, error) {
	_ = ctx
	r, err := h.repo.FindByID(cmd.RestaurantID)
	if err != nil {
		return false, err
	}

	closedMessage := strings.TrimSpace(cmd.ClosedMessage)
	reopeningHours := strings.TrimSpace(cmd.ReopeningHours)
	if err := restaurant.ValidateDescription(closedMessage); err != nil {
		return false, fmt.Errorf("closed message: %w", err)
	}
	if err := restaurant.ValidateDescription(reopeningHours); err != nil {
		return false, fmt.Errorf("reopening hours: %w", err)
	}

	if r.IsOpen {
		r.Close(closedMessage, reopeningHours)
	} else {
		r.Open()
	}

	if err := h.repo.Save(r); err != nil {
		return false, err
	}

	return r.IsOpen, nil
}
