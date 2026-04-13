package menu_test

import (
	menuQuery "bitmerchant/internal/menu/app/query"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPhotoObjectKeyFromStoredValue(t *testing.T) {
	t.Run("plain_key_unchanged", func(t *testing.T) {
		assert.Equal(t, "restaurants/r1/items/k_foo.jpg",
			menuQuery.PhotoObjectKeyFromStoredValue("restaurants/r1/items/k_foo.jpg", "b", "https://ex", "https://pub"))
	})

	t.Run("public_base_prefix", func(t *testing.T) {
		assert.Equal(t, "restaurants/r1/x.jpg",
			menuQuery.PhotoObjectKeyFromStoredValue("https://pub.foo/restaurants/r1/x.jpg", "b", "", "https://pub.foo"))
	})

	t.Run("path_style_endpoint_bucket", func(t *testing.T) {
		assert.Equal(t, "restaurants/r1/x.jpg",
			menuQuery.PhotoObjectKeyFromStoredValue("https://t3.example/mybucket/restaurants/r1/x.jpg", "mybucket", "https://t3.example", ""))
	})

	t.Run("virtual_hosted", func(t *testing.T) {
		assert.Equal(t, "restaurants/r1/x.jpg",
			menuQuery.PhotoObjectKeyFromStoredValue("https://mybucket.s3.us-east-1.amazonaws.com/restaurants/r1/x.jpg", "mybucket", "", ""))
	})

	t.Run("path_without_virtual_host_first_segment_bucket", func(t *testing.T) {
		assert.Equal(t, "a/b",
			menuQuery.PhotoObjectKeyFromStoredValue("https://example.com/buck/a/b", "buck", "", ""))
	})
}
