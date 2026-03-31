package command

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/menu/domain/menu"
)

type UploadPhotoRequest struct {
	RestaurantID common.RestaurantID
	ItemID       common.ItemID
	File         io.Reader
	Filename     string
	ContentType  string
}

type UploadPhotoUseCase struct {
	itemRepo menu.ItemRepository
	storage  menu.PhotoStorage
}

func NewUploadPhotoUseCase(itemRepo menu.ItemRepository, storage menu.PhotoStorage) *UploadPhotoUseCase {
	return &UploadPhotoUseCase{
		itemRepo: itemRepo,
		storage:  storage,
	}
}

func (uc *UploadPhotoUseCase) Execute(ctx context.Context, req UploadPhotoRequest) (string, error) {
	item, err := uc.itemRepo.FindByID(req.ItemID)
	if err != nil {
		return "", err
	}
	if item.RestaurantID != req.RestaurantID {
		return "", fmt.Errorf("item does not belong to restaurant")
	}

	ext := filepath.Ext(req.Filename)
	key := fmt.Sprintf("restaurants/%s/items/%s_%d%s", req.RestaurantID, req.ItemID, time.Now().Unix(), ext)

	storedKey, err := uc.storage.Upload(ctx, key, req.File, req.ContentType)
	if err != nil {
		return "", err
	}

	item.SetPhotoURLs(storedKey, storedKey)
	if err := uc.itemRepo.Update(item); err != nil {
		_ = uc.storage.Delete(ctx, key)
		return "", err
	}

	return storedKey, nil
}
