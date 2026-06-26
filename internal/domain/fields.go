package domain

// The canonical task-frontmatter field-type registry. This is the ONE place
// that knows which fields are ints, which are lists, and which exist at all —
// the core's `--set` coercion and the store's diagnose/fix paths all read it,
// so they can't drift apart (previously each kept its own copy, and `task set`
// wrote forms that the project's own `lint --fix` then rewrote).
//
// Keep in sync with the domain.Task yaml tags + the stamped date fields.

// intFields are frontmatter keys stored as YAML ints.
var intFields = map[string]bool{"tier": true, "autonomy_level": true}

// listFields are frontmatter keys stored as YAML lists.
var listFields = map[string]bool{
	"tags": true, "related_tasks": true, "dependencies": true,
	"blocks": true, "blocked_by": true, "audit_sources": true, "projects": true,
}

// IsIntField reports whether a frontmatter key is stored as a YAML int, and
// IsListField whether it's stored as a YAML list. They are accessors over the
// unexported registries so no sibling package can mutate the canonical type map
// (which would corrupt coercion/fix/diagnose/schema at once) — same tamper-proof
// pattern as KnownTaskField.
func IsIntField(f string) bool  { return intFields[f] }
func IsListField(f string) bool { return listFields[f] }

// knownTaskFields is every frontmatter key tskflwctl itself reads or writes.
var knownTaskFields = map[string]bool{
	"status": true, "epic": true, "description": true, "effort": true,
	"tier": true, "priority": true, "autonomy_level": true, "tags": true,
	"created": true, "updated_at": true, "started_at": true, "revisit_at": true,
	"completed_at": true, "deprecated_at": true, "deferred_at": true, "audited": true,
	"related_tasks": true, "dependencies": true, "blocks": true,
	"blocked_by": true, "audit_sources": true, "projects": true,
}

// KnownTaskField reports whether a frontmatter key is one the tool knows.
// `task set --set` rejects unknown keys unless forced — a typo'd field name
// must not silently persist (decided 2026-06-12).
func KnownTaskField(f string) bool { return knownTaskFields[f] }

// UnsetField is a sentinel value in a SetFields update map: the key is
// removed from the frontmatter instead of being assigned. It exists so field
// removal flows through the same validated, surgical, atomic write path as
// assignment.
type UnsetField struct{}
