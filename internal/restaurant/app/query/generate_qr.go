package query

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"bitmerchant/internal/common"
	"bitmerchant/internal/restaurant/domain/restaurant"
)

type QRCodeService interface {
	GeneratePNG(url string, size int) ([]byte, error)
}

type GenerateRestaurantQRUseCase struct {
	qrService  QRCodeService
	baseURL    string
	restaurant restaurant.Repository
}

func NewGenerateRestaurantQRUseCase(qrService QRCodeService, baseURL string, repo restaurant.Repository) *GenerateRestaurantQRUseCase {
	return &GenerateRestaurantQRUseCase{
		qrService:  qrService,
		baseURL:    baseURL,
		restaurant: repo,
	}
}

func MenuURLForTable(baseURL string, restaurantID common.RestaurantID, tableNumber int) string {
	base := strings.TrimRight(baseURL, "/")
	menuPath, err := url.JoinPath(base, "menu")
	if err != nil {
		return ""
	}
	parsed, err := url.Parse(menuPath)
	if err != nil {
		return ""
	}
	q := parsed.Query()
	q.Set("restaurantID", string(restaurantID))
	q.Set("table", strconv.Itoa(tableNumber))
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

func (uc *GenerateRestaurantQRUseCase) Execute(ctx context.Context, restaurantID common.RestaurantID, tableNumber int) ([]byte, error) {
	if tableNumber < restaurant.MinTableCount {
		return nil, fmt.Errorf("invalid table number")
	}
	rest, err := uc.restaurant.FindByID(restaurantID)
	if err != nil {
		return nil, err
	}
	tc := rest.TableCount
	if tc < restaurant.MinTableCount {
		tc = restaurant.MinTableCount
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
