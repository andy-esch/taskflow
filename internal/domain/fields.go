package domain

// The canonical task-frontmatter field registry. taskFields (below) is the ONE
// place that declares which fields exist and their YAML storage type; the
// known-field set and the per-type maps (int/list/date) are all DERIVED from it,
// so they can no longer drift out of sync — previously knownTaskFields was a
// separate hand-list that had to be kept in lockstep with intFields/listFields/
// dateFields by eye. The core's `--set` coercion, the store's diagnose/fix paths,
// lint, and `schema` all read the accessors below.
//
// Keep taskFields in sync with the domain.Task yaml tags; TestTaskFieldsMatchStruct
// pins that every modelled frontmatter tag (except the store-managed id) is here.

// taskField declares one task frontmatter field and its YAML storage type
// ("string" | "int" | "list" | "date").
type taskField struct {
	Name string
	Type string
}

// taskFields is the single source of truth for task frontmatter fields — every
// field the tool reads or writes, with the type the writer produces. Order is
// cosmetic (all consumers are sets or sort).
var taskFields = []taskField{
	{"status", "string"},
	{"epic", "string"},
	{"description", "string"},
	{"effort", "string"},
	{"priority", "string"},
	{"tier", "int"},
	{"autonomy_level", "int"},
	{"tags", "list"},
	{"related_tasks", "list"},
	{"dependencies", "list"},
	{"blocks", "list"},
	{"blocked_by", "list"},
	{"audit_sources", "list"},
	{"projects", "list"},
	{"created", "date"},
	{"updated_at", "date"},
	{"started_at", "date"},
	{"revisit_at", "date"},
	{"completed_at", "date"},
	{"deprecated_at", "date"},
	{"deferred_at", "date"},
	{"audited", "date"},
}

// taskFieldSet returns the set of task field names whose type == typ; typ == ""
// returns every known field.
func taskFieldSet(typ string) map[string]bool {
	m := make(map[string]bool, len(taskFields))
	for _, f := range taskFields {
		if typ == "" || f.Type == typ {
			m[f.Name] = true
		}
	}
	return m
}

// The derived registries. Deriving them from taskFields means a new field is one
// table row, not four literals kept in lockstep. dateFields' other reader is
// ValidateField (validate.go) and FieldType (schema.go).
var (
	knownTaskFields = taskFieldSet("")
	intFields       = taskFieldSet("int")
	listFields      = taskFieldSet("list")
	dateFields      = taskFieldSet("date")
)

// IsIntField reports whether a frontmatter key is stored as a YAML int, and
// IsListField whether it's stored as a YAML list. They are accessors over the
// unexported registries so no sibling package can mutate the canonical type map
// (which would corrupt coercion/fix/diagnose/schema at once) — same tamper-proof
// pattern as KnownTaskField.
func IsIntField(f string) bool  { return intFields[f] }
func IsListField(f string) bool { return listFields[f] }

// KnownTaskField reports whether a frontmatter key is one the tool knows.
// `task set --set` rejects unknown keys unless forced — a typo'd field name
// must not silently persist (decided 2026-06-12).
func KnownTaskField(f string) bool { return knownTaskFields[f] }

// UnsetField is a sentinel value in a SetFields update map: the key is
// removed from the frontmatter instead of being assigned. It exists so field
// removal flows through the same validated, surgical, atomic write path as
// assignment.
type UnsetField struct{}
