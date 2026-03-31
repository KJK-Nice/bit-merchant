package postgres

import restAdapters "bitmerchant/internal/restaurant/adapters"

type RestaurantRepository = restAdapters.PostgresRestaurantRepository

var NewRestaurantRepository = restAdapters.NewPostgresRestaurantRepository
