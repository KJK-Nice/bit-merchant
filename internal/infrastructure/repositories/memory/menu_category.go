package memory

import (
	"errors"
	"sync"

	"bitmerchant/internal/domain"
)

// MemoryMenuCategoryRepository implements MenuCategoryRepository with in-memory storage
type MemoryMenuCategoryRepository struct {
	mu         sync.RWMutex
	categories map[domain.CategoryID]*domain.MenuCategory
}

// NewMemoryMenuCategoryRepository creates a new in-memory menu category repository
func NewMemoryMenuCategoryRepository() *MemoryMenuCategoryRepository {
	return &MemoryMenuCategoryRepository{
		categories: make(map[domain.CategoryID]*domain.MenuCategory),
	}
}

// Save saves a menu category
func (r *MemoryMenuCategoryRepository) Save(category *domain.MenuCategory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categories[category.ID] = category
	return nil
}

// FindByID finds a menu category by ID
func (r *MemoryMenuCategoryRepository) FindByID(id domain.CategoryID) (*domain.MenuCategory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	category, exists := r.categories[id]
	if !exists {
		return nil, errors.New("menu category not found")
	}
	return category, nil
}

// FindByRestaurantID finds all menu categories for a restaurant
func (r *MemoryMenuCategoryRepository) FindByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.MenuCategory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.MenuCategory
	for _, category := range r.categories {
		if category.RestaurantID == restaurantID {
			result = append(result, category)
		}
	}
	return result, nil
}

// Update updates a menu category
func (r *MemoryMenuCategoryRepository) Update(category *domain.MenuCategory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.categories[category.ID]; !exists {
		return errors.New("menu category not found")
	}
	r.categories[category.ID] = category
	return nil
}

// Delete deletes a menu category
func (r *MemoryMenuCategoryRepository) Delete(id domain.CategoryID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.categories[id]; !exists {
		return errors.New("menu category not found")
	}
	delete(r.categories, id)
	return nil
}
