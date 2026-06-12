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

// epicStatuses is the closed epic-status vocabulary (decided 2026-06-12 — the
// values already in use, plus "completed").
var epicStatuses = []string{"planning", "in-progress", "completed", "archived"}

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
