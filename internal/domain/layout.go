package domain

import "path"

// Planning-tree directory layout. The filesystem store owns absolute paths, but
// the directory *names* and the per-status/bucket subdir convention are domain
// knowledge — shared by the store (`NewFS`/`WatchPaths`), `config.Init`, and
// shell completion. Kept here so a rename, or a new status/bucket, lands in
// exactly one place instead of drifting across three call sites.
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

// TaskStatusDirs returns every task-status directory relative to the planning
// root ("tasks/<status>"), in status order.
func TaskStatusDirs() []string {
	out := make([]string, len(allStatuses))
	for i, st := range allStatuses {
		out[i] = path.Join(TasksDir, st.Dir())
	}
	return out
}

// AuditBucketDirs returns every audit-bucket directory relative to the planning
// root ("audits/<bucket>"), in bucket order.
func AuditBucketDirs() []string {
	out := make([]string, len(auditBuckets))
	for i, b := range auditBuckets {
		out[i] = path.Join(AuditsDir, b.Dir())
	}
	return out
}
