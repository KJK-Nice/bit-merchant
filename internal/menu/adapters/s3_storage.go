package adapters

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

type S3Config struct {
	Bucket            string
	Region            string
	Endpoint          string
	UsePathStyle      bool
	PresignGetExpires time.Duration
}

type S3Storage struct {
	client         *s3.Client
	bucketName     string
	presignExpires time.Duration
}

func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
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
		presignExpires: presignExp,
	}, nil
}

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
