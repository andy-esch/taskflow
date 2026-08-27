package domain

import (
	"fmt"
	"strings"
)

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
	Kind            string          // the `schema <kind>` word: task | epic | audit | research
	Plural          string          // the plural for user-facing prose; NOT Kind+"s" — "research" is a mass noun
	Dir             string          // top-level planning dir (TasksDir / EpicsDir / AuditsDir / ResearchDir)
	AuthoringFields []FieldDoc      // frontmatter a drafter fills in (not tool-managed stamps)
	Conventions     []string        // short, factual "how to write it" rules
	Templates       []NamedTemplate // body scaffolds offered for this kind; the one named DefaultTemplate is used when --template is omitted
	Placeholders    []Placeholder   // the {{key}} tokens this kind's templates fill (real values at create; preview labels at show)
}

// DefaultTemplate is the body-scaffold name used when --template is omitted; every
// kind's descriptor must offer one (guarded by a test).
const DefaultTemplate = "default"

// NamedTemplate is one body scaffold a kind offers under a name. Body uses {{key}}
// placeholders (NOT Printf %s) drawn from the kind's Descriptor.Placeholders, so an
// author's body may contain a literal '%' and may use any subset of the kind's
// placeholders without an arity contract. Description is a one-liner for listing.
type NamedTemplate struct {
	Name        string
	Description string
	Body        string
}

// Placeholder is a {{Key}} token a kind's templates may fill: the real value at
// create time, or Label in a placeholder preview (`template show` / `schema`).
// Declared per kind on the Descriptor so the renderer is registry-driven — a new
// kind lights up without a per-kind switch.
type Placeholder struct {
	Key   string // the {{Key}} token, e.g. "title"
	Label string // the preview label, e.g. "<title>"
}

// entities is the single registry of document kinds. The ORDER is the
// schema/display order — keep it task, epic, audit (consumers and golden output
// depend on it). A new entity is one entry here (plus, for now, its store scan and
// render funcs — see the Descriptor doc).
var entities = []Descriptor{
	{
		Kind:   "task",
		Plural: "tasks",
		Dir:    TasksDir,
		AuthoringFields: []FieldDoc{
			{"epic", "string", true, "ID of the epic this task belongs to; must already exist.", "17-pm-go-cli"},
			{"description", "string", false, fmt.Sprintf("One line summarizing the task (≤%d chars); required once next-up/in-progress.", MaxDescriptionLen), "Add retry backoff to the Strava webhook"},
			{"effort", "string", false, "Rough size estimate (free-form).", "1-2 hours"},
			{"tier", "int", false, "Importance, 1 (highest) – 5 (lowest).", "2"},
			{"priority", "string", false, "One of: high | medium | low.", "medium"},
			{"autonomy_level", "int", false, "How autonomously this can be done, 1–5.", "3"},
			{"tags", "list", true, "At least one topical tag (required at creation).", "[cli, core]"},
		},
		Conventions: []string{
			"status lives in frontmatter (authoritative) and is changed only via the lifecycle verbs (start/next/complete/…), which edit it in place — don't edit it directly.",
			"depends_on is a sorted set of stable task IDs owned by the repository-global DAG. It is preservation-only until guarded `task depend add/remove` commands land: generic task creation, `task set`, `task edit`, and `lint --fix` cannot add, remove, or reinterpret it.",
			fmt.Sprintf("description is a single line, ≤%d characters.", MaxDescriptionLen),
			"at least one tag is required at creation.",
			"the filename slug is derived from the title; any title is accepted (colons, dashes, arrows, …) and the full title is kept as the body H1.",
			// Derived from the vocabulary itself, never retyped beside it: a documented list
			// that drifts from the code is the failure recorded as finding M3 of the
			// 2026-08-17 finding-status-surface audit, where a reader of a stale table
			// believed the vocabulary was closed at six values.
			fmt.Sprintf("an acceptance criterion is `- [x]` (met) or `- [ ]` (not met). A not-met criterion may say WHY with a trailing `· **%s:** reason` — one of: %s. A reason is required for those, and a checked criterion takes no suffix.",
				CriterionDeferred, strings.Join(CriterionSuffixStates(), " | ")),
			// Named here for the same reason the audit conventions name `audit finding`:
			// a criterion reachable only by hand-editing markdown is one the tool cannot
			// keep well-formed, and this is the guidance an agent reads before writing one.
			"never hand-edit the acceptance criteria — `task ac` owns them: --check/--uncheck " +
				"<n>, --defer/--wontfix/--tracked/--na <n> --reason <why> for a state, and " +
				"--add/--remove/--replace <n> to change which criteria exist.",
		},
		Templates: []NamedTemplate{
			{DefaultTemplate, "Standard task scaffold: objective, acceptance criteria, out-of-scope, related epic.", taskBodyTemplate},
		},
		Placeholders: []Placeholder{{"title", "<title>"}, {"epic", "<epic-id>"}},
	},
	{
		Kind:   "epic",
		Plural: "epics",
		Dir:    EpicsDir,
		AuthoringFields: []FieldDoc{
			{"status", "string", true, fmt.Sprintf("One of: %s.", strings.Join(epicStatuses, " | ")), "active"},
			{"description", "string", true, fmt.Sprintf("One-line goal (≤%d chars); required.", MaxDescriptionLen), "Replace the legacy ingest pipeline"},
			{"priority", "string", false, "One of: high | medium | low.", "medium"},
			{"tags", "list", false, "Topical tags.", "[infra]"},
		},
		Conventions: []string{
			"epics are auto-numbered NN-<slug>; do not set the number yourself.",
			fmt.Sprintf("description is required (single line, ≤%d chars).", MaxDescriptionLen),
		},
		Templates: []NamedTemplate{
			{DefaultTemplate, "Standard epic scaffold: goal, why-it's-its-own-epic, out-of-scope.", epicBodyTemplate},
		},
		Placeholders: []Placeholder{{"title", "<title>"}, {"description", "<description>"}},
	},
	{
		Kind:   "audit",
		Plural: "audits",
		Dir:    AuditsDir,
		AuthoringFields: []FieldDoc{
			{"area", "string", true, "Subsystem or topic audited; slugified into the filename.", "dispatcher"},
			{"date", "date", true, "Audit date, YYYY-MM-DD (defaults to today).", "2026-06-16"},
		},
		Conventions: []string{
			"audits are created in the open bucket; move them with audit close/reopen/defer.",
			"the slug is <date>-<area>; findings live in the body as `#### H1. … **Status:** open`.",
			// DERIVED, not transcribed. Finding M3 of 2026-08-17-finding-status-surface was
			// a hand-kept status table that had silently fallen a word behind the code, and
			// readers reasoned from the table. Building the sentence from FindingStatuses()
			// makes it impossible for this guidance to lag the vocabulary again.
			"finding statuses: " + strings.Join(FindingStatuses(), " | ") +
				" — `tracked` means handed to a task and needs the destination (`tracked by <id>`).",
			"never hand-edit a finding's **Status:** or **Resolution:** — `audit finding <audit> " +
				"<code> --status <v> [--pr N] [--note <text>]` writes both in one validated, atomic edit.",
		},
		Templates: []NamedTemplate{
			{DefaultTemplate, "Standard audit scaffold: findings + candidate tasks.", auditBodyTemplate},
			{"security", "Security review: threat model, checklist, severity-tagged findings.", auditSecurityBodyTemplate},
		},
		Placeholders: []Placeholder{{"area", "<area>"}, {"date", "<date>"}},
	},
	{
		Kind:   "research",
		Plural: "research docs",
		Dir:    ResearchDir,
		AuthoringFields: []FieldDoc{
			{"created", "date", true, fmt.Sprintf("Date the research was done, YYYY-MM-DD (defaults to today); the id is minted from it, so ids sort chronologically. Must be %s–%s, the range an id can encode.", MintableDateMin, MintableDateMax), "2026-08-14"},
			{"description", "string", false, fmt.Sprintf("One line summarizing what was explored (≤%d chars).", MaxDescriptionLen), "Compared three theming libraries for the TUI"},
			{"tags", "list", false, "Topical tags.", "[tui, color]"},
		},
		Conventions: []string{
			"research has NO status and NO lifecycle verbs — a later doc supersedes an earlier one. A decision that needs a lifecycle is an ADR, not research.",
			"research carries no `epic:`/`tasks:` fields — link to entities from the body with ordinary relative-path links instead.",
			"created is required; the id is minted from it (ADR-0003 §3), so lexical id order is authorship order — and it is therefore NOT settable via `research set`: re-dating a doc means creating a new one.",
			"only description and tags are writable; created, id, schema, and updated_at are tool-managed (`research set` refuses them, each with its reason).",
			"the filename slug is derived from the title; the full title is kept as the body H1.",
		},
		Templates: []NamedTemplate{
			{DefaultTemplate, "Standard research scaffold: question, findings, recommendation.", researchBodyTemplate},
		},
		Placeholders: []Placeholder{{"title", "<title>"}, {"date", "<date>"}},
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
	return nil, fmt.Errorf("%w: unknown kind %q (%s)", ErrValidation, kind, strings.Join(SchemaKinds(), "|"))
}

// Conventions returns the short, factual authoring rules for a kind — the
// "how to write it" prose that complements the per-field docs (nil if unknown).
func Conventions(kind string) []string {
	if d, ok := descriptorFor(kind); ok {
		return d.Conventions
	}
	return nil
}

// LookupTemplate returns a kind's named template (an empty name selects the
// default). An unknown kind or template name returns ErrValidation naming what's
// available, so the CLI maps it to exit 11. It is the metadata-bearing form behind
// `template show`; Template returns just the Body.
func LookupTemplate(kind, name string) (NamedTemplate, error) {
	d, ok := descriptorFor(kind)
	if !ok {
		return NamedTemplate{}, fmt.Errorf("%w: unknown kind %q (%s)", ErrValidation, kind, strings.Join(SchemaKinds(), "|"))
	}
	if name == "" {
		name = DefaultTemplate
	}
	for _, t := range d.Templates {
		if t.Name == name {
			return t, nil
		}
	}
	return NamedTemplate{}, fmt.Errorf("%w: unknown %s template %q (available: %s)",
		ErrValidation, kind, name, strings.Join(templateNames(d), ", "))
}

// Template returns a kind's named body scaffold (raw, with {{key}} placeholders
// unfilled). An empty name selects the default. The create paths fill the
// placeholders with real values and `template show`/`schema` with preview labels;
// see Placeholders.
func Template(kind, name string) (string, error) {
	t, err := LookupTemplate(kind, name)
	return t.Body, err
}

// Placeholders returns the {{key}} tokens (with preview labels) a kind's templates
// fill — registry-driven, so a new kind needs no renderer switch. Nil for an
// unknown kind.
func Placeholders(kind string) []Placeholder {
	if d, ok := descriptorFor(kind); ok {
		return append([]Placeholder(nil), d.Placeholders...)
	}
	return nil
}

// TemplatesFor returns the templates a kind offers (read-only copy, default first),
// or ErrValidation for an unknown kind — the listable form behind `template list`.
func TemplatesFor(kind string) ([]NamedTemplate, error) {
	if d, ok := descriptorFor(kind); ok {
		return append([]NamedTemplate(nil), d.Templates...), nil
	}
	return nil, fmt.Errorf("%w: unknown kind %q (%s)", ErrValidation, kind, strings.Join(SchemaKinds(), "|"))
}

// TemplateNames lists the body-template names a kind offers (default first), for
// completion, listing, and error messages. Nil for an unknown kind.
func TemplateNames(kind string) []string {
	if d, ok := descriptorFor(kind); ok {
		return templateNames(d)
	}
	return nil
}

func templateNames(d Descriptor) []string {
	out := make([]string, len(d.Templates))
	for i, t := range d.Templates {
		out[i] = t.Name
	}
	return out
}

// The built-in body scaffolds. They live in the domain beside the rest of a kind's
// metadata (alongside FieldDoc prose) so a kind's scaffold isn't a separate,
// drift-prone copy in core, and so the named-template set is one registry. Epic 22
// (the selectable template library) layers repo-local templates over these.
const taskBodyTemplate = `
# {{title}}

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [{{epic}}](../epics/{{epic}}.md)
`

const epicBodyTemplate = `
# {{title}}

**Goal.** {{description}}

## Why this is its own epic

<one paragraph: what makes this its own epic vs folding into a sibling?>

## Out of scope

- <explicitly excluded>
`

// auditBodyTemplate's finding example is fenced so a fresh audit counts zero
// findings until real ones are added (parseAudit excludes fenced blocks). It stays
// generic — a repo with its own conventions doc points at it from its own tooling,
// not from the shared tool's scaffold.
const auditBodyTemplate = "\n# Audit: {{area}} — {{date}}\n\n" +
	"> Edit findings in place and flip each `**Status:**` as you work it.\n\n" +
	"## Findings\n\n" +
	"<!-- One finding per issue, in this shape (un-fence it): -->\n\n" +
	"```\n" +
	"#### H1. <title>  · **Status:** open\n\n" +
	"**File:** <path:line> | **Component:** <component>\n" +
	"**Effort:** <XS|S|M|L> · **Urgency:** <acute|soon|eventually>\n\n" +
	"<what's wrong, why it matters, evidence>\n\n" +
	"**Recommendation:** <minimum fix>\n\n" +
	"**Resolution:** <how it was resolved — written by `audit finding --note`, not by hand>\n" +
	"```\n\n" +
	"## Candidate tasks\n\n" +
	"<!-- Mirror each finding: ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do -->\n\n" +
	"- ⏳ `tskflwctl task new \"<title>\" --epic <id> --tags <tag>` — <one line>\n"

// researchBodyTemplate is the research scaffold. It opens with the QUESTION rather
// than a status line: a research doc is an exploration snapshot, so what it was asking
// (and what it concluded) is the durable part. "As of {{date}}" is baked into the
// recommendation heading to keep the snapshot framing visible in the doc itself —
// nothing here expires or needs re-flipping, because research has no lifecycle.
const researchBodyTemplate = `
# {{title}}

## Question

<what was being decided or explored — one short paragraph>

## Findings

<what the investigation turned up; sources, measurements, trade-offs>

## Recommendation (as of {{date}})

<the shape recommended, and what would change the call>

## Related

- <links to the tasks/epics this informs, or to earlier research it builds on>
`

// auditSecurityBodyTemplate is the `security` audit scaffold: the same finding
// grammar as the default (a fenced example, so a fresh audit counts zero findings)
// plus a threat-model header and a review checklist to anchor a security pass. Uses
// the same {{area}}/{{date}} placeholders as the default audit template.
const auditSecurityBodyTemplate = "\n# Security audit: {{area}} — {{date}}\n\n" +
	"> Security review. Edit findings in place and flip each `**Status:**` as you work it.\n\n" +
	"## Threat model\n\n" +
	"- **Assets / trust boundaries:** <what's worth protecting; where untrusted input crosses in>\n" +
	"- **Attacker & entry points:** <who, and through which surfaces>\n\n" +
	"## Review checklist\n\n" +
	"- [ ] Authn / authz — every privileged path checks identity *and* permission\n" +
	"- [ ] Input validation — untrusted input is parsed/escaped (injection, path traversal)\n" +
	"- [ ] Secrets — no hard-coded creds; least-privilege tokens; nothing sensitive logged\n" +
	"- [ ] Dependencies — known-vuln scan; versions pinned\n" +
	"- [ ] Data at rest / in transit — encryption + safe defaults\n\n" +
	"## Findings\n\n" +
	"<!-- One finding per issue, in this shape (un-fence it): -->\n\n" +
	"```\n" +
	"#### H1. <title>  · **Status:** open\n\n" +
	"**File:** <path:line> | **Component:** <component>\n" +
	"**Severity:** <critical|high|medium|low> · **Effort:** <XS|S|M|L> · **Urgency:** <acute|soon|eventually>\n\n" +
	"<what's exploitable, the impact, and how>\n\n" +
	"**Recommendation:** <the fix>\n" +
	"```\n\n" +
	"## Candidate tasks\n\n" +
	"<!-- Mirror each finding: ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do -->\n\n" +
	"- ⏳ `tskflwctl task new \"<title>\" --epic <id> --tags security` — <one line>\n"

// PluralKind returns the plural of a document kind for user-facing prose (an
// ambiguous-match error listing candidates, say). Suffixing "s" is wrong for a mass
// noun — "research" would become "researchs" — so the plural is declared per kind on
// the Descriptor. An unregistered kind falls back to kind+"s", which is right for the
// regular nouns a future entity is likely to add.
func PluralKind(kind string) string {
	if d, ok := descriptorFor(kind); ok && d.Plural != "" {
		return d.Plural
	}
	return kind + "s"
}
