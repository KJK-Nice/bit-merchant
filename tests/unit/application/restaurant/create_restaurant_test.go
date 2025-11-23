package restaurant_test

import (
	"context"
	"errors"
	"testing"

	"bitmerchant/internal/application/restaurant"
	"bitmerchant/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRestaurantRepository struct {
	mock.Mock
}

func (m *MockRestaurantRepository) Save(r *domain.Restaurant) error {
	args := m.Called(r)
	return args.Error(0)
}

func (m *MockRestaurantRepository) FindByID(id domain.RestaurantID) (*domain.Restaurant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) Update(r *domain.Restaurant) error {
	args := m.Called(r)
	return args.Error(0)
}

func TestCreateRestaurantUseCase_Execute(t *testing.T) {
	t.Run("successfully creates restaurant", func(t *testing.T) {
		repo := new(MockRestaurantRepository)
		useCase := restaurant.NewCreateRestaurantUseCase(repo)

		repo.On("Save", mock.MatchedBy(func(r *domain.Restaurant) bool {
			return r.Name == "My Tasty Place" && r.ID != ""
		})).Return(nil)

		req := restaurant.CreateRestaurantRequest{
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
		useCase := restaurant.NewCreateRestaurantUseCase(repo)

		req := restaurant.CreateRestaurantRequest{
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
		useCase := restaurant.NewCreateRestaurantUseCase(repo)

		repo.On("Save", mock.Anything).Return(errors.New("db error"))

		req := restaurant.CreateRestaurantRequest{
			Name: "My Tasty Place",
		}

		resp, err := useCase.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "db error", err.Error())
	})
}
