package command

import (
	"context"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"log/slog"
)

// UpdateKitchenThresholds sets the per-restaurant kitchen escalation tiers
// (warning / overdue thresholds, in minutes) used by the kitchen board.
type UpdateKitchenThresholds struct {
	RestaurantID   common.RestaurantID
	WarningMinutes int
	OverdueMinutes int
}

type UpdateKitchenThresholdsHandler decorator.CommandHandler[UpdateKitchenThresholds]

type updateKitchenThresholdsHandler struct {
	repo restaurant.Repository
}

func NewUpdateKitchenThresholdsHandler(repo restaurant.Repository, log *slog.Logger, metrics decorator.MetricsClient) UpdateKitchenThresholdsHandler {
	if repo == nil {
		panic("nil restaurant.Repository")
	}
	h := updateKitchenThresholdsHandler{repo: repo}
	return decorator.ApplyCommandDecorators[UpdateKitchenThresholds](h, log, metrics)
}

func (h updateKitchenThresholdsHandler) Handle(ctx context.Context, cmd UpdateKitchenThresholds) error {
	_ = ctx
	if err := restaurant.ValidateKitchenThresholds(cmd.WarningMinutes, cmd.OverdueMinutes); err != nil {
		return err
	}
	rest, err := h.repo.FindByID(cmd.RestaurantID)
	if err != nil {
		return err
	}
	if err := rest.SetKitchenThresholds(cmd.WarningMinutes, cmd.OverdueMinutes); err != nil {
		return err
	}
	return h.repo.Update(rest)
}
