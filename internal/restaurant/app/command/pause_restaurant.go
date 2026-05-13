package command

import (
	"context"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"log/slog"
)

// PauseRestaurant applies a temporary quick-pause window without closing
// the restaurant. Duration <= 0 acts as Resume — clears any active pause.
type PauseRestaurant struct {
	RestaurantID common.RestaurantID
	Duration     time.Duration
}

type PauseRestaurantHandler decorator.CommandHandler[PauseRestaurant]

type pauseRestaurantHandler struct {
	repo restaurant.Repository
	now  func() time.Time
}

func NewPauseRestaurantHandler(repo restaurant.Repository, log *slog.Logger, metrics decorator.MetricsClient) PauseRestaurantHandler {
	if repo == nil {
		panic("nil restaurant.Repository")
	}
	h := pauseRestaurantHandler{repo: repo, now: time.Now}
	return decorator.ApplyCommandDecorators[PauseRestaurant](h, log, metrics)
}

func (h pauseRestaurantHandler) Handle(ctx context.Context, cmd PauseRestaurant) error {
	_ = ctx
	r, err := h.repo.FindByID(cmd.RestaurantID)
	if err != nil {
		return err
	}
	if cmd.Duration <= 0 {
		r.Resume()
	} else {
		if err := r.Pause(h.now(), cmd.Duration); err != nil {
			return err
		}
	}
	return h.repo.Save(r)
}
