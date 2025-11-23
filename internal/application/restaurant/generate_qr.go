package restaurant

import (
	"context"
	"fmt"
)

type QRCodeService interface {
	GeneratePNG(url string, size int) ([]byte, error)
}

type GenerateRestaurantQRUseCase struct {
	qrService QRCodeService
	baseURL   string
}

func NewGenerateRestaurantQRUseCase(qrService QRCodeService, baseURL string) *GenerateRestaurantQRUseCase {
	return &GenerateRestaurantQRUseCase{
		qrService: qrService,
		baseURL:   baseURL,
	}
}

func (uc *GenerateRestaurantQRUseCase) Execute(ctx context.Context, restaurantID string) ([]byte, error) {
	// URL to the menu: e.g. https://bitmerchant.com/menu?restaurantID=... (or just /menu for MVP with single rest)
	url := fmt.Sprintf("%s/menu", uc.baseURL)
	return uc.qrService.GeneratePNG(url, 512)
}

