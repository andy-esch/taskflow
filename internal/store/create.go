package store

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
)

// fmField is one frontmatter key/value, written in declared order for new files.
type fmField struct {
	key string
	val any
}

// buildFile serializes ordered frontmatter fields + a markdown body into a
// complete file. Values go through valueNode, so a description containing a
// colon is correctly quoted (the pm non-conformant-YAML trap, avoided at the
// source).
func buildFile(fields []fmField, body string) ([]byte, error) {
	mapping := &yaml.Node{Kind: yaml.MappingNode}
	for _, f := range fields {
		node, err := valueNode(f.val)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", f.key, err)
		}
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: f.key}, node)
	}
	return assembleFile(mapping, []byte(body), "\n") // new files are always LF
}

// writeNewFile is the shared new-file contract for Create{Task,Epic,Audit}: it
// atomically creates path (never clobbering), mapping an existing file to an
// ErrConflict named by kind/id, and creating dir as needed. dryRun runs the same
// collision check but skips the write — so a dry-run that would clash still fails.
func (s *FS) writeNewFile(dir, path string, content []byte, kind, id string, dryRun bool) error {
	conflict := func() error {
		return fmt.Errorf("%s %q already exists: %w", kind, id, domain.ErrConflict)
	}
	if dryRun {
		if _, statErr := os.Stat(path); statErr == nil {
			return conflict()
		}
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	if err := createFileAtomic(path, content, 0o644); err != nil {
		if os.IsExist(err) {
			return conflict()
		}
		return err
	}
	return nil
}

// taskFields is the canonical frontmatter order for a new task. started_at is
// appended only when set (a `new --start` task) — the one lifecycle stamp a
// create can carry; the others are written only by Move.
func taskFields(t domain.Task) []fmField {
	fields := []fmField{
		{"schema", domain.FileSchemaVersion},
		{"id", t.ID},
		{"status", string(t.Status)},
		{"epic", t.Epic},
		{"description", t.Description},
		{"effort", t.Effort},
		{"tier", t.Tier},
		{"priority", t.Priority},
		{"autonomy_level", t.Autonomy},
		{"tags", t.Tags},
		{"created", t.Created},
	}
	if t.StartedAt != "" {
		fields = append(fields, fmField{"started_at", t.StartedAt})
	}
	return fields
}

// CreateTask writes a new task file at tasks/<id>-<slug>.md (flat, id-led per
// ADR-0003 §4). It refuses to clobber an existing file; the slug, id, and status
// are taken from t.
func (s *FS) CreateTask(t domain.Task, body string, dryRun bool) (domain.Task, error) {
	if t.Slug == "" {
		return domain.Task{}, fmt.Errorf("%w: empty task slug", domain.ErrValidation)
	}
	if t.ID == "" {
		return domain.Task{}, fmt.Errorf("%w: task has no id", domain.ErrValidation)
	}
	// The id makes the flat filename unique, so writeNewFile's O_EXCL is the whole
	// collision guard — no cross-dir slug scan. A duplicate slug (distinct id) is
	// allowed under the flat layout and stays resolvable by id.
	stem := t.ID + "-" + t.Slug
	path := filepath.Join(s.tasksDir, stem+".md")
	content, err := buildFile(taskFields(t), body)
	if err != nil {
		return domain.Task{}, err
	}
	if err := s.writeNewFile(s.tasksDir, path, content, "task", stem, dryRun); err != nil {
		return domain.Task{}, err
	}
	t.Path = path
	return t, nil
}

// auditFields is the canonical frontmatter order for a new audit.
func auditFields(a domain.Audit) []fmField {
	return []fmField{
		{"schema", domain.FileSchemaVersion},
		{"id", a.ID},
		{"bucket", string(a.Bucket)},
		{"area", a.Area},
		{"date", a.Date},
	}
}

// CreateAudit writes a new audit file under audits/open/<slug>.md. New audits
// always start in the open bucket. It refuses to clobber an existing file.
func (s *FS) CreateAudit(a domain.Audit, body string, dryRun bool) (domain.Audit, error) {
	if a.Slug == "" {
		return domain.Audit{}, fmt.Errorf("%w: empty audit slug", domain.ErrValidation)
	}
	// Reject a slug already in ANY bucket, not just open/ — a slug in closed/ or
	// deferred/ would otherwise resolve ambiguously after this create.
	cands, err := s.auditCandidates()
	if err != nil {
		return domain.Audit{}, err
	}
	if occ := slugCollision(a.Slug, cands); occ != "" {
		return domain.Audit{}, fmt.Errorf("audit %q already exists in %s/: %w", a.Slug, occ, domain.ErrConflict)
	}
	a.Bucket = domain.AuditOpen
	dir := filepath.Join(s.auditsDir, a.Bucket.Dir())
	path := filepath.Join(dir, a.Slug+".md")
	content, err := buildFile(auditFields(a), body)
	if err != nil {
		return domain.Audit{}, err
	}
	if err := s.writeNewFile(dir, path, content, "audit", a.Slug, dryRun); err != nil {
		return domain.Audit{}, err
	}
	a.Path = path
	return a, nil
}

var epicNumRe = regexp.MustCompile(`^(\d+)-`)

// epicNum parses the leading NN- number from an epic id (0 if absent). Epics are
// ordered by this, not lexically, so `100-x` sorts after `99-x` rather than
// before it (the `%02d` pad only delays, never fixes, a string compare).
func epicNum(id string) int {
	if m := epicNumRe.FindStringSubmatch(id); m != nil {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	return 0
}

// nextEpicNumber returns max(existing NN- prefix)+1, or 1 if none.
//
// Not serialized against a concurrent CreateEpic: two `epic new` processes
// racing between this scan and their writes could mint the same number with
// different slugs (O_EXCL only guards an identical path). That's accepted — this
// is a single-user local CLI with no daemon, so concurrent creation doesn't
// occur in practice, and the numeric ordering above keeps even a hand-created
// duplicate deterministic rather than flipping on string compare.
func (s *FS) nextEpicNumber() (int, error) {
	entries, err := os.ReadDir(s.epicsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, fmt.Errorf("read epics dir: %w", err)
	}
	next := 1
	for _, e := range entries {
		m := epicNumRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		if n, _ := strconv.Atoi(m[1]); n+1 > next {
			next = n + 1
		}
	}
	return next, nil
}

// epicFields is the canonical frontmatter order for a new epic.
func epicFields(e domain.Epic) []fmField {
	return []fmField{
		{"schema", domain.FileSchemaVersion},
		{"status", e.Status},
		{"description", e.Description},
		{"priority", e.Priority},
		{"tags", e.Tags},
		{"created", e.Created},
	}
}

// CreateEpic writes a new epic at epics/NN-<slug>.md, auto-assigning the next
// number. It refuses to clobber an existing file. Unlike tasks/audits it needs
// no cross-bucket slug check: the auto-numbered id is always fresh, so an exact
// id collision can't occur. Duplicate *name*-slugs (01-billing + 02-billing) are
// deliberately allowed — they stay distinct ids; only `epic show billing` goes
// fuzzy-ambiguous, recoverable by using the full NN-slug.
func (s *FS) CreateEpic(slug string, e domain.Epic, body string, dryRun bool) (domain.Epic, error) {
	if slug == "" {
		return domain.Epic{}, fmt.Errorf("%w: empty epic slug", domain.ErrValidation)
	}
	num, err := s.nextEpicNumber()
	if err != nil {
		return domain.Epic{}, err
	}
	id := fmt.Sprintf("%02d-%s", num, slug)
	path := filepath.Join(s.epicsDir, id+".md")
	content, err := buildFile(epicFields(e), body)
	if err != nil {
		return domain.Epic{}, err
	}
	// The auto-numbered id is always fresh, so the collision check can't actually
	// fire here — but routing through writeNewFile keeps one create contract.
	if err := s.writeNewFile(s.epicsDir, path, content, "epic", id, dryRun); err != nil {
		return domain.Epic{}, err
	}
	e.ID = id
	e.Path = path
	return e, nil
}
