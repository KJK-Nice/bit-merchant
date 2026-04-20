package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
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

	EventBusBackend      string
	NATSURL              string
	NATSAutoProvision    bool
	NATSAckWait          time.Duration
	NATSCloseTimeout     time.Duration
	NATSSubscribersCount int
	NATSInstanceID       string
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

func resolveEventBusBackend() (string, error) {
	backend := strings.ToLower(strings.TrimSpace(os.Getenv("EVENT_BUS_BACKEND")))
	if backend == "" {
		return "nats", nil
	}
	if backend != "memory" && backend != "nats" {
		return "", fmt.Errorf("invalid EVENT_BUS_BACKEND %q: expected memory or nats", backend)
	}
	return backend, nil
}

func resolveBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch v {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return fallback
	}
}

func resolveInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func resolveDuration(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(v)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func resolveNATSInstanceID() string {
	if id := strings.TrimSpace(os.Getenv("NATS_INSTANCE_ID")); id != "" {
		return id
	}
	hostname, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostname) == "" {
		return "bitmerchant"
	}
	return hostname
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
	eventBusBackend, err := resolveEventBusBackend()
	if err != nil {
		return serverConfig{}, err
	}

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
		EventBusBackend:        eventBusBackend,
		NATSURL:                strings.TrimSpace(firstEnv("NATS_URL")),
		NATSAutoProvision:      resolveBool("NATS_AUTO_PROVISION", true),
		NATSAckWait:            resolveDuration("NATS_ACK_WAIT", 30*time.Second),
		NATSCloseTimeout:       resolveDuration("NATS_CLOSE_TIMEOUT", 30*time.Second),
		NATSSubscribersCount:   resolveInt("NATS_SUBSCRIBERS_COUNT", 1),
		NATSInstanceID:         resolveNATSInstanceID(),
	}
	if cfg.NATSURL == "" {
		cfg.NATSURL = "nats://localhost:4222"
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
