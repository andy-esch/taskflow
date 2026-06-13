package domain

import (
	"errors"
	"testing"
)

func TestParseAuditBucket_Valid(t *testing.T) {
	for _, b := range AllAuditBuckets() {
		got, err := ParseAuditBucket(string(b))
		if err != nil || got != b {
			t.Errorf("ParseAuditBucket(%q) = %v, %v", b, got, err)
		}
	}
}

// TestParseAuditBucket_InvalidWrapsValidation pins the error contract: a bad
// bucket must map to exit 11 (ErrValidation), like ParseStatus — otherwise it
// escapes as a generic exit 1 and breaks code-based routing.
func TestParseAuditBucket_InvalidWrapsValidation(t *testing.T) {
	_, err := ParseAuditBucket("bogus")
	if err == nil {
		t.Fatal("expected an error for an invalid audit bucket")
	}
	if !errors.Is(err, ErrValidation) {
		t.Errorf("invalid audit bucket must wrap ErrValidation, got %v", err)
	}
}
