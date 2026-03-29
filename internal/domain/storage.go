package domain

import (
	"context"
	"io"
)

// PhotoStorage defines operations for storing photos.
// Upload returns the S3 object key to persist in the database (presign at read time for private buckets).
type PhotoStorage interface {
	Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error)
	Delete(ctx context.Context, key string) error
	PresignGet(ctx context.Context, key string) (string, error)
}
