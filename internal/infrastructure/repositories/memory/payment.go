package memory

import payAdapters "bitmerchant/internal/payment/adapters"

type MemoryPaymentRepository = payAdapters.MemoryPaymentRepository

var NewMemoryPaymentRepository = payAdapters.NewMemoryPaymentRepository
