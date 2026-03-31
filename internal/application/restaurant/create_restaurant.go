package restaurant

import restCmd "bitmerchant/internal/restaurant/app/command"

type CreateRestaurantRequest = restCmd.CreateRestaurantRequest
type CreateRestaurantUseCase = restCmd.CreateRestaurantUseCase

var NewCreateRestaurantUseCase = restCmd.NewCreateRestaurantUseCase
