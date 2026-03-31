package postgres

import placesAdapters "bitmerchant/internal/places/adapters"

type SessionRestaurantVisitRepository = placesAdapters.PostgresVisitRepository

var NewSessionRestaurantVisitRepository = placesAdapters.NewPostgresVisitRepository
