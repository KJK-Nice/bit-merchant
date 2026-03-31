package postgres

import payAdapters "bitmerchant/internal/payment/adapters"

type PaymentRepository = payAdapters.PostgresPaymentRepository

var NewPaymentRepository = payAdapters.NewPostgresPaymentRepository
