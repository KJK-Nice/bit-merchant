package command

import (
	"context"
	"errors"
	"log/slog"

	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/google/uuid"
)

// ErrNotRestaurantOwner indicates the user is not an owner of the context restaurant.
var ErrNotRestaurantOwner = errors.New("only owners can create restaurants")

// CreateRestaurantUnderOwner creates a new restaurant while the user is acting as owner
// of an existing restaurant (additional location).
type CreateRestaurantUnderOwner struct {
	OwnerUserID              common.UserID
	OwnerContextRestaurantID common.RestaurantID
	Name                     string
}

type CreateRestaurantUnderOwnerHandler decorator.CommandResultHandler[CreateRestaurantUnderOwner, *restaurant.Restaurant]

type createRestaurantUnderOwnerHandler struct {
	memRepo          membership.Repository
	createRestaurant restaurantCmd.CreateRestaurantHandler
}

func NewCreateRestaurantUnderOwnerHandler(memRepo membership.Repository, createRestaurant restaurantCmd.CreateRestaurantHandler, log *slog.Logger, metrics decorator.MetricsClient) CreateRestaurantUnderOwnerHandler {
	if memRepo == nil || createRestaurant == nil {
		panic("nil dependency")
	}
	h := createRestaurantUnderOwnerHandler{memRepo: memRepo, createRestaurant: createRestaurant}
	return decorator.ApplyCommandResultDecorators[CreateRestaurantUnderOwner, *restaurant.Restaurant](h, log, metrics)
}

func (h createRestaurantUnderOwnerHandler) Handle(ctx context.Context, cmd CreateRestaurantUnderOwner) (*restaurant.Restaurant, error) {
	mem, err := h.memRepo.FindByUserAndRestaurant(cmd.OwnerUserID, cmd.OwnerContextRestaurantID)
	if err != nil || mem == nil || mem.Role != common.RoleOwner {
		return nil, ErrNotRestaurantOwner
	}

	rest, err := h.createRestaurant.Handle(ctx, restaurantCmd.CreateRestaurant{Name: cmd.Name})
	if err != nil {
		return nil, err
	}
	newMem, err := membership.NewMembership(
		common.MembershipID(uuid.NewString()),
		cmd.OwnerUserID,
		rest.ID,
		common.RoleOwner,
	)
	if err != nil {
		return nil, err
	}
	if err := h.memRepo.Save(newMem); err != nil {
		return nil, err
	}
	return rest, nil
}
