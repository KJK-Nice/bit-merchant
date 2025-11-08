package memory

import (
	"errors"
	"sync"

	"bitmerchant/internal/domain"
)

// MemoryRestaurantRepository implements RestaurantRepository with in-memory storage
type MemoryRestaurantRepository struct {
	mu          sync.RWMutex
	restaurants map[domain.RestaurantID]*domain.Restaurant
}

// NewMemoryRestaurantRepository creates a new in-memory restaurant repository
func NewMemoryRestaurantRepository() *MemoryRestaurantRepository {
	return &MemoryRestaurantRepository{
		restaurants: make(map[domain.RestaurantID]*domain.Restaurant),
	}
}

// Save saves a restaurant
func (r *MemoryRestaurantRepository) Save(restaurant *domain.Restaurant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.restaurants[restaurant.ID] = restaurant
	return nil
}

// FindByID finds a restaurant by ID
func (r *MemoryRestaurantRepository) FindByID(id domain.RestaurantID) (*domain.Restaurant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	restaurant, exists := r.restaurants[id]
	if !exists {
		return nil, errors.New("restaurant not found")
	}
	return restaurant, nil
}

// Update updates a restaurant
func (r *MemoryRestaurantRepository) Update(restaurant *domain.Restaurant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.restaurants[restaurant.ID]; !exists {
		return errors.New("restaurant not found")
	}
	r.restaurants[restaurant.ID] = restaurant
	return nil
}
