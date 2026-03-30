package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/migrations"
	postgresRepos "bitmerchant/internal/infrastructure/repositories/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

func TestCorePostgresRepositories_BasicRoundTrip(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is required for postgres integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := migrations.EnsureDatabaseExists(ctx, databaseURL)
	require.NoError(t, err)

	db, err := sql.Open("pgx", databaseURL)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, db.PingContext(ctx))
	require.NoError(t, migrations.Up(ctx, db))

	restaurantRepo := postgresRepos.NewRestaurantRepository(db)
	categoryRepo := postgresRepos.NewMenuCategoryRepository(db)
	itemRepo := postgresRepos.NewMenuItemRepository(db)
	orderRepo := postgresRepos.NewOrderRepository(db)
	paymentRepo := postgresRepos.NewPaymentRepository(db)

	restaurantID := domain.RestaurantID("it-postgres-rest-1")
	restaurant, err := domain.NewRestaurant(restaurantID, "Integration Test Place")
	require.NoError(t, err)
	require.NoError(t, restaurantRepo.Save(restaurant))

	category, err := domain.NewMenuCategory("it-cat-1", restaurantID, "Integration Category", 1)
	require.NoError(t, err)
	require.NoError(t, categoryRepo.Save(category))

	item, err := domain.NewMenuItem("it-item-1", category.ID, restaurantID, "Integration Dish", 99.99)
	require.NoError(t, err)
	require.NoError(t, itemRepo.Save(item))

	orderItem, err := domain.NewOrderItem("it-order-item-1", "it-order-1", item.ID, item.Name, 1, item.Price)
	require.NoError(t, err)
	order, err := domain.NewOrder("it-order-1", "IT-0001", restaurantID, "it-session-1", []domain.OrderItem{*orderItem}, 100, domain.PaymentMethodTypeCash)
	require.NoError(t, err)
	require.NoError(t, orderRepo.Save(order))

	payment, err := domain.NewPayment("it-pay-1", order.ID, restaurantID, domain.PaymentMethodTypeCash, 100)
	require.NoError(t, err)
	require.NoError(t, paymentRepo.Save(payment))

	foundOrder, err := orderRepo.FindByOrderNumber(restaurantID, "IT-0001")
	require.NoError(t, err)
	require.Equal(t, order.ID, foundOrder.ID)
	require.Len(t, foundOrder.Items, 1)

	availableItems, err := itemRepo.FindAvailableByRestaurantID(restaurantID)
	require.NoError(t, err)
	require.NotEmpty(t, availableItems)

	foundPayment, err := paymentRepo.FindByOrderID(order.ID)
	require.NoError(t, err)
	require.Equal(t, payment.ID, foundPayment.ID)
}
