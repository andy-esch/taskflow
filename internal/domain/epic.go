package domain

import (
	"fmt"
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
