package main

import (
	"context"
	"database/sql"
	"time"

	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/migrations"
	s3Storage "bitmerchant/internal/infrastructure/storage/s3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func initPhotoStorage(cfg serverConfig, logger *logging.Logger) (domain.PhotoStorage, error) {
	if cfg.S3BucketName == "" || cfg.AWSRegion == "" {
		logger.Info("S3 config missing, photo uploads will fail")
		return nil, nil
	}

	photoStorage, err := s3Storage.NewS3Storage(context.Background(), cfg.S3BucketName, cfg.AWSRegion)
	if err != nil {
		return nil, err
	}

	return photoStorage, nil
}

func connectDatabase(cfg serverConfig, logger *logging.Logger) (*sql.DB, error) {
	if cfg.DatabaseURL == "" {
		return nil, nil
	}

	bootstrapCtx, bootstrapCancel := context.WithTimeout(context.Background(), 10*time.Second)
	createdDB, bootstrapErr := migrations.EnsureDatabaseExists(bootstrapCtx, cfg.DatabaseURL)
	bootstrapCancel()
	if bootstrapErr != nil {
		return nil, bootstrapErr
	}

	if createdDB {
		logger.Info("Created missing database from DATABASE_URL")
	} else {
		logger.Info("Database from DATABASE_URL already exists")
	}

	db, dbErr := sql.Open("pgx", cfg.DatabaseURL)
	if dbErr != nil {
		return nil, dbErr
	}

	const (
		maxPingAttempts = 15
		pingDelay       = 2 * time.Second
	)

	var pingErr error
	for attempt := 1; attempt <= maxPingAttempts; attempt++ {
		pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		pingErr = db.PingContext(pingCtx)
		cancel()
		if pingErr == nil {
			break
		}

		logger.Warn("Database ping failed, retrying", "attempt", attempt, "maxAttempts", maxPingAttempts, "error", pingErr)
		time.Sleep(pingDelay)
	}

	if pingErr != nil {
		_ = db.Close()
		return nil, pingErr
	}

	if migrationErr := migrations.Up(context.Background(), db); migrationErr != nil {
		_ = db.Close()
		return nil, migrationErr
	}

	return db, nil
}
