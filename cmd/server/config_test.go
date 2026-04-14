package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_SurfaceURLs(t *testing.T) {
	t.Setenv("BASE_URL", "https://bitmerchant.com")
	t.Setenv("PUBLIC_BASE_URL", "https://bitmerchant.com")
	t.Setenv("CUSTOMER_BASE_URL", "https://order.bitmerchant.com")
	t.Setenv("MERCHANT_BASE_URL", "https://merchant.bitmerchant.com")

	cfg, err := loadConfig()
	require.NoError(t, err)

	assert.Equal(t, "https://bitmerchant.com", cfg.PublicBaseURL)
	assert.Equal(t, "https://order.bitmerchant.com", cfg.CustomerBaseURL)
	assert.Equal(t, "https://merchant.bitmerchant.com", cfg.MerchantBaseURL)
	assert.Equal(t, "merchant.bitmerchant.com", cfg.RPID)
}

func TestLoadConfig_SurfaceURLFallbacksToBaseURL(t *testing.T) {
	for _, key := range []string{"BASE_URL", "PUBLIC_BASE_URL", "CUSTOMER_BASE_URL", "MERCHANT_BASE_URL"} {
		require.NoError(t, os.Unsetenv(key))
	}
	t.Setenv("BASE_URL", "http://localhost:8080")

	cfg, err := loadConfig()
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:8080", cfg.PublicBaseURL)
	assert.Equal(t, "http://localhost:8080", cfg.CustomerBaseURL)
	assert.Equal(t, "http://localhost:8080", cfg.MerchantBaseURL)
	assert.Equal(t, "localhost", cfg.RPID)
}
