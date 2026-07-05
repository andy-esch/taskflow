package domain

// Planning-tree directory layout. The filesystem store owns absolute paths, but the
// directory *names* are domain knowledge — shared by the store (`NewFS`/`WatchPaths`),
// `config.Init`, and shell completion. Kept in one place. Under the flat layout
// (ADR-0003 §4) entities live directly in these dirs; there are no status/bucket subdirs.
const (
	TasksDir    = "tasks"
	EpicsDir    = "epics"
	AuditsDir   = "audits"
	ProjectsDir = "projects"
)

// FileSchemaVersion is the on-disk frontmatter format version, stamped as the
// `schema:` key into new task/epic/audit scaffolds (Create{Task,Epic,Audit}).
// It is DISTINCT from the --json output contract version (render.SchemaVersion):
// this versions the *files on disk*, so a future format migration has an in-file
// signal to branch on. The loader ignores it (it is not a domain field) and
// surgical edits preserve it; it is reserved and bumped only on a breaking
// frontmatter-shape change — coarse, not per additive field.
const FileSchemaVersion = 1
