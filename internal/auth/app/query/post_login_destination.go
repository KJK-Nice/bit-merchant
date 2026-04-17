package query

import (
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/common"
)

// PostLoginRestaurantContext picks default restaurant and redirect after login
// when the user has one or more memberships.
func PostLoginRestaurantContext(memberships []*membership.Membership) (*common.RestaurantID, string) {
	if len(memberships) == 1 {
		redirect := "/dashboard"
		if memberships[0].Role == common.RoleKitchenStaff {
			redirect = "/kitchen"
		}
		return &memberships[0].RestaurantID, redirect
	}
	if len(memberships) > 1 {
		return nil, "/auth/select-restaurant"
	}
	return nil, "/dashboard"
}
