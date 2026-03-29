package s3

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Config holds S3 or S3-compatible (MinIO, R2, etc.) client settings.
type Config struct {
	Bucket        string
	Region        string
	Endpoint      string // Optional. Custom API base URL, e.g. https://<account>.r2.cloudflarestorage.com
	UsePathStyle  bool   // Often required for MinIO; R2 may use either depending on setup
	PublicBaseURL string // Optional. Prefix for browser-facing URLs after upload (R2 public bucket, CDN, etc.)
}

// S3Storage implements domain.PhotoStorage using AWS S3 or an S3-compatible API.
type S3Storage struct {
	client        *s3.Client
	bucketName    string
	region        string
	endpoint      string
	publicBaseURL string
}

// NewS3Storage creates storage from Config.
func NewS3Storage(ctx context.Context, cfg Config) (*S3Storage, error) {
	loadCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
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
		client:        client,
		bucketName:    cfg.Bucket,
		region:        cfg.Region,
		endpoint:      strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/"),
		publicBaseURL: strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"),
	}, nil
}

// publicObjectURL builds the URL returned after upload (browser / menu img src).
func (s *S3Storage) publicObjectURL(key string) string {
	if s.publicBaseURL != "" {
		return s.publicBaseURL + "/" + key
	}
	if s.endpoint != "" {
		// Path-style: https://endpoint/bucket/key
		return s.endpoint + "/" + s.bucketName + "/" + key
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, key)
}

// Upload uploads a file to S3 and returns a public URL (or path reachable from your CDN / worker).
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

	return s.publicObjectURL(key), nil
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
