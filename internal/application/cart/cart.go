package cart

import orderCart "bitmerchant/internal/ordering/app/cart"

type CartItem = orderCart.CartItem
type Cart = orderCart.Cart
type CartService = orderCart.CartService

var NewCartService = orderCart.NewCartService
