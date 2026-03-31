package postgres_test

import (
	"testing"
	"time"

	authAdapters "bitmerchant/internal/auth/adapters"
	"bitmerchant/internal/auth/domain/invitation"
	"bitmerchant/internal/auth/domain/membership"
	"bitmerchant/internal/auth/domain/session"
	"bitmerchant/internal/auth/domain/user"
	"bitmerchant/internal/common"
	menuAdapters "bitmerchant/internal/menu/adapters"
	"bitmerchant/internal/menu/domain/menu"
	orderAdapters "bitmerchant/internal/ordering/adapters"
	"bitmerchant/internal/ordering/domain/order"
	payAdapters "bitmerchant/internal/payment/adapters"
	"bitmerchant/internal/payment/domain/payment"
	placesAdapters "bitmerchant/internal/places/adapters"
	"bitmerchant/internal/places/domain/visit"
	restAdapters "bitmerchant/internal/restaurant/adapters"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestaurantRepository(t *testing.T) {
	db := setupPostgresContainer(t)
	repo := restAdapters.NewPostgresRestaurantRepository(db)

	id := common.RestaurantID("rest-pg-1")
	r, err := restaurant.NewRestaurant(id, "PG Test Cafe")
	require.NoError(t, err)

	t.Run("Save and FindByID", func(t *testing.T) {
		require.NoError(t, repo.Save(r))

		found, err := repo.FindByID(id)
		require.NoError(t, err)
		assert.Equal(t, id, found.ID)
		assert.Equal(t, "PG Test Cafe", found.Name)
		assert.True(t, found.IsOpen)
	})

	t.Run("Update", func(t *testing.T) {
		r.Close("On vacation", "Next Monday")
		require.NoError(t, repo.Update(r))

		found, err := repo.FindByID(id)
		require.NoError(t, err)
		assert.False(t, found.IsOpen)
		assert.Equal(t, "On vacation", found.ClosedMessage)
		assert.Equal(t, "Next Monday", found.ReopeningHours)
	})

	t.Run("FindByID not found", func(t *testing.T) {
		_, err := repo.FindByID("nonexistent")
		assert.Error(t, err)
	})
}

func TestMenuCategoryRepository(t *testing.T) {
	db := setupPostgresContainer(t)
	restRepo := restAdapters.NewPostgresRestaurantRepository(db)
	repo := menuAdapters.NewPostgresCategoryRepository(db)

	restID := common.RestaurantID("rest-cat-1")
	r, _ := restaurant.NewRestaurant(restID, "Cat Test")
	require.NoError(t, restRepo.Save(r))

	cat, err := menu.NewMenuCategory("cat-1", restID, "Appetizers", 1)
	require.NoError(t, err)

	t.Run("Save and FindByID", func(t *testing.T) {
		require.NoError(t, repo.Save(cat))

		found, err := repo.FindByID("cat-1")
		require.NoError(t, err)
		assert.Equal(t, "Appetizers", found.Name)
		assert.Equal(t, 1, found.DisplayOrder)
		assert.True(t, found.IsActive)
	})

	t.Run("FindByRestaurantID", func(t *testing.T) {
		cats, err := repo.FindByRestaurantID(restID)
		require.NoError(t, err)
		assert.Len(t, cats, 1)
	})

	t.Run("Update", func(t *testing.T) {
		cat.SetActive(false)
		require.NoError(t, repo.Update(cat))

		found, err := repo.FindByID("cat-1")
		require.NoError(t, err)
		assert.False(t, found.IsActive)
	})

	t.Run("Delete", func(t *testing.T) {
		require.NoError(t, repo.Delete("cat-1"))

		_, err := repo.FindByID("cat-1")
		assert.Error(t, err)
	})
}

func TestMenuItemRepository(t *testing.T) {
	db := setupPostgresContainer(t)
	restRepo := restAdapters.NewPostgresRestaurantRepository(db)
	catRepo := menuAdapters.NewPostgresCategoryRepository(db)
	repo := menuAdapters.NewPostgresItemRepository(db)

	restID := common.RestaurantID("rest-item-1")
	r, _ := restaurant.NewRestaurant(restID, "Item Test")
	require.NoError(t, restRepo.Save(r))

	cat, _ := menu.NewMenuCategory("cat-item-1", restID, "Mains", 1)
	require.NoError(t, catRepo.Save(cat))

	item, err := menu.NewMenuItem("item-1", "cat-item-1", restID, "Burger", 12.50)
	require.NoError(t, err)

	t.Run("Save and FindByID", func(t *testing.T) {
		require.NoError(t, repo.Save(item))

		found, err := repo.FindByID("item-1")
		require.NoError(t, err)
		assert.Equal(t, "Burger", found.Name)
		assert.Equal(t, 12.50, found.Price)
		assert.True(t, found.IsAvailable)
	})

	t.Run("FindAvailableByRestaurantID", func(t *testing.T) {
		items, err := repo.FindAvailableByRestaurantID(restID)
		require.NoError(t, err)
		assert.Len(t, items, 1)
	})

	t.Run("Update", func(t *testing.T) {
		item.SetAvailable(false)
		require.NoError(t, repo.Update(item))

		items, err := repo.FindAvailableByRestaurantID(restID)
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("Delete", func(t *testing.T) {
		require.NoError(t, repo.Delete("item-1"))

		_, err := repo.FindByID("item-1")
		assert.Error(t, err)
	})
}

func TestOrderRepository(t *testing.T) {
	db := setupPostgresContainer(t)
	restRepo := restAdapters.NewPostgresRestaurantRepository(db)
	catRepo := menuAdapters.NewPostgresCategoryRepository(db)
	itemRepo := menuAdapters.NewPostgresItemRepository(db)
	repo := orderAdapters.NewPostgresOrderRepository(db)

	restID := common.RestaurantID("rest-ord-1")
	r, _ := restaurant.NewRestaurant(restID, "Order Test")
	require.NoError(t, restRepo.Save(r))

	cat, _ := menu.NewMenuCategory("cat-ord-1", restID, "Food", 1)
	require.NoError(t, catRepo.Save(cat))

	menuItem, _ := menu.NewMenuItem("mi-ord-1", "cat-ord-1", restID, "Pizza", 15.00)
	require.NoError(t, itemRepo.Save(menuItem))

	orderItem, err := order.NewOrderItem("oi-1", "ord-1", "mi-ord-1", "Pizza", 2, 15.00)
	require.NoError(t, err)

	o, err := order.NewOrder("ord-1", "0001", restID, "sess-1",
		[]order.OrderItem{*orderItem}, 3000, common.PaymentMethodTypeCash)
	require.NoError(t, err)
	o.FiatAmount = 30.00

	t.Run("Save and FindByID with items", func(t *testing.T) {
		require.NoError(t, repo.Save(o))

		found, err := repo.FindByID("ord-1")
		require.NoError(t, err)
		assert.Equal(t, common.OrderID("ord-1"), found.ID)
		assert.Equal(t, common.OrderNumber("0001"), found.OrderNumber)
		require.Len(t, found.Items, 1)
		assert.Equal(t, "Pizza", found.Items[0].Name)
		assert.Equal(t, 2, found.Items[0].Quantity)
		assert.Equal(t, 30.0, found.Items[0].Subtotal)
	})

	t.Run("FindByOrderNumber", func(t *testing.T) {
		found, err := repo.FindByOrderNumber(restID, "0001")
		require.NoError(t, err)
		assert.Equal(t, common.OrderID("ord-1"), found.ID)
		require.Len(t, found.Items, 1)
	})

	t.Run("FindBySessionID", func(t *testing.T) {
		orders, err := repo.FindBySessionID("sess-1")
		require.NoError(t, err)
		assert.Len(t, orders, 1)
		assert.Len(t, orders[0].Items, 1)
	})

	t.Run("FindActiveByRestaurantID", func(t *testing.T) {
		orders, err := repo.FindActiveByRestaurantID(restID)
		require.NoError(t, err)
		assert.Len(t, orders, 1)
	})

	t.Run("Update status", func(t *testing.T) {
		o.MarkPaid()
		require.NoError(t, o.StartPreparing())
		require.NoError(t, repo.Update(o))

		found, err := repo.FindByID("ord-1")
		require.NoError(t, err)
		assert.Equal(t, common.PaymentStatusPaid, found.PaymentStatus)
		assert.Equal(t, common.FulfillmentStatusPreparing, found.FulfillmentStatus)
		assert.NotNil(t, found.PaidAt)
		assert.NotNil(t, found.PreparingAt)
	})

	t.Run("FindByID not found", func(t *testing.T) {
		_, err := repo.FindByID("nonexistent")
		assert.Error(t, err)
	})
}

func TestPaymentRepository(t *testing.T) {
	db := setupPostgresContainer(t)
	restRepo := restAdapters.NewPostgresRestaurantRepository(db)
	catRepo := menuAdapters.NewPostgresCategoryRepository(db)
	itemRepo := menuAdapters.NewPostgresItemRepository(db)
	orderRepo := orderAdapters.NewPostgresOrderRepository(db)
	repo := payAdapters.NewPostgresPaymentRepository(db)

	restID := common.RestaurantID("rest-pay-1")
	r, _ := restaurant.NewRestaurant(restID, "Pay Test")
	require.NoError(t, restRepo.Save(r))

	cat, _ := menu.NewMenuCategory("cat-pay-1", restID, "Food", 1)
	require.NoError(t, catRepo.Save(cat))

	menuItem, _ := menu.NewMenuItem("mi-pay-1", "cat-pay-1", restID, "Salad", 10.00)
	require.NoError(t, itemRepo.Save(menuItem))

	oi, _ := order.NewOrderItem("oi-pay-1", "ord-pay-1", "mi-pay-1", "Salad", 1, 10.00)
	o, _ := order.NewOrder("ord-pay-1", "P001", restID, "sess-pay-1",
		[]order.OrderItem{*oi}, 1000, common.PaymentMethodTypeCash)
	require.NoError(t, orderRepo.Save(o))

	pay, err := payment.NewPayment("pay-1", "ord-pay-1", restID, common.PaymentMethodTypeCash, 10.00)
	require.NoError(t, err)

	t.Run("Save and FindByID", func(t *testing.T) {
		require.NoError(t, repo.Save(pay))

		found, err := repo.FindByID("pay-1")
		require.NoError(t, err)
		assert.Equal(t, common.PaymentID("pay-1"), found.ID)
		assert.Equal(t, 10.00, found.Amount)
		assert.Equal(t, common.PaymentStatusPending, found.Status)
	})

	t.Run("FindByOrderID", func(t *testing.T) {
		found, err := repo.FindByOrderID("ord-pay-1")
		require.NoError(t, err)
		assert.Equal(t, common.PaymentID("pay-1"), found.ID)
	})

	t.Run("Update", func(t *testing.T) {
		pay.MarkAsPaid()
		require.NoError(t, repo.Update(pay))

		found, err := repo.FindByID("pay-1")
		require.NoError(t, err)
		assert.Equal(t, common.PaymentStatusPaid, found.Status)
		assert.NotNil(t, found.PaidAt)
	})
}

func TestUserRepository(t *testing.T) {
	db := setupPostgresContainer(t)
	repo := authAdapters.NewPostgresUserRepository(db)

	u, err := user.NewUser("user-1", "Alice")
	require.NoError(t, err)

	t.Run("Save and FindByID", func(t *testing.T) {
		require.NoError(t, repo.Save(u))

		found, err := repo.FindByID("user-1")
		require.NoError(t, err)
		assert.Equal(t, common.UserID("user-1"), found.ID)
		assert.Equal(t, "Alice", found.DisplayName)
		assert.Empty(t, found.Credentials)
	})

	t.Run("Update with credential", func(t *testing.T) {
		u.AddCredential(webauthn.Credential{ID: []byte("cred-1")})
		require.NoError(t, repo.Update(u))

		found, err := repo.FindByID("user-1")
		require.NoError(t, err)
		require.Len(t, found.Credentials, 1)
	})

	t.Run("FindByCredentialID", func(t *testing.T) {
		found, cred, err := repo.FindByCredentialID([]byte("cred-1"))
		require.NoError(t, err)
		assert.Equal(t, common.UserID("user-1"), found.ID)
		assert.Equal(t, []byte("cred-1"), cred.ID)
	})

	t.Run("FindByCredentialID not found", func(t *testing.T) {
		_, _, err := repo.FindByCredentialID([]byte("nonexistent"))
		assert.Error(t, err)
	})

	t.Run("FindByID not found", func(t *testing.T) {
		_, err := repo.FindByID("nonexistent")
		assert.Error(t, err)
	})
}

func TestSessionRepository(t *testing.T) {
	db := setupPostgresContainer(t)
	repo := authAdapters.NewPostgresSessionRepository(db)

	uid := common.UserID("user-sess-1")
	rid := common.RestaurantID("rest-sess-1")
	now := time.Now().Truncate(time.Microsecond)

	s := &session.Session{
		ID:           "sess-1",
		UserID:       &uid,
		RestaurantID: &rid,
		CreatedAt:    now,
		ExpiresAt:    now.Add(24 * time.Hour),
	}

	t.Run("Save and Get", func(t *testing.T) {
		require.NoError(t, repo.Save(s))

		found, err := repo.Get("sess-1")
		require.NoError(t, err)
		assert.Equal(t, "sess-1", found.ID)
		assert.Equal(t, &uid, found.UserID)
		assert.Equal(t, &rid, found.RestaurantID)
	})

	t.Run("Delete", func(t *testing.T) {
		require.NoError(t, repo.Delete("sess-1"))

		_, err := repo.Get("sess-1")
		assert.Error(t, err)
	})

	t.Run("DeleteByUserID", func(t *testing.T) {
		require.NoError(t, repo.Save(s))

		require.NoError(t, repo.DeleteByUserID(uid))

		_, err := repo.Get("sess-1")
		assert.Error(t, err)
	})

	t.Run("Get not found", func(t *testing.T) {
		_, err := repo.Get("nonexistent")
		assert.Error(t, err)
	})
}

func TestMembershipRepository(t *testing.T) {
	db := setupPostgresContainer(t)
	repo := authAdapters.NewPostgresMembershipRepository(db)

	m, err := membership.NewMembership("mem-1", "user-mem-1", "rest-mem-1", common.RoleOwner)
	require.NoError(t, err)

	t.Run("Save and FindByUserID", func(t *testing.T) {
		require.NoError(t, repo.Save(m))

		found, err := repo.FindByUserID("user-mem-1")
		require.NoError(t, err)
		require.Len(t, found, 1)
		assert.Equal(t, common.MembershipID("mem-1"), found[0].ID)
		assert.Equal(t, common.RoleOwner, found[0].Role)
	})

	t.Run("FindByRestaurantID", func(t *testing.T) {
		found, err := repo.FindByRestaurantID("rest-mem-1")
		require.NoError(t, err)
		require.Len(t, found, 1)
	})

	t.Run("FindByUserAndRestaurant", func(t *testing.T) {
		found, err := repo.FindByUserAndRestaurant("user-mem-1", "rest-mem-1")
		require.NoError(t, err)
		assert.Equal(t, common.MembershipID("mem-1"), found.ID)
	})

	t.Run("FindByUserAndRestaurant not found", func(t *testing.T) {
		_, err := repo.FindByUserAndRestaurant("user-mem-1", "nonexistent")
		assert.Error(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		require.NoError(t, repo.Delete("mem-1"))

		found, err := repo.FindByUserID("user-mem-1")
		require.NoError(t, err)
		assert.Empty(t, found)
	})
}

func TestInvitationRepository(t *testing.T) {
	db := setupPostgresContainer(t)
	repo := authAdapters.NewPostgresInvitationRepository(db)

	inv, err := invitation.NewInvitation("inv-1", "rest-inv-1", common.RoleKitchenStaff, "token-abc", time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	t.Run("Save and FindByToken", func(t *testing.T) {
		require.NoError(t, repo.Save(inv))

		found, err := repo.FindByToken("token-abc")
		require.NoError(t, err)
		assert.Equal(t, common.InvitationID("inv-1"), found.ID)
		assert.Equal(t, common.RoleKitchenStaff, found.Role)
		assert.False(t, found.IsUsed())
	})

	t.Run("FindByRestaurantID", func(t *testing.T) {
		found, err := repo.FindByRestaurantID("rest-inv-1")
		require.NoError(t, err)
		require.Len(t, found, 1)
	})

	t.Run("Update with MarkUsed", func(t *testing.T) {
		uid := common.UserID("user-inv-1")
		inv.MarkUsed(uid, time.Now())
		require.NoError(t, repo.Update(inv))

		found, err := repo.FindByToken("token-abc")
		require.NoError(t, err)
		assert.True(t, found.IsUsed())
		assert.NotNil(t, found.UsedByUserID)
		assert.Equal(t, &uid, found.UsedByUserID)
	})

	t.Run("FindByToken not found", func(t *testing.T) {
		_, err := repo.FindByToken("nonexistent")
		assert.Error(t, err)
	})
}

func TestVisitRepository(t *testing.T) {
	db := setupPostgresContainer(t)
	restRepo := restAdapters.NewPostgresRestaurantRepository(db)
	repo := placesAdapters.NewPostgresVisitRepository(db)

	restID := common.RestaurantID("rest-visit-1")
	r, _ := restaurant.NewRestaurant(restID, "Visit Test")
	require.NoError(t, restRepo.Save(r))

	now := time.Now().Truncate(time.Microsecond)

	t.Run("Upsert and FindBySessionID", func(t *testing.T) {
		v := &visit.SessionRestaurantVisit{
			SessionID:      "sess-visit-1",
			RestaurantID:   restID,
			FirstVisitedAt: now,
			LastVisitedAt:  now,
		}
		require.NoError(t, repo.Upsert(v))

		found, err := repo.FindBySessionID("sess-visit-1")
		require.NoError(t, err)
		require.Len(t, found, 1)
		assert.Equal(t, restID, found[0].RestaurantID)
	})

	t.Run("Upsert updates last_visited_at", func(t *testing.T) {
		later := now.Add(time.Hour)
		v := &visit.SessionRestaurantVisit{
			SessionID:      "sess-visit-1",
			RestaurantID:   restID,
			FirstVisitedAt: now,
			LastVisitedAt:  later,
		}
		require.NoError(t, repo.Upsert(v))

		found, err := repo.FindBySessionID("sess-visit-1")
		require.NoError(t, err)
		require.Len(t, found, 1)
		assert.True(t, found[0].LastVisitedAt.After(now) || found[0].LastVisitedAt.Equal(later))
	})

	t.Run("FindBySessionID empty", func(t *testing.T) {
		found, err := repo.FindBySessionID("nonexistent")
		require.NoError(t, err)
		assert.Empty(t, found)
	})
}
