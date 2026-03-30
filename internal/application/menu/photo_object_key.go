package menu

import (
	"net/url"
	"strings"
)

// PhotoObjectKeyFromStoredValue returns an S3 object key for PresignGet.
// If stored is already a key (no scheme), it is returned unchanged.
// If stored is a legacy public URL from this app, the key portion is extracted using
// bucket, endpoint, and publicBaseURL (same semantics as former publicObjectURL).
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

	publicBaseURL = strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if publicBaseURL != "" && strings.HasPrefix(stored, publicBaseURL+"/") {
		return strings.TrimPrefix(stored, publicBaseURL+"/")
	}

	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	bucket = strings.TrimSpace(bucket)
	if endpoint != "" && bucket != "" {
		prefix := endpoint + "/" + bucket + "/"
		if strings.HasPrefix(stored, prefix) {
			return strings.TrimPrefix(stored, prefix)
		}
	}

	path := strings.Trim(u.Path, "/")
	if bucket != "" && strings.HasPrefix(u.Host, bucket+".") {
		return path
	}

	if bucket != "" && path != "" {
		if i := strings.IndexByte(path, '/'); i > 0 && path[:i] == bucket {
			return path[i+1:]
		}
	}

	if path != "" {
		return path
	}
	return stored
}
