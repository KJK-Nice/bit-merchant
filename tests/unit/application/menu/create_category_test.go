package menu_test

import (
	"context"
	"testing"

	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMenuCategoryRepository struct {
	mock.Mock
}

func (m *MockMenuCategoryRepository) Save(c *domain.MenuCategory) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *MockMenuCategoryRepository) FindByRestaurantID(id domain.RestaurantID) ([]*domain.MenuCategory, error) {
	args := m.Called(id)
	return args.Get(0).([]*domain.MenuCategory), args.Error(1)
}

func (m *MockMenuCategoryRepository) FindByID(id domain.CategoryID) (*domain.MenuCategory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MenuCategory), args.Error(1)
}

func (m *MockMenuCategoryRepository) Update(c *domain.MenuCategory) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *MockMenuCategoryRepository) Delete(id domain.CategoryID) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestCreateMenuCategoryUseCase_Execute(t *testing.T) {
	t.Run("successfully creates category", func(t *testing.T) {
		repo := new(MockMenuCategoryRepository)
		useCase := menu.NewCreateMenuCategoryUseCase(repo)

		restaurantID := domain.RestaurantID("rest-1")

		repo.On("Save", mock.MatchedBy(func(c *domain.MenuCategory) bool {
			return c.Name == "Starters" && c.RestaurantID == restaurantID && c.DisplayOrder == 1
		})).Return(nil)

		req := menu.CreateMenuCategoryRequest{
			RestaurantID: restaurantID,
			Name:         "Starters",
			DisplayOrder: 1,
		}

		resp, err := useCase.Execute(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Starters", resp.Name)
		assert.Equal(t, restaurantID, resp.RestaurantID)
		repo.AssertExpectations(t)
	})

	t.Run("fails with invalid name", func(t *testing.T) {
		repo := new(MockMenuCategoryRepository)
		useCase := menu.NewCreateMenuCategoryUseCase(repo)

		req := menu.CreateMenuCategoryRequest{
			RestaurantID: "rest-1",
			Name:         "",
			DisplayOrder: 1,
		}

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "category name must be between 1 and 50 characters")
	})
}
