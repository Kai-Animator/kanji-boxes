package domain

import "testing"

func mustDate(t *testing.T, value string) LocalDate {
	t.Helper()
	date, err := ParseLocalDate(value)
	if err != nil {
		t.Fatalf("failed to parse date %q: %v", value, err)
	}
	return date
}
