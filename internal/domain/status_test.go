package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestParseStatus_Valid(t *testing.T) {
	for _, s := range AllStatuses() {
		got, err := ParseStatus(string(s))
		if err != nil || got != s {
			t.Errorf("ParseStatus(%q) = %v, %v", s, got, err)
		}
	}
}

// TestParseStatus_InvalidWrapsValidationAndEnumerates pins the error contract:
// a typo'd status must map to exit 11 (ErrValidation) and the message must
// list every valid status, so the caller can self-correct.
func TestParseStatus_InvalidWrapsValidationAndEnumerates(t *testing.T) {
	_, err := ParseStatus("bogus")
	if err == nil {
		t.Fatal("expected an error for an invalid status")
	}
	if !errors.Is(err, ErrValidation) {
		t.Errorf("invalid status must wrap ErrValidation, got %v", err)
	}
	for _, s := range AllStatuses() {
		if !strings.Contains(err.Error(), string(s)) {
			t.Errorf("error should enumerate %q: %v", s, err)
		}
	}
}
