package domain

import (
	"fmt"
	"sort"
	"strings"
)

// Epic is a long-lived domain goal. Tasks reference it by ID via their
// `epic:` field. Epics live flat in epics/<id>.md (no status directories;
// status is a frontmatter field with its own — closed — vocabulary, distinct
// from the task lifecycle).
type Epic struct {
	ID   string `yaml:"-"`
	Path string `yaml:"-"`

	Status      string   `yaml:"status"`
	Description string   `yaml:"description"`
	Priority    string   `yaml:"priority"`
	Created     string   `yaml:"created"`
	Tags        []string `yaml:"tags"`
}

// Epic status values — the closed, intent-level vocabulary (decided 2026-06-25).
// An epic is a long-lived domain category, not a task that marches to "done", so
// the states answer "is this bucket live, finished, or dead", not "what stage".
// Crucially they do NOT track how busy the bucket is: an `active` epic may be
// working (open tasks) or dormant (drained but still a valid domain). That
// working/dormant distinction is a DERIVED signal computed from the task rollup
// (see core.EpicSummary.Liveness), deliberately not a stored status, so a quiet
// domain never needs hand-maintenance to stay correct.
const (
	// EpicStatusActive is a live domain bucket — organizing current or future work.
	// The dashboard/epics-tab default view; refined at read time into working vs
	// dormant liveness. Most epics live here for their whole life.
	EpicStatusActive = "active"
	// EpicStatusRetired means the goals were satisfied and the epic closed
	// successfully — kept for history, not expected to take new work.
	EpicStatusRetired = "retired"
	// EpicStatusDeprecated means the epic wasn't useful or was replaced; a
	// superseded epic is deprecated with a "superseded by X" note in its
	// description rather than getting a separate state.
	EpicStatusDeprecated = "deprecated"
)

// epicStatuses is the closed epic-status vocabulary, in declared order.
var epicStatuses = []string{EpicStatusActive, EpicStatusRetired, EpicStatusDeprecated}

// AllEpicStatuses returns the closed epic-status vocabulary, in declared order.
func AllEpicStatuses() []string { return epicStatuses }

// ValidateEpicStatus rejects an epic status outside the closed vocabulary,
// wrapping ErrValidation and enumerating the valid values.
func ValidateEpicStatus(s string) error {
	for _, v := range epicStatuses {
		if s == v {
			return nil
		}
	}
	return fmt.Errorf("%w: invalid epic status %q (valid: %s)",
		ErrValidation, s, strings.Join(epicStatuses, ", "))
}

// IsKnownEpicStatus reports whether s is one of the canonical epic statuses. It's
// the boolean form of ValidateEpicStatus (no error allocation), used by the read
// surfaces to flag a non-conforming epic — e.g. one ported from a repo with a
// different lifecycle vocabulary (planning/in-progress/completed) — without
// rejecting it: the tool still lists, rolls up, and lets you fix it.
func IsKnownEpicStatus(s string) bool { return ValidateEpicStatus(s) == nil }

// IsEpicArchived reports whether an epic is in a closed/terminal state — retired
// (finished) or deprecated (abandoned/replaced). This is the ONLY status check the
// dashboard and the epics-tab default view hide on: visibility fails OPEN, so an
// unknown/foreign/missing status reads as live (and gets flagged, not dropped)
// rather than vanishing. Hiding the known terminals keeps a conforming repo's
// behavior identical while a non-conforming one stays usable.
func IsEpicArchived(s string) bool {
	return s == EpicStatusRetired || s == EpicStatusDeprecated
}

// knownEpicFields is every frontmatter key tskflwctl itself reads or writes on an
// epic — the epic analog of knownTaskFields. Mirrors the domain.Epic yaml tags;
// `epic set --set` rejects keys outside it unless forced (a typo'd field name must
// not silently persist). `tags` is the only epic list field; epics have no int or
// date-stamped fields, so no IsIntField/dateFields entry is needed here.
var knownEpicFields = map[string]bool{
	"status": true, "description": true, "priority": true,
	"created": true, "tags": true,
}

// KnownEpicField reports whether a frontmatter key is one the tool knows for an
// epic. `epic set --set` rejects unknown keys unless forced (mirrors
// KnownTaskField for tasks).
func KnownEpicField(f string) bool { return knownEpicFields[f] }

// KnownEpicFieldNames returns every frontmatter key the tool knows for an epic,
// sorted for a stable schema dump (the epic analog of KnownTaskFieldNames).
func KnownEpicFieldNames() []string {
	names := make([]string, 0, len(knownEpicFields))
	for f := range knownEpicFields {
		names = append(names, f)
	}
	sort.Strings(names)
	return names
}

// IsEpicListField reports whether an epic frontmatter key is stored as a YAML
// list (only `tags`), so the `--set key=value` coercion writes a sequence rather
// than a corrupting !!str. The epic counterpart to IsListField.
func IsEpicListField(f string) bool { return f == "tags" }

// ValidateEpicField checks a constrained epic frontmatter field from its string
// value — the `epic set` path, where every value arrives as a string. The epic
// analog of ValidateField: status moves via `epic move` (rejected here),
// priority/description are validated, everything else passes.
func ValidateEpicField(field, value string) error {
	switch field {
	case "status":
		return fmt.Errorf("%w: set epic status with `epic move`, not `set`", ErrValidation)
	case "priority":
		return ValidatePriority(value)
	case "description":
		return ValidateDescription(value)
	case "created":
		return ValidateDate(value)
	}
	return nil
}
