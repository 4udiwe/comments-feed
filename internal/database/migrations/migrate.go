package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Конвертируем pgxpool.Pool в *sql.DB (goose требует database/sql)
	db, err := pgxPoolToStdlib(ctx, pool)
	if err != nil {
		return fmt.Errorf("failed to convert pool: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Try both paths: for production (database/migrations) and tests (internal/database/migrations)
	// Get current working directory
	cwd, _ := os.Getwd()

	// Try absolute paths first
	paths := []string{
		"/app/database/migrations",          // Production path
		"/app/internal/database/migrations", // Test path (absolute)
		"database/migrations",               // Production path (relative)
		"internal/database/migrations",      // Test path (relative)
	}

	var migrationPath string
	var found bool
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			// Check if directory has .sql files
			entries, err := os.ReadDir(path)
			if err == nil && len(entries) > 0 {
				migrationPath = path
				found = true
				log.Printf("Found migrations at: %s (cwd: %s)", migrationPath, cwd)
				break
			}
		}
	}

	if !found {
		return fmt.Errorf("migration directory not found. Tried paths: %v (cwd: %s)", paths, cwd)
	}

	if err := goose.Up(db, migrationPath); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	version, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get DB version: %w", err)
	}
	log.Printf("Migrations applied. Current version: %d", version)

	return nil
}

func pgxPoolToStdlib(ctx context.Context, pool *pgxpool.Pool) (*sql.DB, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	db := stdlib.OpenDB(*conn.Conn().Config())
	return db, nil
}
