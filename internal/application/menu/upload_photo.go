package menu

import (
	"bitmerchant/internal/domain"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"
)

type UploadPhotoRequest struct {
	RestaurantID domain.RestaurantID
	ItemID       domain.ItemID
	File         io.Reader
	Filename     string
	ContentType  string
}

type UploadPhotoUseCase struct {
	itemRepo domain.MenuItemRepository
	storage  domain.PhotoStorage
}

func NewUploadPhotoUseCase(itemRepo domain.MenuItemRepository, storage domain.PhotoStorage) *UploadPhotoUseCase {
	return &UploadPhotoUseCase{
		itemRepo: itemRepo,
		storage:  storage,
	}
}

func (uc *UploadPhotoUseCase) Execute(ctx context.Context, req UploadPhotoRequest) (string, error) {
	// 1. Validate Item ownership
	item, err := uc.itemRepo.FindByID(req.ItemID)
	if err != nil {
		return "", err
	}
	if item.RestaurantID != req.RestaurantID {
		return "", fmt.Errorf("item does not belong to restaurant")
	}

	// 2. Generate Key
	ext := filepath.Ext(req.Filename)
	key := fmt.Sprintf("restaurants/%s/items/%s_%d%s", req.RestaurantID, req.ItemID, time.Now().Unix(), ext)

	// 3. Upload to Storage
	url, err := uc.storage.Upload(ctx, key, req.File, req.ContentType)
	if err != nil {
		return "", err
	}

	// 4. Update Item
	// Assuming we store the URL. Original vs Thumbnail logic can be added later.
	item.SetPhotoURLs(url, url)
	if err := uc.itemRepo.Update(item); err != nil {
		// Cleanup upload if db save fails?
		_ = uc.storage.Delete(ctx, key)
		return "", err
	}

	return url, nil
}

