package wiring

// Config is runtime configuration for the composition root (DB, URLs, S3, WebAuthn, cookies).
type Config struct {
	PublicBaseURL          string
	CustomerBaseURL        string
	MerchantBaseURL        string
	RPID                   string
	ForceSecureCookie      bool
	DatabaseURL            string
	S3BucketName           string
	AWSRegion              string
	S3Endpoint             string
	S3UsePathStyle         bool
	S3PublicBaseURL        string
	S3PresignGetExpiresSec int
}
