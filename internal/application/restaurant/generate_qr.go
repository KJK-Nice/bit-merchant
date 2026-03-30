package restaurant

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"bitmerchant/internal/domain"
)

type QRCodeService interface {
	GeneratePNG(url string, size int) ([]byte, error)
}

// GenerateRestaurantQRUseCase encodes per-table menu URLs into PNG QR images.
type GenerateRestaurantQRUseCase struct {
	qrService  QRCodeService
	baseURL    string
	restaurant domain.RestaurantRepository
}

// NewGenerateRestaurantQRUseCase builds the use case with restaurant resolution for table bounds.
func NewGenerateRestaurantQRUseCase(qrService QRCodeService, baseURL string, restaurant domain.RestaurantRepository) *GenerateRestaurantQRUseCase {
	return &GenerateRestaurantQRUseCase{
		qrService:  qrService,
		baseURL:    baseURL,
		restaurant: restaurant,
	}
}

// MenuURLForTable returns the absolute customer menu URL for a restaurant table (for tests and diagnostics).
func MenuURLForTable(baseURL string, restaurantID domain.RestaurantID, tableNumber int) string {
	base := strings.TrimRight(baseURL, "/")
	menu, err := url.JoinPath(base, "menu")
	if err != nil {
		return ""
	}
	parsed, err := url.Parse(menu)
	if err != nil {
		return ""
	}
	q := parsed.Query()
	q.Set("restaurantID", string(restaurantID))
	q.Set("table", strconv.Itoa(tableNumber))
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

// Execute generates a PNG QR for the given table if it exists for the restaurant.
func (uc *GenerateRestaurantQRUseCase) Execute(ctx context.Context, restaurantID domain.RestaurantID, tableNumber int) ([]byte, error) {
	if tableNumber < domain.MinTableCount {
		return nil, fmt.Errorf("invalid table number")
	}
	rest, err := uc.restaurant.FindByID(restaurantID)
	if err != nil {
		return nil, err
	}
	tc := rest.TableCount
	if tc < domain.MinTableCount {
		tc = domain.MinTableCount
	}
	if tableNumber > tc {
		return nil, fmt.Errorf("table number out of range")
	}
	menuURL := MenuURLForTable(uc.baseURL, restaurantID, tableNumber)
	if menuURL == "" {
		return nil, fmt.Errorf("invalid base URL for QR")
	}
	return uc.qrService.GeneratePNG(menuURL, 512)
}
