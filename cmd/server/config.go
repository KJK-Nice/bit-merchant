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
	PublicBaseURL     string
	CustomerBaseURL   string
	MerchantBaseURL   string
	RPID              string
	ForceSecureCookie bool
	DisableRateLimit  bool
	DatabaseURL       string
	S3BucketName      string
	AWSRegion         string
	// S3-compatible API (optional; non-AWS providers)
	S3Endpoint             string
	S3UsePathStyle         bool
	S3PublicBaseURL        string
	S3PresignGetExpiresSec int // unset/0 → 1-hour presign TTL (initPhotoStorage)
}

func loadBaseURL() string {
	base := strings.TrimSpace(firstEnv("BASE_URL", "PUBLIC_BASE_URL"))
	if base == "" {
		return "http://localhost:8080"
	}
	return base
}

func resolvePort(port string) string {
	if strings.TrimSpace(port) == "" {
		return "8080"
	}
	return strings.TrimSpace(port)
}

func resolvePathStyleDefault(s3Endpoint string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("S3_USE_PATH_STYLE"))) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		// Path-style is the safe default for most S3-compatible endpoints (MinIO, many gateways).
		return s3Endpoint != ""
	}
}

func resolvePresignExpiresSec() int {
	v := strings.TrimSpace(os.Getenv("S3_PRESIGN_GET_EXPIRES"))
	if v == "" {
		return 0
	}
	sec, err := strconv.Atoi(v)
	if err != nil || sec <= 0 {
		return 0
	}
	return sec
}

func normalizeSurfaceURLs(base, public, customer, merchant string) (string, string, string) {
	if public == "" {
		public = base
	}
	if customer == "" {
		customer = base
	}
	if merchant == "" {
		merchant = base
	}
	return public, customer, merchant
}

func loadConfig() (serverConfig, error) {
	base := loadBaseURL()

	cfg := serverConfig{
		Port:                   resolvePort(os.Getenv("PORT")),
		BaseURL:                base,
		PublicBaseURL:          strings.TrimSpace(firstEnv("PUBLIC_BASE_URL", "BASE_URL")),
		CustomerBaseURL:        strings.TrimSpace(firstEnv("CUSTOMER_BASE_URL", "BASE_URL", "PUBLIC_BASE_URL")),
		MerchantBaseURL:        strings.TrimSpace(firstEnv("MERCHANT_BASE_URL", "BASE_URL", "PUBLIC_BASE_URL")),
		ForceSecureCookie:      os.Getenv("COOKIE_SECURE") == "true",
		DisableRateLimit:       os.Getenv("DISABLE_RATE_LIMIT") == "true",
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		S3BucketName:           firstEnv("AWS_S3_BUCKET_NAME", "S3_BUCKET_NAME"),
		AWSRegion:              firstEnv("AWS_DEFAULT_REGION", "AWS_REGION"),
		S3Endpoint:             firstEnv("AWS_ENDPOINT_URL", "S3_ENDPOINT"),
		S3UsePathStyle:         resolvePathStyleDefault(firstEnv("AWS_ENDPOINT_URL", "S3_ENDPOINT")),
		S3PublicBaseURL:        strings.TrimSpace(os.Getenv("S3_PUBLIC_BASE_URL")),
		S3PresignGetExpiresSec: resolvePresignExpiresSec(),
	}
	cfg.PublicBaseURL, cfg.CustomerBaseURL, cfg.MerchantBaseURL = normalizeSurfaceURLs(cfg.BaseURL, cfg.PublicBaseURL, cfg.CustomerBaseURL, cfg.MerchantBaseURL)

	parsedBaseURL, err := url.Parse(cfg.MerchantBaseURL)
	if err != nil {
		return serverConfig{}, err
	}

	cfg.RPID = parsedBaseURL.Hostname()
	if cfg.RPID == "" {
		cfg.RPID = "localhost"
	}

	return cfg, nil
}
