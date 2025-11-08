package memory

import (
	"errors"
	"sync"

	"bitmerchant/internal/domain"
)

// MemoryMenuItemRepository implements MenuItemRepository with in-memory storage
type MemoryMenuItemRepository struct {
	mu    sync.RWMutex
	items map[domain.ItemID]*domain.MenuItem
}

// NewMemoryMenuItemRepository creates a new in-memory menu item repository
func NewMemoryMenuItemRepository() *MemoryMenuItemRepository {
	return &MemoryMenuItemRepository{
		items: make(map[domain.ItemID]*domain.MenuItem),
	}
}

// Save saves a menu item
func (r *MemoryMenuItemRepository) Save(item *domain.MenuItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = item
	return nil
}

// FindByID finds a menu item by ID
func (r *MemoryMenuItemRepository) FindByID(id domain.ItemID) (*domain.MenuItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, exists := r.items[id]
	if !exists {
		return nil, errors.New("menu item not found")
	}
	return item, nil
}

// FindByCategoryID finds all menu items for a category
func (r *MemoryMenuItemRepository) FindByCategoryID(categoryID domain.CategoryID) ([]*domain.MenuItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.MenuItem
	for _, item := range r.items {
		if item.CategoryID == categoryID {
			result = append(result, item)
		}
	}
	return result, nil
}

// FindByRestaurantID finds all menu items for a restaurant
func (r *MemoryMenuItemRepository) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.MenuItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.MenuItem
	for _, item := range r.items {
		if item.RestaurantID == restaurantID {
			result = append(result, item)
		}
	}
	return result, nil
}

// FindAvailableByRestaurantID finds available menu items for a restaurant
func (r *MemoryMenuItemRepository) FindAvailableByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.MenuItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.MenuItem
	for _, item := range r.items {
		if item.RestaurantID == restaurantID && item.IsAvailable {
			result = append(result, item)
		}
	}
	return result, nil
}

// Update updates a menu item
func (r *MemoryMenuItemRepository) Update(item *domain.MenuItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[item.ID]; !exists {
		return errors.New("menu item not found")
	}
	r.items[item.ID] = item
	return nil
}

// Delete deletes a menu item
func (r *MemoryMenuItemRepository) Delete(id domain.ItemID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[id]; !exists {
		return errors.New("menu item not found")
	}
	delete(r.items, id)
	return nil
}

// CountByRestaurantID counts menu items with photos for a restaurant
func (r *MemoryMenuItemRepository) CountByRestaurantID(restaurantID domain.RestaurantID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, item := range r.items {
		if item.RestaurantID == restaurantID && item.PhotoURL != "" {
			count++
		}
	}
	return count, nil
}
