package domain

import (
	"bitmerchant/internal/common"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

type RestaurantID = common.RestaurantID
type Restaurant = restaurant.Restaurant

var NewRestaurant = restaurant.NewRestaurant
var ValidateRestaurantName = restaurant.ValidateRestaurantName
var ValidateTableCount = restaurant.ValidateTableCount

const MinTableCount = restaurant.MinTableCount
const MaxTableCount = restaurant.MaxTableCount
