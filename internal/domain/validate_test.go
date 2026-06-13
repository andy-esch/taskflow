package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateField_OK(t *testing.T) {
	ok := []struct{ field, value string }{
		{"priority", "high"}, {"priority", "low"},
		{"tier", "1"}, {"tier", "5"},
		{"autonomy_level", "3"},
		{"description", "a fine one-line description"},
		{"effort", "anything goes"}, // unconstrained
		{"tags", "unconstrained"},
	}
	for _, c := range ok {
		if err := ValidateField(c.field, c.value); err != nil {
			t.Errorf("ValidateField(%q, %q) = %v, want nil", c.field, c.value, err)
		}
	}
}

func TestValidateField_Invalid(t *testing.T) {
	bad := []struct{ field, value string }{
		{"status", "in-progress"}, // must go via move, not set
		{"priority", "urgent"},
		{"tier", "9"},
		{"tier", "notanumber"},
		{"autonomy_level", "0"},
		{"description", strings.Repeat("x", MaxDescriptionLen+1)},
		{"description", "two\nlines"},
		{"created", "yesterday"},     // date fields must be YYYY-MM-DD
		{"updated_at", "2026/06/09"}, // wrong separator
	}
	for _, c := range bad {
		err := ValidateField(c.field, c.value)
		if err == nil || !errors.Is(err, ErrValidation) {
			t.Errorf("ValidateField(%q, %q) = %v, want ErrValidation", c.field, c.value, err)
		}
	}
}

func TestTypedValidators(t *testing.T) {
	for _, ok := range []func() error{
		func() error { return ValidateTier(1) }, func() error { return ValidateTier(5) },
		func() error { return ValidateAutonomy(3) },
		func() error { return ValidatePriority("medium") },
		func() error { return ValidateDescription("fine") },
		func() error { return ValidateDate("2026-06-09") },
	} {
		if err := ok(); err != nil {
			t.Errorf("valid input rejected: %v", err)
		}
	}
	for name, bad := range map[string]func() error{
		"tier 0":      func() error { return ValidateTier(0) },
		"tier 6":      func() error { return ValidateTier(6) },
		"autonomy 9":  func() error { return ValidateAutonomy(9) },
		"prio urgent": func() error { return ValidatePriority("urgent") },
		"multiline":   func() error { return ValidateDescription("a\nb") },
		"bad date":    func() error { return ValidateDate("nope") },
	} {
		if err := bad(); !errors.Is(err, ErrValidation) {
			t.Errorf("%s: want ErrValidation, got %v", name, err)
		}
	}
}

func TestStatus_IsActive(t *testing.T) {
	for _, s := range []Status{StatusNextUp, StatusReadyToStart, StatusInProgress} {
		if !s.IsActive() {
			t.Errorf("%s should be active", s)
		}
	}
	for _, s := range []Status{StatusCompleted, StatusDeprecated, StatusDeferred, Status("")} {
		if s.IsActive() {
			t.Errorf("%s should be inactive", s)
		}
	}
}

// TestValidateDescription_CountsCharacters pins the byte→rune cap change: a
// multibyte description must not hit the 150 cap early just for being UTF-8.
func TestValidateDescription_CountsCharacters(t *testing.T) {
	cjk := strings.Repeat("情", 140) // 140 chars, 420 bytes
	if err := ValidateDescription(cjk); err != nil {
		t.Errorf("140 CJK chars should pass the 150-char cap, got %v", err)
	}
	if err := ValidateDescription(strings.Repeat("a", 151)); err == nil {
		t.Error("151 chars should fail the cap")
	}
}
