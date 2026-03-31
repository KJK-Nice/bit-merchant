package memory

import restAdapters "bitmerchant/internal/restaurant/adapters"

type MemoryRestaurantRepository = restAdapters.MemoryRestaurantRepository

var NewMemoryRestaurantRepository = restAdapters.NewMemoryRestaurantRepository
