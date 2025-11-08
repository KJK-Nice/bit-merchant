package cart

// GetCartUseCase retrieves current cart contents
type GetCartUseCase struct {
	cartStore *CartStore
}

// NewGetCartUseCase creates a new GetCartUseCase
func NewGetCartUseCase(cartStore *CartStore) *GetCartUseCase {
	return &GetCartUseCase{
		cartStore: cartStore,
	}
}

// Execute retrieves cart for session
func (uc *GetCartUseCase) Execute(sessionID string) *Cart {
	return uc.cartStore.GetCart(sessionID)
}
