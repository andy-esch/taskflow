package domain

import (
	"fmt"
	"sort"
)

// This file backs `tskflwctl schema` — the tool's self-description for agents.
// Everything here is DERIVED from the same data the rest of the tool runs on
// (the type maps, the status/bucket enums), so the schema output cannot drift
// from real behavior. The only hand-authored content is per-field prose
// (description/example), which a sync test pins to the real field set.

// FieldType reports the YAML storage type of a known task field: "int", "list",
// "date", or "string". It reads the same maps the `--set` coercion and the
// store's diagnose/fix paths use, so the schema can't claim a type the writer
// won't produce.
func FieldType(name string) string {
	switch {
	case IntFields[name]:
		return "int"
	case ListFields[name]:
		return "list"
	case dateFields[name]:
		return "date"
	default:
		return "string"
	}
}

// KnownTaskFieldNames returns every frontmatter key the tool knows for a task,
// sorted for a stable schema dump.
func KnownTaskFieldNames() []string {
	names := make([]string, 0, len(knownTaskFields))
	for f := range knownTaskFields {
		names = append(names, f)
	}
	sort.Strings(names)
	return names
}

// FieldDoc documents one *authoring* field: the human/agent-set frontmatter a
// well-formed document declares. Type is the YAML storage type; for task fields
// a sync test pins it to FieldType so it can't drift.
type FieldDoc struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// The authoring field sets are the frontmatter a drafter fills in — not the
// tool-managed stamps (created/*_at/audited) or status (which is the directory).
var (
	taskAuthoringFields = []FieldDoc{
		{"epic", "string", true, "ID of the epic this task belongs to; must already exist.", "17-pm-go-cli"},
		{"description", "string", false, "One line summarizing the task (≤150 chars); required once next-up/in-progress.", "Add retry backoff to the Strava webhook"},
		{"effort", "string", false, "Rough size estimate (free-form).", "1-2 hours"},
		{"tier", "int", false, "Importance, 1 (highest) – 5 (lowest).", "2"},
		{"priority", "string", false, "One of: high | medium | low.", "medium"},
		{"autonomy_level", "int", false, "How autonomously this can be done, 1–5.", "3"},
		{"tags", "list", true, "At least one topical tag (required at creation).", "[cli, core]"},
	}
	epicAuthoringFields = []FieldDoc{
		{"status", "string", true, "One of: planning | in-progress | completed | archived.", "planning"},
		{"description", "string", true, "One-line goal (≤150 chars); required.", "Replace the legacy ingest pipeline"},
		{"priority", "string", false, "One of: high | medium | low.", "medium"},
		{"tags", "list", false, "Topical tags.", "[infra]"},
	}
	auditAuthoringFields = []FieldDoc{
		{"area", "string", true, "Subsystem or topic audited; slugified into the filename.", "dispatcher"},
		{"date", "date", true, "Audit date, YYYY-MM-DD (defaults to today).", "2026-06-16"},
	}
)

// SchemaKinds are the document kinds `schema <kind>` describes.
func SchemaKinds() []string { return []string{"task", "epic", "audit"} }

// AuthoringFields returns the documented authoring frontmatter for a kind.
func AuthoringFields(kind string) ([]FieldDoc, error) {
	switch kind {
	case "task":
		return taskAuthoringFields, nil
	case "epic":
		return epicAuthoringFields, nil
	case "audit":
		return auditAuthoringFields, nil
	}
	return nil, fmt.Errorf("%w: unknown kind %q (task|epic|audit)", ErrValidation, kind)
}

// Conventions returns the short, factual authoring rules for a kind — the
// "how to write it" prose that complements the per-field docs.
func Conventions(kind string) []string {
	switch kind {
	case "task":
		return []string{
			"status is the directory — set it with the lifecycle verbs (start/promote/complete/…), never in frontmatter.",
			"description is a single line, ≤150 characters.",
			"at least one tag is required at creation.",
			"the slug is derived from the title; keep titles filename-safe.",
		}
	case "epic":
		return []string{
			"epics are auto-numbered NN-<slug>; do not set the number yourself.",
			"description is required (single line, ≤150 chars).",
		}
	case "audit":
		return []string{
			"audits are created in the open bucket; move them with audit close/reopen/defer.",
			"the slug is <date>-<area>; findings live in the body as `#### H1. … **Status:** open`.",
		}
	}
	return nil
}
