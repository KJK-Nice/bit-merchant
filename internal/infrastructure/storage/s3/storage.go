package s3

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Config holds S3 or S3-compatible (MinIO, R2, etc.) client settings.
type Config struct {
	Bucket       string
	Region       string
	Endpoint     string // Optional. Custom API base URL, e.g. https://<account>.r2.cloudflarestorage.com
	UsePathStyle bool   // Often required for MinIO; R2 may use either depending on setup
	// PresignGetExpires is how long presigned GET URLs remain valid (default 1h if zero).
	PresignGetExpires time.Duration
}

// S3Storage implements domain.PhotoStorage using AWS S3 or an S3-compatible API.
type S3Storage struct {
	client         *s3.Client
	bucketName     string
	region         string
	endpoint       string
	presignExpires time.Duration
}

// NewS3Storage creates storage from Config.
func NewS3Storage(ctx context.Context, cfg Config) (*S3Storage, error) {
	loadCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	presignExp := cfg.PresignGetExpires
	if presignExp <= 0 {
		presignExp = time.Hour
	}

	var client *s3.Client
	if strings.TrimSpace(cfg.Endpoint) != "" {
		ep := strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/")
		client = s3.NewFromConfig(loadCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(ep)
			o.UsePathStyle = cfg.UsePathStyle
		})
	} else {
		client = s3.NewFromConfig(loadCfg)
	}

	return &S3Storage{
		client:         client,
		bucketName:     cfg.Bucket,
		region:         cfg.Region,
		endpoint:       strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/"),
		presignExpires: presignExp,
	}, nil
}

// Upload uploads a file to S3 and returns the object key to store in the database.
func (s *S3Storage) Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        data,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return key, nil
}

// PresignGet returns a time-limited URL to GET the object (for private buckets).
func (s *S3Storage) PresignGet(ctx context.Context, key string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", fmt.Errorf("presign get: empty key")
	}
	presignClient := s3.NewPresignClient(s.client)
	out, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(s.presignExpires))
	if err != nil {
		return "", fmt.Errorf("presign get object: %w", err)
	}
	return out.URL, nil
}

// Delete deletes a file from S3
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}
	return nil
}
