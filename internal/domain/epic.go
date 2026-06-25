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

// epicStatuses is the closed epic-status vocabulary (decided 2026-06-25). An
// epic is a long-lived domain category, not a task that marches to "done", so
// the states answer "is this bucket live, finished, or dead", not "what stage":
// active = live (organizing current or future work); retired = goals satisfied,
// closed successfully, kept for history; deprecated = it wasn't useful or was
// replaced (a superseded epic is deprecated with a "superseded by X" note in
// its description, not a separate state).
var epicStatuses = []string{"active", "retired", "deprecated"}

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
