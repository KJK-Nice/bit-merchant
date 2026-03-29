package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var dbNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// EnsureDatabaseExists ensures the database in DATABASE_URL exists.
// Useful in dev containers where postgres data volume already exists but target DB does not.
// Returns true when the database had to be created.
func EnsureDatabaseExists(ctx context.Context, databaseURL string) (bool, error) {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return false, err
	}

	dbName := strings.TrimPrefix(parsed.Path, "/")
	if dbName == "" {
		return false, errors.New("database name missing in DATABASE_URL")
	}
	if !dbNamePattern.MatchString(dbName) {
		return false, fmt.Errorf("unsafe database name: %s", dbName)
	}

	adminURL := *parsed
	adminURL.Path = "/postgres"
	adminDB, err := sql.Open("pgx", adminURL.String())
	if err != nil {
		return false, err
	}
	defer adminDB.Close()

	if err := adminDB.PingContext(ctx); err != nil {
		return false, err
	}

	var exists bool
	if err := adminDB.QueryRowContext(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`,
		dbName,
	).Scan(&exists); err != nil {
		return false, err
	}

	if exists {
		return false, nil
	}

	// dbName is restricted by regex above, so quoting is safe here.
	_, err = adminDB.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
	if err != nil {
		return false, err
	}
	return true, nil
}
