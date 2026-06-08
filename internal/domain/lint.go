package domain

import (
	"fmt"
	"strings"
)

// Issue is a single frontmatter lint finding.
type Issue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// LintTask returns the frontmatter issues for an (active) task. validEpic
// reports whether an epic id exists; pass nil to skip the epic-existence check.
func LintTask(t Task, validEpic func(string) bool) []Issue {
	var issues []Issue
	add := func(field, msg string) { issues = append(issues, Issue{Field: field, Message: msg}) }

	if t.Status == "" {
		add("status", "missing")
	}
	switch {
	case t.Epic == "":
		add("epic", "missing")
	case validEpic != nil && !validEpic(t.Epic):
		add("epic", fmt.Sprintf("unknown epic %q", t.Epic))
	}
	switch {
	case t.Tier == 0:
		add("tier", "missing")
	case t.Tier < 1 || t.Tier > 5:
		add("tier", "must be 1-5")
	}
	switch {
	case t.Priority == "":
		add("priority", "missing")
	case !validPriorities[t.Priority]:
		add("priority", "must be high|medium|low")
	}
	if t.Autonomy != 0 && (t.Autonomy < 1 || t.Autonomy > 5) {
		add("autonomy_level", "must be 1-5")
	}
	if t.Effort == "" {
		add("effort", "missing")
	}
	if t.Created == "" {
		add("created", "missing")
	}
	if len(t.Tags) == 0 {
		add("tags", "missing")
	}

	switch {
	case t.Description == "":
		if t.Status == StatusNextUp || t.Status == StatusInProgress {
			add("description", "required for next-up/in-progress")
		}
	case strings.ContainsAny(t.Description, "\r\n"):
		add("description", "must be a single line")
	case len(t.Description) > MaxDescriptionLen:
		add("description", fmt.Sprintf("too long (%d > %d)", len(t.Description), MaxDescriptionLen))
	}

	return issues
}
