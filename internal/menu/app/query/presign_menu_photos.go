package query

import (
	"context"
	"fmt"

	"bitmerchant/internal/menu/domain/menu"
)

// ItemsWithPresignedPhotos returns a copy of items with PhotoURL replaced by a time-limited GET URL when storage is configured.
func ItemsWithPresignedPhotos(ctx context.Context, items []*menu.MenuItem, photos menu.PhotoStorage, cfg PhotoSignerConfig) ([]*menu.MenuItem, error) {
	if photos == nil {
		return items, nil
	}

	out := make([]*menu.MenuItem, len(items))
	for i, item := range items {
		cp := *item
		if cp.PhotoURL != "" {
			key := PhotoObjectKeyFromStoredValue(cp.PhotoURL, cfg.Bucket, cfg.Endpoint, cfg.PublicBaseURL)
			if key == "" {
				key = cp.PhotoURL
			}
			signed, err := photos.PresignGet(ctx, key)
			if err != nil {
				return nil, fmt.Errorf("presign menu photo: %w", err)
			}
			cp.PhotoURL = signed
		}
		out[i] = &cp
	}
	return out, nil
}
