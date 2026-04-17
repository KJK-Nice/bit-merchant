package places_test

import (
	"context"
	"testing"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/infrastructure/repositories/memory"
	placesCmd "bitmerchant/internal/places/app/command"
	"bitmerchant/internal/restaurant/domain/restaurant"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordMenuVisitHandler(t *testing.T) {
	rest := memory.NewMemoryRestaurantRepository()
	visits := memory.NewMemorySessionRestaurantVisitRepository()
	r, _ := restaurant.NewRestaurant("r1", "Cafe")
	require.NoError(t, rest.Save(r))

	h := placesCmd.NewRecordMenuVisitHandler(rest, visits, nil, nil)
	ctx := context.Background()

	require.NoError(t, h.Handle(ctx, placesCmd.RecordMenuVisit{SessionID: "sess-a", RestaurantID: "r1"}))
	require.NoError(t, h.Handle(ctx, placesCmd.RecordMenuVisit{SessionID: "sess-a", RestaurantID: "r1"}))

	got, err := visits.FindBySessionID(ctx, "sess-a")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, common.RestaurantID("r1"), got[0].RestaurantID())
	assert.True(t, got[0].LastVisitedAt().After(got[0].FirstVisitedAt()) || got[0].LastVisitedAt().Equal(got[0].FirstVisitedAt()))
	assert.WithinDuration(t, time.Now(), got[0].LastVisitedAt(), 2*time.Second)
}
