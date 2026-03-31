package query

import (
	"net/url"
	"strings"
)

// PhotoObjectKeyFromStoredValue returns an S3 object key for PresignGet.
func PhotoObjectKeyFromStoredValue(stored, bucket, endpoint, publicBaseURL string) string {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return ""
	}
	lower := strings.ToLower(stored)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return stored
	}
	u, err := url.Parse(stored)
	if err != nil {
		return stored
	}
	if key, ok := stripURLPrefix(stored, publicBaseURL, endpoint, bucket); ok {
		return key
	}
	path := strings.Trim(u.Path, "/")
	if key, ok := extractKeyFromPath(path, bucket, u.Host); ok {
		return key
	}
	if path != "" {
		return path
	}
	return stored
}

func stripURLPrefix(stored, publicBaseURL, endpoint, bucket string) (string, bool) {
	publicBaseURL = strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if publicBaseURL != "" && strings.HasPrefix(stored, publicBaseURL+"/") {
		return strings.TrimPrefix(stored, publicBaseURL+"/"), true
	}
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	bucket = strings.TrimSpace(bucket)
	if endpoint != "" && bucket != "" {
		prefix := endpoint + "/" + bucket + "/"
		if strings.HasPrefix(stored, prefix) {
			return strings.TrimPrefix(stored, prefix), true
		}
	}
	return "", false
}

func extractKeyFromPath(path, bucket, host string) (string, bool) {
	if bucket != "" && strings.HasPrefix(host, bucket+".") {
		return path, true
	}
	if bucket != "" && path != "" {
		if i := strings.IndexByte(path, '/'); i > 0 && path[:i] == bucket {
			return path[i+1:], true
		}
	}
	return "", false
}
