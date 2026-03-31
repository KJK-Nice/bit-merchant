package memory

import orderAdapters "bitmerchant/internal/ordering/adapters"

type MemoryOrderRepository = orderAdapters.MemoryOrderRepository

var NewMemoryOrderRepository = orderAdapters.NewMemoryOrderRepository
