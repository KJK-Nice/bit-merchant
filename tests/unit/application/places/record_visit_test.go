package places_test

import (
	"context"
	"testing"
	"time"

	"bitmerchant/internal/application/places"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/repositories/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordMenuVisitUseCase(t *testing.T) {
	rest := memory.NewMemoryRestaurantRepository()
	visits := memory.NewMemorySessionRestaurantVisitRepository()
	r, _ := domain.NewRestaurant("r1", "Cafe")
	require.NoError(t, rest.Save(r))

	uc := places.NewRecordMenuVisitUseCase(rest, visits)
	ctx := context.Background()

	require.NoError(t, uc.Execute(ctx, "sess-a", "r1"))
	require.NoError(t, uc.Execute(ctx, "sess-a", "r1"))

	got, err := visits.FindBySessionID("sess-a")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, domain.RestaurantID("r1"), got[0].RestaurantID)
	assert.True(t, got[0].LastVisitedAt.After(got[0].FirstVisitedAt) || got[0].LastVisitedAt.Equal(got[0].FirstVisitedAt))
	assert.WithinDuration(t, time.Now(), got[0].LastVisitedAt, 2*time.Second)
}
