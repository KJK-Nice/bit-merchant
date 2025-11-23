package menu_test

import (
	"context"
	"testing"

	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMenuItemRepository struct {
	mock.Mock
}

func (m *MockMenuItemRepository) Save(i *domain.MenuItem) error {
	args := m.Called(i)
	return args.Error(0)
}

func (m *MockMenuItemRepository) FindByCategoryID(id domain.CategoryID) ([]*domain.MenuItem, error) {
	args := m.Called(id)
	return args.Get(0).([]*domain.MenuItem), args.Error(1)
}

func (m *MockMenuItemRepository) FindByID(id domain.ItemID) (*domain.MenuItem, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MenuItem), args.Error(1)
}

func (m *MockMenuItemRepository) FindByRestaurantID(id domain.RestaurantID) ([]*domain.MenuItem, error) {
	args := m.Called(id)
	return args.Get(0).([]*domain.MenuItem), args.Error(1)
}

func (m *MockMenuItemRepository) FindAvailableByRestaurantID(restaurantID domain.RestaurantID) ([]*domain.MenuItem, error) {
	args := m.Called(restaurantID)
	return args.Get(0).([]*domain.MenuItem), args.Error(1)
}

func (m *MockMenuItemRepository) Update(item *domain.MenuItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockMenuItemRepository) Delete(id domain.ItemID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockMenuItemRepository) CountByRestaurantID(restaurantID domain.RestaurantID) (int, error) {
	args := m.Called(restaurantID)
	return args.Get(0).(int), args.Error(1)
}

func TestCreateMenuItemUseCase_Execute(t *testing.T) {
	t.Run("successfully creates item", func(t *testing.T) {
		repo := new(MockMenuItemRepository)
		useCase := menu.NewCreateMenuItemUseCase(repo)

		restaurantID := domain.RestaurantID("rest-1")
		categoryID := domain.CategoryID("cat-1")

		repo.On("Save", mock.MatchedBy(func(i *domain.MenuItem) bool {
			return i.Name == "Burger" && i.RestaurantID == restaurantID && i.CategoryID == categoryID && i.Price == 12.50
		})).Return(nil)

		req := menu.CreateMenuItemRequest{
			RestaurantID: restaurantID,
			CategoryID:   categoryID,
			Name:         "Burger",
			Description:  "Delicious burger",
			Price:        12.50,
			Available:    true,
		}

		resp, err := useCase.Execute(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Burger", resp.Name)
		assert.Equal(t, 12.50, resp.Price)
		repo.AssertExpectations(t)
	})

	t.Run("fails with invalid price", func(t *testing.T) {
		repo := new(MockMenuItemRepository)
		useCase := menu.NewCreateMenuItemUseCase(repo)

		req := menu.CreateMenuItemRequest{
			RestaurantID: "rest-1",
			CategoryID:   "cat-1",
			Name:         "Burger",
			Price:        -10.0,
		}

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "price must be greater than 0")
	})
}
