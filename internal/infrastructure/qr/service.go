package qr

import (
	"fmt"
	
	"github.com/skip2/go-qrcode"
)

// QRCodeService generates QR codes
type QRCodeService struct {}

func NewQRCodeService() *QRCodeService {
	return &QRCodeService{}
}

// GeneratePNG returns a PNG byte slice of the QR code for the given URL
func (s *QRCodeService) GeneratePNG(url string, size int) ([]byte, error) {
	if size <= 0 {
		size = 256
	}
	
	png, err := qrcode.Encode(url, qrcode.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	
	return png, nil
}

