// Package config locates the planning data within a single repo.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/andy-esch/taskflow/internal/domain"
)

// Config records where the planning data lives (the "taskflow root": the dir
// holding tasks/, epics/, ...). One planning repo per product; no cross-product
// registry.
//
// PlanningRepo, when set, points at an EXTERNAL planning repo (the sanctioned
// out-of-tree escape): Discover resolves and validates it to choose Root, and it
// wins over taskflow_root. The raw (unresolved) value is also carried here for
// the linkback checks. TrackedRepos stays RAW and UNRESOLVED — a downstream task
// reads it.
type Config struct {
	Root         string
	PlanningRepo string
	TrackedRepos []string
}

// configFile mirrors the on-disk .tskflwctl.toml schema for a real TOML decode.
// Defaults (taskflow_root ".", the rest empty) are applied by readConfigFile.
type configFile struct {
	TaskflowRoot string   `toml:"taskflow_root"`
	PlanningRepo string   `toml:"planning_repo"`
	TrackedRepos []string `toml:"tracked_repos"`
}

// Discover walks up from start to find the planning root. At each level it
// prefers an explicit `.tskflwctl.toml` marker (honoring its planning_repo if
// set, else taskflow_root), then falls back to a tasks/ directory or a
// planning/tasks/ subdir. It terminates at a .git boundary or the filesystem
// root — never an infinite climb.
func Discover(start string) (*Config, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", start, err)
	}
	// Resolve symlinks so the .git-boundary walk-up and the discovered root use
	// PHYSICAL ancestry — a symlinked worktree must climb its real parents, not its
	// logical ones. evalOr falls back to the lexical path when start doesn't exist
	// yet (a fresh dir), so discovery still errs sensibly there.
	dir = evalOr(dir)
	for {
		if fileExists(filepath.Join(dir, ConfigFile)) {
			cf, err := readConfigFile(filepath.Join(dir, ConfigFile))
			if err != nil {
				return nil, err // malformed TOML is loud, never a silent default
			}
			root, err := resolveRoot(dir, cf)
			if err != nil {
				return nil, err // loud, never a silently empty/forked tree
			}
			// resolveRoot already followed/validated planning_repo into Root; the
			// raw planning_repo + tracked_repos ride along for the linkback checks.
			return &Config{
				Root:         root,
				PlanningRepo: cf.PlanningRepo,
				TrackedRepos: cf.TrackedRepos,
			}, nil
		}
		if isDir(filepath.Join(dir, domain.TasksDir)) {
			return &Config{Root: dir}, nil
		}
		if isDir(filepath.Join(dir, "planning", domain.TasksDir)) {
			return &Config{Root: filepath.Join(dir, "planning")}, nil
		}
		parent := filepath.Dir(dir)
		// .git is the climb boundary whether it's a directory OR a file — in a
		// git worktree/submodule it's a file pointing elsewhere, and missing it
		// would over-climb into a parent's planning tree.
		if parent == dir || exists(filepath.Join(dir, ".git")) {
			return nil, fmt.Errorf(
				"not a taskflow planning repo (no %s or tasks/ found from %s up) — run `tskflwctl init`",
				ConfigFile, start)
		}
		dir = parent
	}
}

// configuredRoot resolves the planning root from dir, honoring the already-parsed
// taskflow_root rel value (default "" → "."). The result must stay within dir's
// tree and must look like a planning root (contain tasks/): a typo'd or
// escaping value previously presented a clean EMPTY project — and `task new`
// would then fork the data into a second tree — so both are loud errors now.
func configuredRoot(dir, rel string) (string, error) {
	if rel == "" {
		rel = "."
	}
	root := filepath.Join(dir, filepath.FromSlash(rel))
	// Containment must be PHYSICAL: filepath.Rel is lexical and can't see that a
	// `planning -> /etc` symlink escapes dir's real tree. Resolve both ends before
	// the no-`..` check (evalOr leaves a not-yet-existing path lexical, which then
	// fails the tasks/-exists check below with a clear message).
	if r, err := filepath.Rel(evalOr(dir), evalOr(root)); err != nil ||
		r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: taskflow_root %q escapes %s's directory (%s)", domain.ErrValidation, rel, ConfigFile, dir)
	}
	if !isDir(filepath.Join(root, domain.TasksDir)) {
		return "", fmt.Errorf(
			"%w: taskflow_root %q points at %s, which has no tasks/ — fix %s or run `tskflwctl init`",
			domain.ErrValidation, rel, root, ConfigFile)
	}
	return root, nil
}

// resolveRoot picks the planning root from a parsed config. planning_repo, when
// set, is the sanctioned out-of-tree escape and wins; taskflow_root is the
// in-tree (containment-checked) default. Setting both to different roots is a
// conflict, not a silent override — taskflow_root "" / "." is the only value
// that coexists with planning_repo (it just means "the default in-tree root",
// which planning_repo overrides).
func resolveRoot(dir string, cf configFile) (string, error) {
	if cf.PlanningRepo == "" {
		return configuredRoot(dir, cf.TaskflowRoot)
	}
	// Only a default-equivalent taskflow_root coexists with planning_repo; anything
	// else names a SECOND root, which is a conflict, not a silent override.
	// Normalize first so "./", " . ", "./." — all the in-tree root — aren't
	// mistaken for a different root than ".".
	if tr := strings.TrimSpace(cf.TaskflowRoot); tr != "" && filepath.Clean(filepath.FromSlash(tr)) != "." {
		return "", fmt.Errorf(
			"%w: %s sets both planning_repo (%q) and taskflow_root (%q), which name different roots — keep one",
			domain.ErrConflict, ConfigFile, cf.PlanningRepo, cf.TaskflowRoot)
	}
	return resolvePlanningRepo(dir, cf.PlanningRepo)
}

// resolvePlanningRepo resolves planning_repo relative to the config dir (an
// absolute value is used as-is) and validates it is a real planning root. Unlike
// taskflow_root, it is ALLOWED to escape dir's tree — that is the entire point
// of pointing an impl repo at an external planning repo. A target without tasks/
// is a loud error (the "require + validate" contract), never a clean empty tree.
func resolvePlanningRepo(dir, planningRepo string) (string, error) {
	root := filepath.FromSlash(planningRepo)
	if !filepath.IsAbs(root) {
		root = filepath.Join(dir, root)
	}
	root = filepath.Clean(root)
	if !isDir(filepath.Join(root, domain.TasksDir)) {
		return "", fmt.Errorf(
			"%w: planning_repo %q points at %s, which has no tasks/ — run `tskflwctl init` there first",
			domain.ErrValidation, planningRepo, root)
	}
	// Resolve symlinks so Root is PHYSICAL, matching the in-tree branch (Discover
	// evalOr's dir, which configuredRoot inherits). The linkback work compares
	// physical paths, so a symlinked external planning repo must not leave Root
	// logical here. evalOr falls back to the lexical path if it can't resolve.
	return evalOr(root), nil
}

// readConfigFile decodes ConfigFile into a configFile with a real TOML parser.
// A missing file decodes to zero values (the caller already gated on existence,
// but this stays lenient on a vanished file). Malformed TOML is a LOUD error
// wrapping domain.ErrValidation — never a silent default that would present a
// clean empty project. Absent keys keep their zero values: taskflow_root is
// defaulted to "." downstream in configuredRoot; planning_repo/tracked_repos
// stay "" / nil.
func readConfigFile(configPath string) (configFile, error) {
	var cf configFile
	if _, err := toml.DecodeFile(configPath, &cf); err != nil {
		if os.IsNotExist(err) {
			return configFile{}, nil
		}
		return configFile{}, fmt.Errorf("%w: parse %s: %w", domain.ErrValidation, ConfigFile, err)
	}
	return cf, nil
}

// evalOr resolves symlinks in p, falling back to p itself when it can't (e.g. p
// doesn't exist yet) — so containment and boundary checks operate on physical
// paths without breaking on a not-yet-created directory.
func evalOr(p string) string {
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return p
}

// exists reports whether path exists as anything (file or directory).
func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// ConfigFile is the per-repo config filename written by Init.
const ConfigFile = ".tskflwctl.toml"

// gitKeep is the placeholder Init drops in each scaffolded dir so an empty
// planning tree is still git-committable (git won't track empty directories).
const gitKeep = ".gitkeep"

const defaultConfigTOML = `# tskflwctl planning config — also the marker that anchors discovery.
# taskflow_root: planning dir relative to this file (default ".").
taskflow_root = "."
# tracked_repos: reserved for future multi-repo tracking (not yet read).
tracked_repos = []
`

// Init scaffolds the planning directory tree and writes the config file under
// root. It is idempotent: existing dirs/config are left untouched. Returns the
// relative paths created (empty if nothing was needed).
func Init(root string, dryRun bool) ([]string, error) {
	// Derive the task-status and audit-bucket dirs from the domain layout helpers
	// so a new status/bucket is scaffolded automatically — a hardcoded list would
	// silently drift (init not creating a dir the watcher already watches).
	// Guarded by TestInitScaffoldsEveryStatusAndBucket.
	dirs := domain.TaskStatusDirs()
	dirs = append(dirs, domain.EpicsDir, domain.ProjectsDir)
	dirs = append(dirs, domain.AuditBucketDirs()...)
	var created []string
	for _, d := range dirs {
		p := filepath.Join(root, filepath.FromSlash(d))
		if !isDir(p) {
			if !dryRun {
				if err := os.MkdirAll(p, 0o755); err != nil {
					return created, fmt.Errorf("mkdir %s: %w", p, err)
				}
			}
			created = append(created, d)
		}
		// A .gitkeep makes the (otherwise empty) dir git-committable. Written if
		// absent even when the dir already exists, so re-running init also repairs
		// a tree scaffolded before this existed.
		keep := filepath.Join(p, gitKeep)
		if !fileExists(keep) {
			if !dryRun {
				if err := os.WriteFile(keep, nil, 0o644); err != nil {
					return created, fmt.Errorf("write %s: %w", keep, err)
				}
			}
			created = append(created, d+"/"+gitKeep)
		}
	}
	cfg := filepath.Join(root, ConfigFile)
	// Exclusive create instead of exists-then-write: a concurrent init must not
	// clobber a config the other process just wrote (idempotency falls out of
	// O_EXCL rather than a racy stat).
	if dryRun {
		if !fileExists(cfg) {
			created = append(created, ConfigFile)
		}
		return created, nil
	}
	switch err := writeFileExclusive(cfg, []byte(defaultConfigTOML), 0o644); {
	case err == nil:
		created = append(created, ConfigFile)
	case !os.IsExist(err):
		return created, fmt.Errorf("write config: %w", err)
	}
	return created, nil
}

// writeFileExclusive creates a new file with O_EXCL semantics. (The store's
// createFileAtomic is the same idea; config can't import store without
// inverting the dependency, so the few lines are inlined.)
func writeFileExclusive(path string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err // os.IsExist(err) when the config already exists
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return err
	}
	return f.Close()
}
