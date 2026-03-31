package postgres

import orderAdapters "bitmerchant/internal/ordering/adapters"

type OrderRepository = orderAdapters.PostgresOrderRepository

var NewOrderRepository = orderAdapters.NewPostgresOrderRepository
