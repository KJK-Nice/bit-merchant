package query

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/restaurant/domain/restaurant"
	"log/slog"
)

// QRCodeService generates PNG bytes for a URL.
type QRCodeService interface {
	GeneratePNG(url string, size int) ([]byte, error)
}

// RestaurantTableQRImage resolves a QR PNG for a restaurant table.
type RestaurantTableQRImage struct {
	RestaurantID common.RestaurantID
	TableNumber  int
}

type RestaurantTableQRImageHandler decorator.QueryHandler[RestaurantTableQRImage, []byte]

type restaurantTableQRImageHandler struct {
	qrService  QRCodeService
	baseURL    string
	restaurant restaurant.Repository
}

func NewRestaurantTableQRImageHandler(qrService QRCodeService, baseURL string, repo restaurant.Repository, log *slog.Logger, metrics decorator.MetricsClient) RestaurantTableQRImageHandler {
	if qrService == nil {
		panic("nil QRCodeService")
	}
	if repo == nil {
		panic("nil restaurant.Repository")
	}
	h := restaurantTableQRImageHandler{qrService: qrService, baseURL: baseURL, restaurant: repo}
	return decorator.ApplyQueryDecorators[RestaurantTableQRImage, []byte](h, log, metrics)
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

func (h restaurantTableQRImageHandler) Handle(ctx context.Context, q RestaurantTableQRImage) ([]byte, error) {
	_ = ctx
	if q.TableNumber < restaurant.MinTableCount {
		return nil, fmt.Errorf("invalid table number")
	}
	rest, err := h.restaurant.FindByID(q.RestaurantID)
	if err != nil {
		return nil, err
	}
	tc := rest.TableCount
	if tc < restaurant.MinTableCount {
		tc = restaurant.MinTableCount
	}
	if q.TableNumber > tc {
		return nil, fmt.Errorf("table number out of range")
	}
	menuURL := MenuURLForTable(h.baseURL, q.RestaurantID, q.TableNumber)
	if menuURL == "" {
		return nil, fmt.Errorf("invalid base URL for QR")
	}
	return h.qrService.GeneratePNG(menuURL, 512)
}
