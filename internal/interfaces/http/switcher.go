package http

import (
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"

	"bitmerchant/internal/interfaces/templates/layouts"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"context"

	"github.com/labstack/echo/v4"
)

// RestaurantSwitchOptionsFromMemberships builds switcher rows using the same display rules as the restaurant picker.
func RestaurantSwitchOptionsFromMemberships(ctx context.Context, memberships []*membership.Membership, restaurantRepo restaurant.Repository) []layouts.RestaurantSwitchOption {
	options := make([]layouts.RestaurantSwitchOption, 0, len(memberships))
	for _, membership := range memberships {
		option := layouts.RestaurantSwitchOption{
			RestaurantID: string(membership.RestaurantID),
			Role:         string(membership.Role),
			DisplayName:  string(membership.RestaurantID),
		}
		if restaurantRepo != nil {
			if rest, restErr := restaurantRepo.FindByID(membership.RestaurantID); restErr == nil && rest != nil && rest.Name != "" {
				option.DisplayName = rest.Name
			}
		}
		options = append(options, option)
	}
	return options
}

// ActiveRestaurantRoleForMemberships returns the member role for the active restaurant ID, or "".
func ActiveRestaurantRoleForMemberships(activeRestaurantID string, memberships []*membership.Membership) string {
	if activeRestaurantID == "" {
		return ""
	}
	for _, m := range memberships {
		if string(m.RestaurantID) == activeRestaurantID {
			return string(m.Role)
		}
	}
	return ""
}

// RestaurantSwitcherData loads memberships for the current user and returns switcher options, active role,
// and whether the user may create a restaurant (owner of the active restaurant).
func RestaurantSwitcherData(c echo.Context, membershipRepo membership.Repository, restaurantRepo restaurant.Repository) ([]layouts.RestaurantSwitchOption, string, bool, error) {
	user, ok := getAuthenticatedUser(c)
	if !ok || user == nil || membershipRepo == nil {
		return nil, "", false, nil
	}
	memberships, err := membershipRepo.FindByUserID(user.ID)
	if err != nil {
		return nil, "", false, err
	}
	opts := RestaurantSwitchOptionsFromMemberships(c.Request().Context(), memberships, restaurantRepo)
	var activeRole string
	if rid, rerr := getRestaurantIDFromContext(c); rerr == nil {
		activeRole = ActiveRestaurantRoleForMemberships(string(rid), memberships)
	}
	canCreate := activeRole == string(common.RoleOwner)
	return opts, activeRole, canCreate, nil
}
