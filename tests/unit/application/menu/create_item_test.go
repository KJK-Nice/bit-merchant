package menu_test

import (
	"bitmerchant/internal/common"
	menuCmd "bitmerchant/internal/menu/app/command"
	"bitmerchant/internal/menu/domain/menu"
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockMenuItemRepository struct {
	mock.Mock
}

func (m *MockMenuItemRepository) Save(i *menu.MenuItem) error {
	args := m.Called(i)
	return args.Error(0)
}

func (m *MockMenuItemRepository) FindByCategoryID(id common.CategoryID) ([]*menu.MenuItem, error) {
	args := m.Called(id)
	return args.Get(0).([]*menu.MenuItem), args.Error(1)
}

func (m *MockMenuItemRepository) FindByID(id common.ItemID) (*menu.MenuItem, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*menu.MenuItem), args.Error(1)
}

func (m *MockMenuItemRepository) FindByRestaurantID(id common.RestaurantID) ([]*menu.MenuItem, error) {
	args := m.Called(id)
	return args.Get(0).([]*menu.MenuItem), args.Error(1)
}

func (m *MockMenuItemRepository) FindAvailableByRestaurantID(restaurantID common.RestaurantID) ([]*menu.MenuItem, error) {
	args := m.Called(restaurantID)
	return args.Get(0).([]*menu.MenuItem), args.Error(1)
}

func (m *MockMenuItemRepository) Update(item *menu.MenuItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockMenuItemRepository) Delete(id common.ItemID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockMenuItemRepository) CountByRestaurantID(restaurantID common.RestaurantID) (int, error) {
	args := m.Called(restaurantID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockMenuItemRepository) ReorderItemsInCategory(restaurantID common.RestaurantID, categoryID common.CategoryID, orderedItemIDs []common.ItemID) error {
	args := m.Called(restaurantID, categoryID, orderedItemIDs)
	return args.Error(0)
}

func TestCreateMenuItemHandler_Handle(t *testing.T) {
	t.Run("successfully creates item", func(t *testing.T) {
		repo := new(MockMenuItemRepository)
		useCase := menuCmd.NewCreateMenuItemHandler(repo, nil, nil)

		restaurantID := common.RestaurantID("rest-1")
		categoryID := common.CategoryID("cat-1")

		repo.On("FindByCategoryID", categoryID).Return([]*menu.MenuItem{}, nil)
		repo.On("Save", mock.MatchedBy(func(i *menu.MenuItem) bool {
			return i.Name == "Burger" && i.RestaurantID == restaurantID && i.CategoryID == categoryID && i.Price == 12.50 && i.DisplayOrder == 0
		})).Return(nil)

		req := menuCmd.CreateMenuItem{
			RestaurantID: restaurantID,
			CategoryID:   categoryID,
			Name:         "Burger",
			Description:  "Delicious burger",
			Price:        12.50,
			Available:    true,
		}

		resp, err := useCase.Handle(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Burger", resp.Name)
		assert.Equal(t, 12.50, resp.Price)
		repo.AssertExpectations(t)
	})

	t.Run("fails with invalid price", func(t *testing.T) {
		repo := new(MockMenuItemRepository)
		useCase := menuCmd.NewCreateMenuItemHandler(repo, nil, nil)

		req := menuCmd.CreateMenuItem{
			RestaurantID: "rest-1",
			CategoryID:   "cat-1",
			Name:         "Burger",
			Price:        -10.0,
		}

		resp, err := useCase.Handle(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "price must be greater than 0")
	})
}
