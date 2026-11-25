package service

import (
	"strings"
	"time"

	"github.com/Arya-teja/monetra/backend/internal/domain"
)

const (
	dateLayout  = "2006-01-02"
	monthLayout = "2006-01"
)

// parseDate parses a "YYYY-MM-DD" date. On failure it records a message under
// "occurredAt" in fields.
func parseDate(value string, fields map[string]string) time.Time {
	date, err := time.Parse(dateLayout, strings.TrimSpace(value))
	if err != nil {
		fields["occurredAt"] = "must be a date in YYYY-MM-DD format"
	}
	return date
}

// parseMonth turns "YYYY-MM" into the first day of that month. On failure it
// records a message under "month" in fields and reports false.
func parseMonth(value string, fields map[string]string) (time.Time, bool) {
	start, err := time.Parse(monthLayout, value)
	if err != nil {
		fields["month"] = "must be YYYY-MM"
		return time.Time{}, false
	}
	return start, true
}

// requireMonth parses the month of an operation whose only input is that month.
func requireMonth(value string) (time.Time, error) {
	fields := map[string]string{}
	start, ok := parseMonth(value, fields)
	if !ok {
		return time.Time{}, &domain.ValidationError{Fields: fields}
	}
	return start, nil
}
