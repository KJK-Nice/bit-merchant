package main

import (
	"net/url"
	"os"
	"strconv"
	"strings"
)

// firstEnv returns the first non-empty trimmed value from the given keys.
func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

type serverConfig struct {
	Port              string
	BaseURL           string
	RPID              string
	ForceSecureCookie bool
	DatabaseURL       string
	S3BucketName      string
	AWSRegion         string
	// S3-compatible API (optional; non-AWS providers)
	S3Endpoint             string
	S3UsePathStyle         bool
	S3PublicBaseURL        string
	S3PresignGetExpiresSec int // unset/0 → 1-hour presign TTL (initPhotoStorage)
}

func loadConfig() (serverConfig, error) {
	cfg := serverConfig{
		Port:              os.Getenv("PORT"),
		BaseURL:           os.Getenv("BASE_URL"),
		ForceSecureCookie: os.Getenv("COOKIE_SECURE") == "true",
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		S3BucketName:      firstEnv("AWS_S3_BUCKET_NAME", "S3_BUCKET_NAME"),
		AWSRegion:         firstEnv("AWS_DEFAULT_REGION", "AWS_REGION"),
		S3Endpoint:        firstEnv("AWS_ENDPOINT_URL", "S3_ENDPOINT"),
		S3PublicBaseURL:   strings.TrimSpace(os.Getenv("S3_PUBLIC_BASE_URL")),
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("S3_USE_PATH_STYLE"))) {
	case "true", "1", "yes":
		cfg.S3UsePathStyle = true
	case "false", "0", "no":
		cfg.S3UsePathStyle = false
	default:
		// Path-style is the safe default for most S3-compatible endpoints (MinIO, many gateways).
		cfg.S3UsePathStyle = cfg.S3Endpoint != ""
	}

	if v := strings.TrimSpace(os.Getenv("S3_PRESIGN_GET_EXPIRES")); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			cfg.S3PresignGetExpiresSec = sec
		}
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080"
	}

	parsedBaseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return serverConfig{}, err
	}

	cfg.RPID = parsedBaseURL.Hostname()
	if cfg.RPID == "" {
		cfg.RPID = "localhost"
	}

	return cfg, nil
}
