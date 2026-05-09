package command

import (
	"context"
	"log/slog"

	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"

	"github.com/google/uuid"
)

// CompleteSignupNewRestaurant creates the first restaurant and owner membership during signup.
type CompleteSignupNewRestaurant struct {
	OwnerUserID    common.UserID
	RestaurantName string
	CurrencyCode   string
}

type CompleteSignupNewRestaurantHandler decorator.CommandResultHandler[CompleteSignupNewRestaurant, RegistrationOutcome]

type completeSignupNewRestaurantHandler struct {
	memRepo          membership.Repository
	createRestaurant restaurantCmd.CreateRestaurantHandler
}

func NewCompleteSignupNewRestaurantHandler(memRepo membership.Repository, createRestaurant restaurantCmd.CreateRestaurantHandler, log *slog.Logger, metrics decorator.MetricsClient) CompleteSignupNewRestaurantHandler {
	if memRepo == nil || createRestaurant == nil {
		panic("nil dependency")
	}
	h := completeSignupNewRestaurantHandler{memRepo: memRepo, createRestaurant: createRestaurant}
	return decorator.ApplyCommandResultDecorators[CompleteSignupNewRestaurant, RegistrationOutcome](h, log, metrics)
}

func (h completeSignupNewRestaurantHandler) Handle(ctx context.Context, cmd CompleteSignupNewRestaurant) (RegistrationOutcome, error) {
	rest, err := h.createRestaurant.Handle(ctx, restaurantCmd.CreateRestaurant{Name: cmd.RestaurantName, CurrencyCode: cmd.CurrencyCode})
	if err != nil {
		return RegistrationOutcome{}, err
	}
	mem, err := membership.NewMembership(
		common.MembershipID(uuid.NewString()),
		cmd.OwnerUserID,
		rest.ID,
		common.RoleOwner,
	)
	if err != nil {
		return RegistrationOutcome{}, err
	}
	if err := h.memRepo.Save(mem); err != nil {
		return RegistrationOutcome{}, err
	}
	rid := rest.ID
	return RegistrationOutcome{RestaurantID: &rid, Redirect: "/dashboard"}, nil
}
