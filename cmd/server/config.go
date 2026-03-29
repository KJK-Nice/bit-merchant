package main

import (
	"net/url"
	"os"
	"strings"
)

type serverConfig struct {
	Port              string
	BaseURL           string
	RPID              string
	ForceSecureCookie bool
	DatabaseURL       string
	S3BucketName      string
	AWSRegion         string
	// S3-compatible API (optional; non-AWS providers)
	S3Endpoint       string
	S3UsePathStyle   bool
	S3PublicBaseURL  string
}

func loadConfig() (serverConfig, error) {
	cfg := serverConfig{
		Port:              os.Getenv("PORT"),
		BaseURL:           os.Getenv("BASE_URL"),
		ForceSecureCookie: os.Getenv("COOKIE_SECURE") == "true",
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		S3BucketName:      os.Getenv("S3_BUCKET_NAME"),
		AWSRegion:         os.Getenv("AWS_REGION"),
		S3Endpoint:        strings.TrimSpace(os.Getenv("S3_ENDPOINT")),
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
