package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrPostNotFound    = errors.New("post not found")
	ErrCommentNotFound = errors.New("comment not found")

	ErrDatabaseConnection  = errors.New("database connection error")
	ErrDatabaseQuery       = errors.New("database query error")
	ErrDuplicateKey        = errors.New("duplicate key constraint violation")
	ErrForeignKeyViolation = errors.New("foreign key constraint violation")
	ErrUniqueViolation     = errors.New("unique constraint violation")
)

// MapPgError maps PostgreSQL errors to application errors
func MapPgError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		// Not a PostgreSQL error
		return ErrDatabaseQuery
	}

	// Map PostgreSQL error codes to application errors
	switch pgErr.SQLState() {
	case "23505": // UNIQUE VIOLATION
		return ErrUniqueViolation
	case "23503": // FOREIGN KEY VIOLATION
		return ErrForeignKeyViolation
	case "23502": // NOT NULL VIOLATION
		return fmt.Errorf("%w: %s", ErrDatabaseQuery, pgErr.Message)
	case "42702": // AMBIGUOUS COLUMN
		return fmt.Errorf("%w: %s", ErrDatabaseQuery, pgErr.Message)
	case "08006", "08003": // CONNECTION FAILURE
		return ErrDatabaseConnection
	default:
		return fmt.Errorf("%w: %s (%s)", ErrDatabaseQuery, pgErr.Message, pgErr.SQLState())
	}
}
