package domain

import "sort"

// This file backs `tskflwctl schema` — the tool's self-description for agents.
// Field/type inventories come from the same registries the tool runs on.
// Descriptions, examples, and authoring conventions are hand-authored metadata
// in entity.go and need review alongside behavior changes. Sync tests check
// field membership/types; they do not prove every prose claim.

// FieldType reports the YAML storage type of a known task field: "int", "list",
// "date", or "string". It reads the same maps the `--set` coercion and the
// store's diagnose/fix paths use, so the schema can't claim a type the writer
// won't produce.
func FieldType(name string) string {
	switch {
	case IsIntField(name):
		return "int"
	case IsListField(name):
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
// a sync test pins it to FieldType so it can't drift. The per-entity field sets,
// conventions, and the SchemaKinds/AuthoringFields/Conventions lookups now live on
// the entity registry in entity.go.
type FieldDoc struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Example     string `json:"example"`
}
