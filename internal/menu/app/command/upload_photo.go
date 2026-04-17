package command

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"bitmerchant/internal/common"
	"bitmerchant/internal/common/decorator"
	"bitmerchant/internal/menu/domain/menu"
	"log/slog"
)

// UploadMenuItemPhoto stores a new photo for a menu item.
type UploadMenuItemPhoto struct {
	RestaurantID common.RestaurantID
	ItemID       common.ItemID
	File         io.Reader
	Filename     string
	ContentType  string
}

type UploadMenuItemPhotoHandler decorator.CommandResultHandler[UploadMenuItemPhoto, string]

type uploadMenuItemPhotoHandler struct {
	itemRepo menu.ItemRepository
	storage  menu.PhotoStorage
}

func NewUploadMenuItemPhotoHandler(itemRepo menu.ItemRepository, storage menu.PhotoStorage, log *slog.Logger, metrics decorator.MetricsClient) UploadMenuItemPhotoHandler {
	if itemRepo == nil {
		panic("nil menu.ItemRepository")
	}
	h := uploadMenuItemPhotoHandler{
		itemRepo: itemRepo,
		storage:  storage,
	}
	return decorator.ApplyCommandResultDecorators[UploadMenuItemPhoto, string](h, log, metrics)
}

func (h uploadMenuItemPhotoHandler) Handle(ctx context.Context, cmd UploadMenuItemPhoto) (string, error) {
	item, err := h.itemRepo.FindByID(cmd.ItemID)
	if err != nil {
		return "", err
	}
	if item.RestaurantID != cmd.RestaurantID {
		return "", fmt.Errorf("item does not belong to restaurant")
	}

	ext := filepath.Ext(cmd.Filename)
	key := fmt.Sprintf("restaurants/%s/items/%s_%d%s", cmd.RestaurantID, cmd.ItemID, time.Now().Unix(), ext)

	storedKey, err := h.storage.Upload(ctx, key, cmd.File, cmd.ContentType)
	if err != nil {
		return "", err
	}

	item.SetPhotoURLs(storedKey, storedKey)
	if err := h.itemRepo.Update(item); err != nil {
		_ = h.storage.Delete(ctx, key)
		return "", err
	}

	return storedKey, nil
}
