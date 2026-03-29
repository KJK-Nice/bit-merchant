package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"bitmerchant/internal/infrastructure/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	bootstrapCtx, bootstrapCancel := context.WithTimeout(context.Background(), 10*time.Second)
	createdDB, bootstrapErr := migrations.EnsureDatabaseExists(bootstrapCtx, databaseURL)
	bootstrapCancel()
	if bootstrapErr != nil {
		log.Fatalf("failed to ensure database exists: %v", bootstrapErr)
	}
	if createdDB {
		log.Println("created missing database from DATABASE_URL")
	} else {
		log.Println("database from DATABASE_URL already exists")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	if err := migrations.Up(context.Background(), db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	log.Println("migrations completed successfully")
}
