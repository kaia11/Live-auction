package repository

import (
	"database/sql"
	"time"
)

const mysqlDateTimeLayout = "2006-01-02 15:04:05"

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

func nullableTimeValue(value string) any {
	if value == "" {
		return nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Format(mysqlDateTimeLayout)
	}
	if parsed, err := time.Parse(mysqlDateTimeLayout, value); err == nil {
		return parsed.Format(mysqlDateTimeLayout)
	}
	return value
}

func nullableInt64Value(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}
