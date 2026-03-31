package restaurant

import restQuery "bitmerchant/internal/restaurant/app/query"

type QRCodeService = restQuery.QRCodeService
type GenerateRestaurantQRUseCase = restQuery.GenerateRestaurantQRUseCase

var NewGenerateRestaurantQRUseCase = restQuery.NewGenerateRestaurantQRUseCase
var MenuURLForTable = restQuery.MenuURLForTable
