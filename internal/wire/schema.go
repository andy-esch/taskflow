package wire

import "github.com/andy-esch/taskflow/internal/domain"

const (
	// ExitCodeStateActive means the running binary can deliberately return the code.
	ExitCodeStateActive = "active"
	// ExitCodeStateReserved preserves a retired number and explicitly says the
	// running binary does not emit it.
	ExitCodeStateReserved = "reserved"
)

// This file holds the schema-contract DTOs — the wire shape of `tskflwctl schema`
// (the tool's self-description for agents) and `schema <kind>` (per-kind authoring
// guidance) — embedded by SchemaEnvelope / SchemaKindEnvelope. They are wire types
// (machine contract), so they live here; the human renderers consume them.

// SchemaStatus is one task status and whether it is part of the working set.
type SchemaStatus struct {
	Value  string `json:"value"`
	Active bool   `json:"active"`
}

// SchemaField is one known frontmatter field and its YAML storage type.
type SchemaField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// SchemaExitCode is one process exit and its stable machine name. Active domain
// and generic failures use the name as `error.code` in --json envelopes. Success
// has no error envelope, interactive abort is human-only, and reserved rows are
// explicitly not emitted.
type SchemaExitCode struct {
	Code    int    `json:"code" jsonschema:"description=process exit code"`
	Name    string `json:"name" jsonschema:"description=stable machine name; envelope-producing failures use it as error.code"`
	State   string `json:"state" jsonschema:"enum=active,enum=reserved,description=whether this binary may emit the code"`
	Meaning string `json:"meaning" jsonschema:"description=stable semantic meaning for callers"`
}

// SchemaRevisionPolicy makes ADR-0008's compatibility rules discoverable to a
// machine consumer without requiring access to source comments or planning docs.
type SchemaRevisionPolicy struct {
	Scheme                 string `json:"scheme"`
	Scope                  string `json:"scope"`
	DefaultCompatibility   string `json:"default_compatibility"`
	CurrentCompatibility   string `json:"current_compatibility"`
	ClassifiedSince        string `json:"classified_since"`
	ReaderExpectation      string `json:"reader_expectation"`
	GeneratedJSONSchemaFor string `json:"generated_json_schema_for"`
	JSONSchemaValidation   string `json:"json_schema_validation"`
}

// CurrentSchemaRevisionPolicy returns the policy for the running binary. Keep
// its constants beside SchemaVersion so one changelog test can guard them.
func CurrentSchemaRevisionPolicy() SchemaRevisionPolicy {
	return SchemaRevisionPolicy{
		Scheme:                 SchemaRevisionScheme,
		Scope:                  SchemaRevisionScope,
		DefaultCompatibility:   SchemaRevisionCompatibilityDefault,
		CurrentCompatibility:   SchemaRevisionCompatibility,
		ClassifiedSince:        SchemaRevisionClassificationSince,
		ReaderExpectation:      SchemaRevisionReaderExpectation,
		GeneratedJSONSchemaFor: JSONSchemaScope,
		JSONSchemaValidation:   JSONSchemaValidationMode,
	}
}

// NormalizeSchemaContract stamps wire-owned metadata onto an assembled schema
// contract. Primary adapters supply the domain registries below, but they must
// not be able to omit or contradict the revision policy of the running binary.
func NormalizeSchemaContract(c SchemaContract) SchemaContract {
	c.RevisionPolicy = CurrentSchemaRevisionPolicy()
	return c
}

// SchemaContract is the global machine contract (`tskflwctl schema`): everything
// an agent needs to drive the tool without parsing --help prose.
type SchemaContract struct {
	RevisionPolicy  SchemaRevisionPolicy `json:"revision_policy"`
	Statuses        []SchemaStatus       `json:"statuses"`
	EpicStatuses    []string             `json:"epic_statuses"`
	ThreadStatuses  []string             `json:"thread_statuses"`
	AuditBuckets    []string             `json:"audit_buckets"`
	FindingStatuses []string             `json:"finding_statuses"`
	// CriterionStates are the non-binary acceptance-criterion states — the words legal as a
	// `· **<state>:** <reason>` suffix beside a checkbox. Published for the same reason
	// finding_statuses is: `CriterionJSON.state` has carried one of these since 1.46, and
	// without the set an agent had to trigger an error and parse prose to learn it.
	CriterionStates []string      `json:"criterion_states"`
	TaskFields      []SchemaField `json:"task_fields"`
	EpicFields      []string      `json:"epic_fields"`
	// ResearchFields mirrors TaskFields for research docs. Added because `research set`
	// gates unknown keys on a field registry, and without this an agent had to trigger an
	// error and parse prose to learn the set — the same reason epic_fields was added when
	// `epic set` landed. Which of these are WRITABLE is narrower: see `schema research`,
	// whose conventions name the protected ones.
	ResearchFields []SchemaField    `json:"research_fields"`
	ExitCodes      []SchemaExitCode `json:"exit_codes"`
	Kinds          []string         `json:"kinds"`
}

// KindSchema is the per-kind authoring guidance (`tskflwctl schema <kind>`): how
// to compose a well-formed document of that kind.
type KindSchema struct {
	Kind         string            `json:"kind"`
	Sections     []string          `json:"sections"`
	BodyTemplate string            `json:"body_template"`
	Fields       []domain.FieldDoc `json:"fields"`
	Conventions  []string          `json:"conventions"`
	Templates    []TemplateInfo    `json:"templates"`
}

// TemplateInfo is one body template's listable metadata (kind/name/description),
// populated by the cli for `template list`/`show`. The rendered body is carried
// separately (TemplateShowEnvelope), since the list view never needs it.
type TemplateInfo struct {
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
