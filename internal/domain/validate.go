package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/andy-esch/taskflow/internal/id"
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

// ValidateMintableDate checks a YYYY-MM-DD date that a stable id will be MINTED FROM
// (ADR-0003 §3) — stricter than ValidateDate, which only checks the shape.
//
// An id carries 43 bits of Unix milliseconds, so only 1970-01-01 through ~2248-09-26
// round-trips. Outside that window the timestamp wraps — a pre-1970 date becomes a huge
// value and sorts last, a post-range date wraps toward zero and sorts first — silently
// destroying the property that makes a backdated id useful: that sorting by id sorts by
// authorship date. A plausible year typo (`1026-06-15` for `2026-06-15`) is exactly the
// input that trips it, so the check is worth its weight.
//
// Only minting callers need this. A date that is merely STORED (a task's `created`, an
// audit's `date`) has no such constraint and keeps using ValidateDate.
func ValidateMintableDate(value string) error {
	if err := ValidateDate(value); err != nil {
		return err
	}
	t, err := time.Parse(time.DateOnly, value)
	if err != nil { // unreachable: ValidateDate already parsed it
		return fmt.Errorf("%w: must be YYYY-MM-DD, got %q", ErrValidation, value)
	}
	if !id.Representable(t.UnixMilli()) {
		return fmt.Errorf("%w: date %q is outside the range a stable id can encode (%s to %s) — a date this far out is almost always a typo",
			ErrValidation, value, MintableDateMin, MintableDateMax)
	}
	return nil
}

// The inclusive bounds ValidateMintableDate enforces, as dates — derived from
// id.MaxMillis so the prose and the check can't drift, and exported so the authoring
// docs (`schema research`) can state the window instead of restating a magic year.
var (
	MintableDateMin = time.UnixMilli(0).UTC().Format(time.DateOnly)
	MintableDateMax = time.UnixMilli(id.MaxMillis).UTC().Format(time.DateOnly)
)

// ValidateDate checks a YYYY-MM-DD date string.
func ValidateDate(value string) error {
	if _, err := time.Parse(time.DateOnly, value); err != nil {
		return fmt.Errorf("%w: must be YYYY-MM-DD, got %q", ErrValidation, value)
	}
	return nil
}

// dateFields (the YAML "date" typed keys) is derived from the taskFields registry
// in fields.go, alongside the other per-type sets.

// IsRevisitDue reports whether a deferred task's revisit ("snooze until") date
// has arrived: revisitAt is on or before now's calendar day. An empty or invalid
// date is never due (parking a task with no revisit date stays indefinite, and a
// malformed value must not masquerade as overdue). now is injected so the
// dashboard's "is it due?" comparison stays deterministic in tests. Compared as
// calendar days so a date set for *today* nudges today, not only once it has
// strictly passed.
func IsRevisitDue(revisitAt string, now time.Time) bool {
	d, err := time.Parse(time.DateOnly, revisitAt)
	if err != nil {
		return false
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return !d.After(today)
}

// IsTaskRevisitDue reports whether a task is parked in deferred AND its revisit
// ("snooze until") date has arrived — the single definition of "due for revisit",
// shared by the status nudge, `task list --revisit-due`, and the TUI marker/view,
// so the three can't drift. A revisit_at on a non-deferred task never counts.
func IsTaskRevisitDue(t Task, now time.Time) bool {
	return t.Status == StatusDeferred && IsRevisitDue(t.RevisitAt, now)
}

// relativeDate matches a future offset typed at the interactive defer prompt: a
// count and a unit of days or weeks, optionally space-separated (e.g. "10d",
// "2 weeks"). Months are deliberately NOT supported — their calendar arithmetic
// rolls month-end dates into the following month (Jan 31 + "1m" → Mar 3), and a
// snooze only needs days/weeks granularity.
var relativeDate = regexp.MustCompile(`^(\d+)\s*(d|day|days|w|week|weeks)$`)

// ParseRevisitDate interprets a revisit ("snooze until") date typed at the
// interactive defer prompt. It accepts EITHER an absolute YYYY-MM-DD (the same
// form `--until` and `task set revisit_at=` take) OR a relative offset into the
// future — a count plus a unit of days or weeks ("10d", "2w", "3 weeks"). A blank
// input returns "" (no date — park the task indefinitely). now is injected so the
// relative arithmetic stays deterministic in tests; the result is always a
// validated YYYY-MM-DD string (or "").
func ParseRevisitDate(input string, now time.Time) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", nil // blank = no revisit date
	}
	// An absolute date wins — it's the canonical form the field is stored in.
	if _, err := time.Parse(time.DateOnly, s); err == nil {
		return s, nil
	}
	// Otherwise a relative offset: <count><unit>, computed forward from now's day.
	if m := relativeDate.FindStringSubmatch(strings.ToLower(s)); m != nil {
		n, err := strconv.Atoi(m[1])
		if err != nil { // unreachable for \d+, but keep the conversion honest
			return "", fmt.Errorf("%w: revisit offset %q: %v", ErrValidation, input, err)
		}
		var d time.Time
		switch m[2] {
		case "d", "day", "days":
			d = now.AddDate(0, 0, n)
		case "w", "week", "weeks":
			d = now.AddDate(0, 0, 7*n)
		}
		return d.Format(time.DateOnly), nil
	}
	return "", fmt.Errorf("%w: revisit date %q must be YYYY-MM-DD or a relative offset like 2w / 10d", ErrValidation, input)
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
