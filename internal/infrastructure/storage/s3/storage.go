package s3

import menuAdapters "bitmerchant/internal/menu/adapters"

type Config = menuAdapters.S3Config
type S3Storage = menuAdapters.S3Storage

var NewS3Storage = menuAdapters.NewS3Storage
