package menu

import (
	"context"
	"io"
)

// PhotoStorage defines operations for storing photos.
type PhotoStorage interface {
	Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error)
	Delete(ctx context.Context, key string) error
	PresignGet(ctx context.Context, key string) (string, error)
}
