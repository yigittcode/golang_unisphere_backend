package dberrors

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn" // Import pgconn for PgError
)

// IsDuplicateConstraintError checks if the error is a PostgreSQL unique violation error
// for a specific constraint.
func IsDuplicateConstraintError(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	// Check if the error is a PgError, if the code is unique_violation (23505),
	// and if the constraint name matches the provided one.
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == constraintName
}
