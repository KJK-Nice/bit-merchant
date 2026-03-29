package main

import (
	"net/url"
	"os"
)

type serverConfig struct {
	Port              string
	BaseURL           string
	RPID              string
	ForceSecureCookie bool
	DatabaseURL       string
	S3BucketName      string
	AWSRegion         string
}

func loadConfig() (serverConfig, error) {
	cfg := serverConfig{
		Port:              os.Getenv("PORT"),
		BaseURL:           os.Getenv("BASE_URL"),
		ForceSecureCookie: os.Getenv("COOKIE_SECURE") == "true",
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		S3BucketName:      os.Getenv("S3_BUCKET_NAME"),
		AWSRegion:         os.Getenv("AWS_REGION"),
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
