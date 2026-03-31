package adapters

import (
	"errors"
	"sync"

	"bitmerchant/internal/common"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

type MemoryRestaurantRepository struct {
	mu          sync.RWMutex
	restaurants map[common.RestaurantID]*restaurant.Restaurant
}

func NewMemoryRestaurantRepository() *MemoryRestaurantRepository {
	return &MemoryRestaurantRepository{
		restaurants: make(map[common.RestaurantID]*restaurant.Restaurant),
	}
}

func (r *MemoryRestaurantRepository) Save(rest *restaurant.Restaurant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *rest
	r.restaurants[rest.ID] = &cp
	return nil
}

func (r *MemoryRestaurantRepository) FindByID(id common.RestaurantID) (*restaurant.Restaurant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rest, exists := r.restaurants[id]
	if !exists {
		return nil, errors.New("restaurant not found")
	}
	return rest, nil
}

func (r *MemoryRestaurantRepository) Update(rest *restaurant.Restaurant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.restaurants[rest.ID]; !exists {
		return errors.New("restaurant not found")
	}
	r.restaurants[rest.ID] = rest
	return nil
}
