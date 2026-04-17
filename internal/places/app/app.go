package app

import (
	"bitmerchant/internal/places/app/command"
	"bitmerchant/internal/places/app/query"
)

// Application bundles places bounded-context use cases.
type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	RecordMenuVisit command.RecordMenuVisitHandler
}

type Queries struct {
	SessionVisitedPlaces query.SessionVisitedPlacesHandler
}
