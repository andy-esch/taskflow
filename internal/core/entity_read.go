package core

import (
	"fmt"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// EntityKind identifies a planning record without naming an adapter directory or
// transport. It is shared by ordinary reads, lint, dashboards, and repair
// diagnostics so every primary adapter sees one stable vocabulary.
type EntityKind string

const (
	EntityTask     EntityKind = "task"
	EntityEpic     EntityKind = "epic"
	EntityAudit    EntityKind = "audit"
	EntityResearch EntityKind = "research"
	EntityThread   EntityKind = "thread"
)

// RecordSource is adapter-supplied identity and optional explanatory source
// context for one readable record. ID is the canonical stable identity used by
// the application. Location is opaque and must never be parsed for identity or
// assumed to be a local path. LocationIsPath is adapter-supplied presentation
// evidence that the location is already the separately resolvable local path;
// it is not a path capability or permission to open Location as a file.
type RecordSource struct {
	ID             string
	Location       string
	LocationIsPath bool
}

// LoadedRecord pairs semantic data with the source identity that selected it.
// It is a shared value building block; ports remain entity- and use-case-specific.
type LoadedRecord[T any] struct {
	Value  T
	Source RecordSource
}

// Readable source context is useful when it adds information beyond an already
// available local path; keep location and path conceptually independent.
func readableDiagnosticLocation(source RecordSource) string {
	if source.LocationIsPath {
		return ""
	}
	return source.Location
}

// VersionedRecord keeps guarded snapshot evidence beside, but outside, the
// semantic record. Ordinary projections use Record and never expose the opaque
// version or optional local repair path to renderers or domain values. LocalPath
// is supplied explicitly by a local guarded adapter; Location alone never
// authorizes repair.
type VersionedRecord[T any] struct {
	Record        LoadedRecord[T]
	SourceVersion string `json:"-" yaml:"-"`
	LocalPath     string `json:"-" yaml:"-"`
}

// LoadProblem is one non-fatal failed-record diagnostic. Stable identity is
// explicit when the adapter can recover it. Location is optional opaque context;
// LocalPath is an independently optional repair handle for local adapters.
type LoadProblem struct {
	EntityKind EntityKind
	EntityID   string
	EntitySlug string
	Location   string
	LocalPath  string
	Message    string
}

// Entity-specific snapshots keep use-case ports explicit while sharing the
// immutable record and diagnostic vocabulary.
type TaskRead struct {
	Records  []LoadedRecord[domain.Task]
	Problems []LoadProblem
}

type EpicRead struct {
	Records  []LoadedRecord[domain.Epic]
	Problems []LoadProblem
}

type AuditRead struct {
	Records  []LoadedRecord[domain.Audit]
	Problems []LoadProblem
}

type ResearchRead struct {
	Records  []LoadedRecord[domain.Research]
	Problems []LoadProblem
}

// Body-bearing values keep one selected source read together with its markdown
// body. They do not carry local-path or revision evidence.
type EpicWithBody struct {
	Epic domain.Epic
	Body string
}

type AuditWithBody struct {
	Audit domain.Audit
	Body  string
}

type ResearchWithBody struct {
	Research domain.Research
	Body     string
}

type ThreadWithBody struct {
	Thread domain.Thread
	Body   string
}

func requireSourceID(kind EntityKind, source RecordSource) error {
	if strings.TrimSpace(source.ID) != "" {
		return nil
	}
	return fmt.Errorf("%w: %s record has no canonical source ID", domain.ErrValidation, kind)
}

// loadedRecordsWithIDs keeps a malformed adapter record visible as a portable
// partial-read diagnostic, but never publishes it as an addressable entity.
// Its declared ID is intentionally not copied into EntityID: the adapter did
// not establish a canonical identity for that physical occurrence.
func loadedRecordsWithIDs[T any](kind EntityKind, records []LoadedRecord[T], problems []LoadProblem, slug func(T) string) ([]LoadedRecord[T], []LoadProblem) {
	valid := make([]LoadedRecord[T], 0, len(records))
	problems = append([]LoadProblem(nil), problems...)
	for _, record := range records {
		if requireSourceID(kind, record.Source) == nil {
			valid = append(valid, record)
			continue
		}
		problems = append(problems, LoadProblem{
			EntityKind: kind, EntitySlug: slug(record.Value),
			Location: record.Source.Location,
			Message:  "record has no canonical source ID",
		})
	}
	return valid, problems
}

func boolRank(value bool) int {
	if value {
		return 1
	}
	return 0
}
