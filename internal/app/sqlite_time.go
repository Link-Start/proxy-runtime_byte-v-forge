package app

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func sqliteTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func sqliteTimestamp(value *timestamppb.Timestamp) string {
	if value == nil {
		return ""
	}
	return sqliteTime(value.AsTime())
}

func parseSQLiteTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed.UTC()
	}
	if parsed, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return parsed.UTC()
	}
	return time.Time{}
}

func sqliteBool(value bool) int {
	if value {
		return 1
	}
	return 0
}
