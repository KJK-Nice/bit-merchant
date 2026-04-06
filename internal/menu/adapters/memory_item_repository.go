package adapters

import (
	"errors"
	"sync"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

type MemoryItemRepository struct {
	mu    sync.RWMutex
	items map[common.ItemID]*menu.MenuItem
}

func NewMemoryItemRepository() *MemoryItemRepository {
	return &MemoryItemRepository{
		items: make(map[common.ItemID]*menu.MenuItem),
	}
}

func (r *MemoryItemRepository) Save(item *menu.MenuItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = item
	return nil
}

func (r *MemoryItemRepository) FindByID(id common.ItemID) (*menu.MenuItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, exists := r.items[id]
	if !exists {
		return nil, errors.New("menu item not found")
	}
	return item, nil
}

func (r *MemoryItemRepository) FindByCategoryID(categoryID common.CategoryID) ([]*menu.MenuItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*menu.MenuItem
	for _, item := range r.items {
		if item.CategoryID == categoryID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *MemoryItemRepository) FindByRestaurantID(restaurantID common.RestaurantID) ([]*menu.MenuItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*menu.MenuItem
	for _, item := range r.items {
		if item.RestaurantID == restaurantID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *MemoryItemRepository) FindAvailableByRestaurantID(restaurantID common.RestaurantID) ([]*menu.MenuItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*menu.MenuItem
	for _, item := range r.items {
		if item.RestaurantID == restaurantID && item.IsAvailable {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *MemoryItemRepository) Update(item *menu.MenuItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[item.ID]; !exists {
		return errors.New("menu item not found")
	}
	r.items[item.ID] = item
	return nil
}

func (r *MemoryItemRepository) Delete(id common.ItemID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[id]; !exists {
		return errors.New("menu item not found")
	}
	delete(r.items, id)
	return nil
}

func (r *MemoryItemRepository) CountByRestaurantID(restaurantID common.RestaurantID) (int, error) {
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

func (r *MemoryItemRepository) ReorderItemsInCategory(restaurantID common.RestaurantID, categoryID common.CategoryID, orderedItemIDs []common.ItemID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var inCat []*menu.MenuItem
	for _, item := range r.items {
		if item.RestaurantID == restaurantID && item.CategoryID == categoryID {
			inCat = append(inCat, item)
		}
	}
	if len(orderedItemIDs) != len(inCat) {
		return errors.New("item list does not match category")
	}
	want := make(map[common.ItemID]struct{}, len(inCat))
	for _, it := range inCat {
		want[it.ID] = struct{}{}
	}
	for _, id := range orderedItemIDs {
		if _, ok := want[id]; !ok {
			return errors.New("invalid item in reorder list")
		}
		delete(want, id)
	}
	if len(want) != 0 {
		return errors.New("item list does not match category")
	}
	for i, id := range orderedItemIDs {
		r.items[id].DisplayOrder = i
	}
	return nil
}
