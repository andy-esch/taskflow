package domain

import "fmt"

// Descriptor is the per-entity metadata the tool would otherwise hand-enumerate
// in a `switch kind` at every layer. One registry entry (entities, below) gives a
// document kind its top-level directory, its authoring frontmatter, and its
// authoring conventions; the schema/authoring/convention lookups all read the
// registry, so adding a new entity (the scaffolded project/adr) is a registry
// entry rather than edits scattered across the domain.
//
// Scope (M1, staged): this collapses the DOMAIN enumeration. Body templates
// (core), the store scan, and the render/TUI delegates are deliberately still
// per-entity — tracked by the rest of epic 21 (M9/M10) and the later M1 steps.
// The descriptor names the directory but is NOT a second source of truth for a
// task's status: the per-status/bucket subdirs stay derived from the status/bucket
// enums via layout.go's TaskStatusDirs/AuditBucketDirs.
type Descriptor struct {
	Kind            string     // the `schema <kind>` word: task | epic | audit
	Dir             string     // top-level planning dir (TasksDir / EpicsDir / AuditsDir)
	AuthoringFields []FieldDoc // frontmatter a drafter fills in (not tool-managed stamps)
	Conventions     []string   // short, factual "how to write it" rules
	BodyTemplate    string     // default body scaffold (a Printf format; the placeholder arity is kind-specific — callers fill it)
}

// entities is the single registry of document kinds. The ORDER is the
// schema/display order — keep it task, epic, audit (consumers and golden output
// depend on it). A new entity is one entry here (plus, for now, its store scan and
// render funcs — see the Descriptor doc).
var entities = []Descriptor{
	{
		Kind: "task",
		Dir:  TasksDir,
		AuthoringFields: []FieldDoc{
			{"epic", "string", true, "ID of the epic this task belongs to; must already exist.", "17-pm-go-cli"},
			{"description", "string", false, "One line summarizing the task (≤150 chars); required once next-up/in-progress.", "Add retry backoff to the Strava webhook"},
			{"effort", "string", false, "Rough size estimate (free-form).", "1-2 hours"},
			{"tier", "int", false, "Importance, 1 (highest) – 5 (lowest).", "2"},
			{"priority", "string", false, "One of: high | medium | low.", "medium"},
			{"autonomy_level", "int", false, "How autonomously this can be done, 1–5.", "3"},
			{"tags", "list", true, "At least one topical tag (required at creation).", "[cli, core]"},
		},
		Conventions: []string{
			"status is the directory — set it with the lifecycle verbs (start/promote/complete/…), never in frontmatter.",
			"description is a single line, ≤150 characters.",
			"at least one tag is required at creation.",
			"the slug is derived from the title; keep titles filename-safe.",
		},
		BodyTemplate: taskBodyTemplate,
	},
	{
		Kind: "epic",
		Dir:  EpicsDir,
		AuthoringFields: []FieldDoc{
			{"status", "string", true, "One of: planning | in-progress | completed | archived.", "planning"},
			{"description", "string", true, "One-line goal (≤150 chars); required.", "Replace the legacy ingest pipeline"},
			{"priority", "string", false, "One of: high | medium | low.", "medium"},
			{"tags", "list", false, "Topical tags.", "[infra]"},
		},
		Conventions: []string{
			"epics are auto-numbered NN-<slug>; do not set the number yourself.",
			"description is required (single line, ≤150 chars).",
		},
		BodyTemplate: epicBodyTemplate,
	},
	{
		Kind: "audit",
		Dir:  AuditsDir,
		AuthoringFields: []FieldDoc{
			{"area", "string", true, "Subsystem or topic audited; slugified into the filename.", "dispatcher"},
			{"date", "date", true, "Audit date, YYYY-MM-DD (defaults to today).", "2026-06-16"},
		},
		Conventions: []string{
			"audits are created in the open bucket; move them with audit close/reopen/defer.",
			"the slug is <date>-<area>; findings live in the body as `#### H1. … **Status:** open`.",
		},
		BodyTemplate: auditBodyTemplate,
	},
}

// Descriptors returns the entity registry (read-only copy) in schema/display
// order, so consumers can iterate the document kinds without re-listing them —
// the store's layout, a future `schema --type cli`, and the template library all
// read one source instead of hardcoding the entity set.
func Descriptors() []Descriptor {
	return append([]Descriptor(nil), entities...)
}

// descriptorFor returns the descriptor for a document kind.
func descriptorFor(kind string) (Descriptor, bool) {
	for _, d := range entities {
		if d.Kind == kind {
			return d, true
		}
	}
	return Descriptor{}, false
}

// SchemaKinds are the document kinds `schema <kind>` describes, in registry order.
func SchemaKinds() []string {
	out := make([]string, len(entities))
	for i, d := range entities {
		out[i] = d.Kind
	}
	return out
}

// AuthoringFields returns the documented authoring frontmatter for a kind.
func AuthoringFields(kind string) ([]FieldDoc, error) {
	if d, ok := descriptorFor(kind); ok {
		return d.AuthoringFields, nil
	}
	return nil, fmt.Errorf("%w: unknown kind %q (task|epic|audit)", ErrValidation, kind)
}

// Conventions returns the short, factual authoring rules for a kind — the
// "how to write it" prose that complements the per-field docs (nil if unknown).
func Conventions(kind string) []string {
	if d, ok := descriptorFor(kind); ok {
		return d.Conventions
	}
	return nil
}

// BodyTemplate returns the default body scaffold for a kind ("" if unknown). It
// is the SAME template the create use-cases write and `schema <kind>` shows, now
// single-sourced on the descriptor. The result is a Printf format; callers fill
// the kind-specific placeholders (task: title, epic-id; epic: title, description;
// audit: area, date).
func BodyTemplate(kind string) string {
	if d, ok := descriptorFor(kind); ok {
		return d.BodyTemplate
	}
	return ""
}

// The default body scaffolds. They live in the domain beside the rest of a kind's
// metadata (alongside FieldDoc prose) so a kind's scaffold isn't a separate,
// drift-prone copy in core. (A future selectable template library would generalize
// this single default into a named set — see epic 21.)
const taskBodyTemplate = `
# %s

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [[%s]]
`

const epicBodyTemplate = `
# %s

**Goal.** %s

## Why this is its own epic

<one paragraph: what makes this its own epic vs folding into a sibling?>

## Out of scope

- <explicitly excluded>
`

// auditBodyTemplate's finding example is fenced so a fresh audit counts zero
// findings until real ones are added (parseAudit excludes fenced blocks). It stays
// generic — a repo with its own conventions doc points at it from its own tooling,
// not from the shared tool's scaffold.
const auditBodyTemplate = "\n# Audit: %s — %s\n\n" +
	"> Edit findings in place and flip each `**Status:**` as you work it.\n\n" +
	"## Findings\n\n" +
	"<!-- One finding per issue, in this shape (un-fence it): -->\n\n" +
	"```\n" +
	"#### H1. <title>  · **Status:** open\n\n" +
	"**File:** <path:line> | **Component:** <component>\n" +
	"**Effort:** <XS|S|M|L> · **Urgency:** <acute|soon|eventually>\n\n" +
	"<what's wrong, why it matters, evidence>\n\n" +
	"**Recommendation:** <minimum fix>\n" +
	"```\n\n" +
	"## Candidate tasks\n\n" +
	"<!-- Mirror each finding: ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do -->\n\n" +
	"- ⏳ `tskflwctl task new \"<title>\" --epic <id> --tags <tag>` — <one line>\n"
