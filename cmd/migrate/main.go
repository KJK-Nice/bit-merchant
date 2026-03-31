package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"time"

	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logging.NewLogger()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	bootstrapCtx, bootstrapCancel := context.WithTimeout(context.Background(), 10*time.Second)
	createdDB, bootstrapErr := migrations.EnsureDatabaseExists(bootstrapCtx, databaseURL)
	bootstrapCancel()
	if bootstrapErr != nil {
		slog.Error("failed to ensure database exists", "err", bootstrapErr)
		os.Exit(1)
	}
	if createdDB {
		slog.Info("created missing database from DATABASE_URL")
	} else {
		slog.Info("database from DATABASE_URL already exists")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		slog.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		slog.Error("failed to ping database", "err", err)
		os.Exit(1)
	}

	if err := migrations.Up(context.Background(), db); err != nil {
		slog.Error("failed to run migrations", "err", err)
		os.Exit(1)
	}

	slog.Info("migrations completed successfully")
}
