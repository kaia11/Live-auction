package repository

import (
	"database/sql"
	"time"
)

func nullableString(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func nullableInt64Ptr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}

func nullableTimeString(value sql.NullTime) string {
	if value.Valid {
		return value.Time.Format(time.RFC3339)
	}
	return ""
}

func nullableEmptyToNull(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableInt64Value(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}
