package restaurant_test

import (
	"bitmerchant/internal/common"
	restaurantCmd "bitmerchant/internal/restaurant/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"context"
	"errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockRestaurantRepository struct {
	mock.Mock
}

func (m *MockRestaurantRepository) Save(r *restaurant.Restaurant) error {
	args := m.Called(r)
	return args.Error(0)
}

func (m *MockRestaurantRepository) FindByID(id common.RestaurantID) (*restaurant.Restaurant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*restaurant.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) Update(r *restaurant.Restaurant) error {
	args := m.Called(r)
	return args.Error(0)
}

func TestCreateRestaurantUseCase_Execute(t *testing.T) {
	t.Run("successfully creates restaurant", func(t *testing.T) {
		repo := new(MockRestaurantRepository)
		useCase := restaurantCmd.NewCreateRestaurantUseCase(repo)

		repo.On("Save", mock.MatchedBy(func(r *restaurant.Restaurant) bool {
			return r.Name == "My Tasty Place" && r.ID != "" && r.TableCount == restaurant.MinTableCount
		})).Return(nil)

		req := restaurantCmd.CreateRestaurantRequest{
			Name: "My Tasty Place",
		}

		resp, err := useCase.Execute(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "My Tasty Place", resp.Name)
		assert.NotEmpty(t, resp.ID)
		repo.AssertExpectations(t)
	})

	t.Run("fails with empty name", func(t *testing.T) {
		repo := new(MockRestaurantRepository)
		useCase := restaurantCmd.NewCreateRestaurantUseCase(repo)

		req := restaurantCmd.CreateRestaurantRequest{
			Name: "",
		}

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "restaurant name must be between 1 and 100 characters")
		repo.AssertNotCalled(t, "Save")
	})

	t.Run("fails when repository error", func(t *testing.T) {
		repo := new(MockRestaurantRepository)
		useCase := restaurantCmd.NewCreateRestaurantUseCase(repo)

		repo.On("Save", mock.Anything).Return(errors.New("db error"))

		req := restaurantCmd.CreateRestaurantRequest{
			Name: "My Tasty Place",
		}

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "db error", err.Error())
	})
}
