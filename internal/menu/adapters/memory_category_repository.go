package adapters

import (
	"errors"
	"sync"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

type MemoryCategoryRepository struct {
	mu         sync.RWMutex
	categories map[common.CategoryID]*menu.MenuCategory
}

func NewMemoryCategoryRepository() *MemoryCategoryRepository {
	return &MemoryCategoryRepository{
		categories: make(map[common.CategoryID]*menu.MenuCategory),
	}
}

func (r *MemoryCategoryRepository) Save(category *menu.MenuCategory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categories[category.ID] = category
	return nil
}

func (r *MemoryCategoryRepository) FindByID(id common.CategoryID) (*menu.MenuCategory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cat, exists := r.categories[id]
	if !exists {
		return nil, errors.New("menu category not found")
	}
	return cat, nil
}

func (r *MemoryCategoryRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*menu.MenuCategory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*menu.MenuCategory
	for _, cat := range r.categories {
		if cat.RestaurantID == restaurantID {
			result = append(result, cat)
		}
	}
	return result, nil
}

func (r *MemoryCategoryRepository) Update(category *menu.MenuCategory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.categories[category.ID]; !exists {
		return errors.New("menu category not found")
	}
	r.categories[category.ID] = category
	return nil
}

func (r *MemoryCategoryRepository) Delete(id common.CategoryID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.categories[id]; !exists {
		return errors.New("menu category not found")
	}
	delete(r.categories, id)
	return nil
}
