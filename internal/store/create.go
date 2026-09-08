package store

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
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

// entityFileCreation is one fully prepared ordinary entity file. Any repository
// scan or identifier allocation needed to prepare it belongs in the callback
// passed to createEntityFile, where the store can serialize that work with the
// no-clobber write.
type entityFileCreation struct {
	dir     string
	path    string
	content []byte
	kind    string
	name    string
}

// createEntityFile owns the check-and-create transaction for ordinary single-file
// entities. On a real write, prepare runs while the repository guard is held and
// the file is then created without releasing that guard. This ordering prevents a
// future entity from accidentally recreating the audit/research race by scanning
// identity first and acquiring the write lock later.
//
// Dry-run invokes the same preparation and exact-path collision checks without
// creating the planning root or taking a writer lock. Compound graph-aware creates
// retain their richer guarded planners and finish through writeNewFileUnlocked.
func (s *FS) createEntityFile(dryRun bool, prepare func() (entityFileCreation, error)) (entityFileCreation, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return entityFileCreation{}, err
	}
	if prepare == nil {
		return entityFileCreation{}, fmt.Errorf("%w: entity file preparation is required", domain.ErrValidation)
	}
	if dryRun {
		creation, err := prepare()
		if err != nil {
			return entityFileCreation{}, err
		}
		if _, err := os.Stat(creation.path); err == nil {
			return entityFileCreation{}, entityAlreadyExistsError(creation.kind, creation.name)
		} else if !os.IsNotExist(err) {
			return entityFileCreation{}, fmt.Errorf("stat %s %s: %w", creation.kind, creation.path, err)
		}
		return creation, nil
	}

	// Preserve the store's historical ability to create the first entity in a
	// not-yet-existing root; the directory-backed Unix lock needs the root first.
	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return entityFileCreation{}, fmt.Errorf("mkdir planning root %s: %w", s.root, err)
	}
	unlock, err := s.writeLock()
	if err != nil {
		return entityFileCreation{}, err
	}
	defer unlock()

	creation, err := prepare()
	if err != nil {
		return entityFileCreation{}, err
	}
	if err := s.writeNewFileUnlocked(creation.dir, creation.path, creation.content, creation.kind, creation.name); err != nil {
		return entityFileCreation{}, err
	}
	return creation, nil
}

func entityAlreadyExistsError(kind, name string) error {
	return fmt.Errorf("%s %q already exists: %w", kind, name, domain.ErrConflict)
}

// writeNewFileUnlocked is the repository-guard-compatible create primitive.
// Ordinary entity creation enters through createEntityFile and takes the lock;
// guarded compound operations may call this helper only while already holding it.
func (s *FS) writeNewFileUnlocked(dir, path string, content []byte, kind, id string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	if err := createFileAtomic(path, content, 0o644); err != nil {
		if os.IsExist(err) {
			return entityAlreadyExistsError(kind, id)
		}
		return err
	}
	return nil
}

// taskFields is the canonical frontmatter order for a new task. started_at is
// appended only when set (a `new --start` task) — the one lifecycle stamp a
// create can carry; guarded lifecycle mutations write the others.
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
	if len(t.DependsOn) > 0 {
		dependencies := append([]string(nil), t.DependsOn...)
		sort.Strings(dependencies)
		fields = append(fields, fmField{"depends_on", dependencies})
	}
	if t.StartedAt != "" {
		fields = append(fields, fmField{"started_at", t.StartedAt})
	}
	return fields
}

// CreateTask writes a new task file at tasks/<id>-<slug>.md (flat, id-led per
// ADR-0003 §4). It refuses to clobber an existing file; the slug, id, and status
// are taken from t.
// validEntityID rejects an id that cannot legally appear in a flat filename. The creates
// build `<id>-<slug>.md` directly from it, so an illegal id would produce a file the
// scanner refuses to parse — caught here rather than discovered by the next `lint`.
func validEntityID(entityID string) error {
	if id.Valid(entityID) {
		return nil
	}
	if c, pos, bad := id.InvalidChar(entityID); bad {
		return fmt.Errorf("%w: id %q contains %q at position %d, which Crockford base32 excludes (no i, l, o, u)",
			domain.ErrValidation, entityID, string(c), pos)
	}
	return fmt.Errorf("%w: id %q is not a valid id (%d lowercase Crockford base32 characters)",
		domain.ErrValidation, entityID, id.Length)
}

func (s *FS) CreateTask(t domain.Task, body string, dryRun bool) (domain.Task, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return domain.Task{}, err
	}
	if t.Slug == "" {
		return domain.Task{}, fmt.Errorf("%w: empty task slug", domain.ErrValidation)
	}
	if t.ID == "" {
		return domain.Task{}, fmt.Errorf("%w: task has no id", domain.ErrValidation)
	}
	if err := validEntityID(t.ID); err != nil {
		return domain.Task{}, err
	}
	switch t.Status {
	case domain.StatusReadyToStart, domain.StatusNextUp:
		// Ordinary creation may place a task only in a non-start lifecycle state.
		// In-progress creation belongs exclusively to guarded create-and-start.
	default:
		return domain.Task{}, fmt.Errorf("%w: ordinary task creation supports only ready-to-start or next-up status, got %q", domain.ErrValidation, t.Status)
	}
	if len(t.DependsOn) > 0 || len(t.LegacyBlockedBy) > 0 || len(t.LegacyDependencies) > 0 || len(t.LegacyBlocks) > 0 {
		return domain.Task{}, fmt.Errorf("%w: task creation cannot set graph-owned dependency fields until guarded dependency creation is available", domain.ErrValidation)
	}
	// A duplicate slug with a distinct id is legal and remains resolvable by id.
	// A duplicate stable id with a different slug has a different path, so its
	// identity scan must be serialized with the O_EXCL write.
	stem := t.ID + "-" + t.Slug
	path := filepath.Join(s.tasksDir, stem+".md")
	content, err := buildFile(taskFields(t), body)
	if err != nil {
		return domain.Task{}, err
	}
	creation, err := s.createEntityFile(dryRun, func() (entityFileCreation, error) {
		if err := ensureCandidateIDUnique("task", t.ID, s.taskCandidates); err != nil {
			return entityFileCreation{}, err
		}
		if err := s.ensureTaskIDNotThread(t.ID); err != nil {
			return entityFileCreation{}, err
		}
		return entityFileCreation{dir: s.tasksDir, path: path, content: content, kind: "task", name: stem}, nil
	})
	if err != nil {
		return domain.Task{}, err
	}
	t.Path = creation.path
	return t, nil
}

// ensureCandidateIDUnique applies the same same-kind stable-identity rule to
// every flat entity directory. Filename candidates are deliberately used so an
// unreadable record still retains ownership of its recoverable id.
func ensureCandidateIDUnique(kind, entityID string, list func() ([]candidate, error)) error {
	owner, err := candidateIDOwner(entityID, list)
	if err != nil {
		return err
	}
	if owner != "" {
		return fmt.Errorf("%s id %q already used by %q: %w",
			kind, entityID, owner, domain.ErrConflict)
	}
	return nil
}

// candidateIDOwner returns the basename of the first flat entity whose filename
// owns entityID. Keeping this lower-level query separate lets repair paths report
// a safe refusal through their own result channel instead of turning an expected
// identity collision into a batch-level error.
func candidateIDOwner(entityID string, list func() ([]candidate, error)) (string, error) {
	candidates, err := list()
	if err != nil {
		return "", err
	}
	for _, candidate := range candidates {
		if candidate.id == entityID {
			return filepath.Base(candidate.path), nil
		}
	}
	return "", nil
}

func (s *FS) ensureTaskIDNotThread(taskID string) error {
	candidates, err := s.threadCandidates()
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		if candidate.id == taskID {
			return fmt.Errorf("stable id %s is already used by Thread %s: %w", taskID, filepath.Base(candidate.path), domain.ErrConflict)
		}
	}
	return nil
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

// CreateAudit writes a new audit at audits/<id>-<slug>.md (flat, id-led per
// ADR-0003 §4). New audits always start in the open bucket; it refuses to clobber
// either the exact path or a different audit already using the same stable id.
func (s *FS) CreateAudit(a domain.Audit, body string, dryRun bool) (domain.Audit, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return domain.Audit{}, err
	}
	if a.Slug == "" {
		return domain.Audit{}, fmt.Errorf("%w: empty audit slug", domain.ErrValidation)
	}
	if a.ID == "" {
		return domain.Audit{}, fmt.Errorf("%w: audit has no id", domain.ErrValidation)
	}
	if err := validEntityID(a.ID); err != nil {
		return domain.Audit{}, err
	}
	// A duplicate slug with a distinct id is legal and remains resolvable by id. A
	// duplicate ID with a different slug is a different path, however, so O_EXCL
	// cannot see it and both records would become ambiguous and unwritable.
	a.Bucket = domain.AuditOpen
	stem := a.ID + "-" + a.Slug
	path := filepath.Join(s.auditsDir, stem+".md")
	content, err := buildFile(auditFields(a), body)
	if err != nil {
		return domain.Audit{}, err
	}
	creation, err := s.createEntityFile(dryRun, func() (entityFileCreation, error) {
		if err := s.ensureAuditIDUnique(a.ID); err != nil {
			return entityFileCreation{}, err
		}
		return entityFileCreation{dir: s.auditsDir, path: path, content: content, kind: "audit", name: stem}, nil
	})
	if err != nil {
		return domain.Audit{}, err
	}
	a.Path = creation.path
	return a, nil
}

func (s *FS) ensureAuditIDUnique(auditID string) error {
	return ensureCandidateIDUnique("audit", auditID, s.auditCandidates)
}

// researchFields is the canonical frontmatter order for a new research doc. Thin by
// design: no status/bucket (research has no lifecycle) and no cross-references. tags
// and description are written even when empty so the keys are visible to fill in.
func researchFields(r domain.Research) []fmField {
	return []fmField{
		{"schema", domain.FileSchemaVersion},
		{"id", r.ID},
		{"created", r.Created},
		{"description", r.Description},
		{"tags", r.Tags},
	}
}

// CreateResearch writes a new research doc at research/<id>-<slug>.md (flat, id-led
// per ADR-0003 §4). It refuses to clobber; the slug and id are taken from r. The id is
// minted from r.Created by the caller (core), so ids stay chronological.
func (s *FS) CreateResearch(r domain.Research, body string, dryRun bool) (domain.Research, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return domain.Research{}, err
	}
	if r.Slug == "" {
		return domain.Research{}, fmt.Errorf("%w: empty research slug", domain.ErrValidation)
	}
	if r.ID == "" {
		return domain.Research{}, fmt.Errorf("%w: research doc has no id", domain.ErrValidation)
	}
	if err := validEntityID(r.ID); err != nil {
		return domain.Research{}, err
	}
	// O_EXCL alone is NOT the whole collision guard here. Research ids are minted from a
	// DAY (ADR-0003 §3), so
	// every doc sharing a `created` date draws from the same random tail — and a duplicate
	// id on a DIFFERENT slug is a different path, which O_EXCL never sees. Two docs sharing
	// an id are unresolvable by id and, worse, both become unwritable: the write paths'
	// CAS re-resolve returns ErrAmbiguous, which surfaces as a retryable conflict forever.
	// So check the id against what is already on disk. Cheap: researchCandidates is a
	// ReadDir + filename split, no parsing. createEntityFile keeps that scan and the
	// eventual write inside one repository critical section.
	stem := r.ID + "-" + r.Slug
	path := filepath.Join(s.researchDir, stem+".md")
	content, err := buildFile(researchFields(r), body)
	if err != nil {
		return domain.Research{}, err
	}
	creation, err := s.createEntityFile(dryRun, func() (entityFileCreation, error) {
		if err := ensureCandidateIDUnique("research", r.ID, s.researchCandidates); err != nil {
			return entityFileCreation{}, err
		}
		return entityFileCreation{dir: s.researchDir, path: path, content: content, kind: "research doc", name: stem}, nil
	})
	if err != nil {
		return domain.Research{}, err
	}
	r.Path = creation.path
	return r, nil
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

// nextEpicNumber returns max(existing NN- prefix)+1, or 1 if none. CreateEpic
// invokes it from createEntityFile's preparation callback, so allocation and the
// corresponding no-clobber write share one repository critical section.
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
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return domain.Epic{}, err
	}
	if slug == "" {
		return domain.Epic{}, fmt.Errorf("%w: empty epic slug", domain.ErrValidation)
	}
	creation, err := s.createEntityFile(dryRun, func() (entityFileCreation, error) {
		num, err := s.nextEpicNumber()
		if err != nil {
			return entityFileCreation{}, err
		}
		id := fmt.Sprintf("%02d-%s", num, slug)
		path := filepath.Join(s.epicsDir, id+".md")
		content, err := buildFile(epicFields(e), body)
		if err != nil {
			return entityFileCreation{}, err
		}
		return entityFileCreation{dir: s.epicsDir, path: path, content: content, kind: "epic", name: id}, nil
	})
	if err != nil {
		return domain.Epic{}, err
	}
	e.ID = creation.name
	e.Path = creation.path
	return e, nil
}
