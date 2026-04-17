package command

import "bitmerchant/internal/common"

// RegistrationOutcome is returned after signup flows that establish restaurant context.
type RegistrationOutcome struct {
	RestaurantID *common.RestaurantID
	Redirect     string
}
