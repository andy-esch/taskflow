package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// MaxDescriptionLen is the frontmatter description cap, in characters.
const MaxDescriptionLen = 200

var validPriorities = map[string]bool{"high": true, "medium": true, "low": true}

// Typed validators are the canonical field rules. Creation paths call them with
// native types (no string round-trip); ValidateField and LintTask delegate here
// so the rules live in exactly one place.

// ValidateTier checks a 1–5 tier.
func ValidateTier(tier int) error {
	if tier < 1 || tier > 5 {
		return fmt.Errorf("%w: tier must be 1-5, got %d", ErrValidation, tier)
	}
	return nil
}

// ValidateAutonomy checks a 1–5 autonomy level.
func ValidateAutonomy(level int) error {
	if level < 1 || level > 5 {
		return fmt.Errorf("%w: autonomy_level must be 1-5, got %d", ErrValidation, level)
	}
	return nil
}

// ValidatePriority checks the priority enum.
func ValidatePriority(p string) error {
	if !validPriorities[p] {
		return fmt.Errorf("%w: priority must be high|medium|low, got %q", ErrValidation, p)
	}
	return nil
}

// ValidateDescription checks the single-line + length rules.
func ValidateDescription(d string) error {
	if strings.ContainsAny(d, "\r\n") {
		return fmt.Errorf("%w: description must be a single line", ErrValidation)
	}
	if n := utf8.RuneCountInString(d); n > MaxDescriptionLen {
		// Count CHARACTERS, not bytes — a non-ASCII description must not hit the
		// cap early just for being multibyte.
		return fmt.Errorf("%w: description too long (%d > %d chars)", ErrValidation, n, MaxDescriptionLen)
	}
	return nil
}

// ActiveTaskFieldErr enforces the field invariants that creation guarantees and
// LintTask flags on active tasks: every active task needs at least one tag, and a
// next-up/in-progress task needs a non-empty description. Both NewTask (creation)
// and the SetFields write path call it, so the create and mutate paths cannot
// drift apart and write a file the tool's own linter immediately rejects. Archived
// tasks (completed/deprecated/deferred) are not held to these active-only rules,
// mirroring Service.Lint (which only lints active tasks). Returns nil when met.
func ActiveTaskFieldErr(t Task) error {
	if !t.Status.IsActive() {
		return nil
	}
	if len(t.Tags) == 0 {
		return fmt.Errorf("%w: at least one tag is required", ErrValidation)
	}
	if (t.Status == StatusNextUp || t.Status == StatusInProgress) && strings.TrimSpace(t.Description) == "" {
		return fmt.Errorf("%w: a description is required for a next-up/in-progress task", ErrValidation)
	}
	return nil
}

// ValidateDate checks a YYYY-MM-DD date string.
func ValidateDate(value string) error {
	if _, err := time.Parse(time.DateOnly, value); err != nil {
		return fmt.Errorf("%w: must be YYYY-MM-DD, got %q", ErrValidation, value)
	}
	return nil
}

var dateFields = map[string]bool{
	"created": true, "updated_at": true, "started_at": true,
	"completed_at": true, "deprecated_at": true, "deferred_at": true, "audited": true,
}

// ValidateField checks a constrained frontmatter field from its string value —
// the `task set` path, where every value arrives as a string. It delegates to
// the typed validators; unconstrained fields pass.
func ValidateField(field, value string) error {
	switch {
	case field == "status":
		return fmt.Errorf("%w: set status with `task <verb>`/`task move`, not `set`", ErrValidation)
	case field == "updated_at":
		// Stamped on every mutation; an explicit value would validate and then be
		// silently clobbered — reject it instead (decided 2026-06-12).
		return fmt.Errorf("%w: updated_at is stamped automatically and cannot be set", ErrValidation)
	case field == "priority":
		return ValidatePriority(value)
	case field == "tier" || field == "autonomy_level":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("%w: %s must be an integer 1-5, got %q", ErrValidation, field, value)
		}
		if field == "tier" {
			return ValidateTier(n)
		}
		return ValidateAutonomy(n)
	case field == "description":
		return ValidateDescription(value)
	case dateFields[field]:
		return ValidateDate(value)
	}
	return nil
}
