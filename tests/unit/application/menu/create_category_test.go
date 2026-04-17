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

type MockMenuCategoryRepository struct {
	mock.Mock
}

func (m *MockMenuCategoryRepository) Save(c *menu.MenuCategory) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *MockMenuCategoryRepository) FindByRestaurantID(id common.RestaurantID) ([]*menu.MenuCategory, error) {
	args := m.Called(id)
	return args.Get(0).([]*menu.MenuCategory), args.Error(1)
}

func (m *MockMenuCategoryRepository) FindByID(id common.CategoryID) (*menu.MenuCategory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*menu.MenuCategory), args.Error(1)
}

func (m *MockMenuCategoryRepository) Update(c *menu.MenuCategory) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *MockMenuCategoryRepository) Delete(id common.CategoryID) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestCreateMenuCategoryHandler_Handle(t *testing.T) {
	t.Run("successfully creates category", func(t *testing.T) {
		repo := new(MockMenuCategoryRepository)
		useCase := menuCmd.NewCreateMenuCategoryHandler(repo, nil, nil)

		restaurantID := common.RestaurantID("rest-1")

		repo.On("Save", mock.MatchedBy(func(c *menu.MenuCategory) bool {
			return c.Name == "Starters" && c.RestaurantID == restaurantID && c.DisplayOrder == 1
		})).Return(nil)

		req := menuCmd.CreateMenuCategory{
			RestaurantID: restaurantID,
			Name:         "Starters",
			DisplayOrder: 1,
		}

		resp, err := useCase.Handle(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Starters", resp.Name)
		assert.Equal(t, restaurantID, resp.RestaurantID)
		repo.AssertExpectations(t)
	})

	t.Run("fails with invalid name", func(t *testing.T) {
		repo := new(MockMenuCategoryRepository)
		useCase := menuCmd.NewCreateMenuCategoryHandler(repo, nil, nil)

		req := menuCmd.CreateMenuCategory{
			RestaurantID: "rest-1",
			Name:         "",
			DisplayOrder: 1,
		}

		resp, err := useCase.Handle(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "category name must be between 1 and 50 characters")
	})
}
