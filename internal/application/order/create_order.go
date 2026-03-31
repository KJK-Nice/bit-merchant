package order

import orderCmd "bitmerchant/internal/ordering/app/command"

type CreateOrderRequest = orderCmd.CreateOrderRequest
type CreateOrderResponse = orderCmd.CreateOrderResponse
type CreateOrderUseCase = orderCmd.CreateOrderUseCase

var NewCreateOrderUseCase = orderCmd.NewCreateOrderUseCase
