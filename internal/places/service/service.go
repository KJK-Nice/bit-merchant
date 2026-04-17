package service

import (
	placesCmd "bitmerchant/internal/places/app/command"
	placesQuery "bitmerchant/internal/places/app/query"
	placeshttp "bitmerchant/internal/places/ports/http"
	"bitmerchant/internal/wiring"
)

// Places bundles places application handlers and the HTTP adapter.
type Places struct {
	RecordMenuVisit        placesCmd.RecordMenuVisitHandler
	ListVisitedRestaurants placesQuery.SessionVisitedPlacesHandler
	HTTP                   *placeshttp.PlacesHandler
}

// New wires places bounded-context handlers and HTTP port.
func New(repos wiring.Repositories) Places {
	recordMenuVisitUC := placesCmd.NewRecordMenuVisitHandler(repos.Restaurant, repos.SessionRestaurantVisits, nil, nil)
	listVisitedUC := placesQuery.NewSessionVisitedPlacesHandler(repos.SessionRestaurantVisits, repos.Restaurant, repos.Order, nil, nil)
	return Places{
		RecordMenuVisit:        recordMenuVisitUC,
		ListVisitedRestaurants: listVisitedUC,
		HTTP:                   placeshttp.NewPlacesHandler(listVisitedUC),
	}
}
