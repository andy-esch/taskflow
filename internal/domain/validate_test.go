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
	}
	for _, c := range bad {
		err := ValidateField(c.field, c.value)
		if err == nil || !errors.Is(err, ErrValidation) {
			t.Errorf("ValidateField(%q, %q) = %v, want ErrValidation", c.field, c.value, err)
		}
	}
}
