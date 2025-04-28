package helpers

import "database/sql"

// GetNullString converts a string pointer to sql.NullString.
// If the pointer is nil, returns an empty NullString.
// Otherwise, returns a valid NullString with the pointer's value.
func GetNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// GetContentNullString converts a string value to sql.NullString.
// If the string is empty, returns an empty NullString.
// Otherwise, returns a valid NullString with the string value.
func GetContentNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// GetNullInt64 converts an int64 to sql.NullInt64.
// If the value is 0, returns an empty NullInt64.
// Otherwise, returns a valid NullInt64 with the value.
func GetNullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}
