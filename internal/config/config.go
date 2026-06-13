// Package config locates the planning data within a single repo.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config records where the planning data lives (the "taskflow root": the dir
// holding tasks/, epics/, ...). One planning repo per product; no cross-product
// registry.
type Config struct {
	Root string
}

// Discover walks up from start to find the planning root. At each level it
// prefers an explicit `.tskflwctl.toml` marker (honoring its taskflow_root),
// then falls back to a tasks/ directory or a planning/tasks/ subdir. It
// terminates at a .git boundary or the filesystem root — never an infinite
// climb.
func Discover(start string) (*Config, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", start, err)
	}
	for {
		if fileExists(filepath.Join(dir, ConfigFile)) {
			root, err := configuredRoot(dir)
			if err != nil {
				return nil, err // loud, never a silently empty/forked tree
			}
			return &Config{Root: root}, nil
		}
		if isDir(filepath.Join(dir, "tasks")) {
			return &Config{Root: dir}, nil
		}
		if isDir(filepath.Join(dir, "planning", "tasks")) {
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

// configuredRoot resolves the planning root from a dir holding ConfigFile,
// honoring its taskflow_root (default "."). The result must stay within dir's
// tree and must look like a planning root (contain tasks/): a typo'd or
// escaping value previously presented a clean EMPTY project — and `task new`
// would then fork the data into a second tree — so both are loud errors now.
func configuredRoot(dir string) (string, error) {
	rel := taskflowRoot(filepath.Join(dir, ConfigFile))
	root := filepath.Join(dir, filepath.FromSlash(rel))
	if r, err := filepath.Rel(dir, root); err != nil ||
		r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("taskflow_root %q escapes %s's directory (%s)", rel, ConfigFile, dir)
	}
	if !isDir(filepath.Join(root, "tasks")) {
		return "", fmt.Errorf(
			"taskflow_root %q points at %s, which has no tasks/ — fix %s or run `tskflwctl init`",
			rel, root, ConfigFile)
	}
	return root, nil
}

// taskflowRoot reads the taskflow_root value from a config file with a minimal
// line scan (no TOML dependency for one string key). Defaults to ".".
// Quoted values are extracted properly and an inline `# comment` (valid TOML)
// is stripped — previously `taskflow_root = "." # note` yielded a garbage path.
func taskflowRoot(configPath string) string {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "."
	}
	for _, ln := range strings.Split(string(data), "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		k, v, ok := strings.Cut(ln, "=")
		if !ok || strings.TrimSpace(k) != "taskflow_root" {
			continue
		}
		if v = tomlStringValue(v); v != "" {
			return v
		}
	}
	return "."
}

// tomlStringValue extracts a string value from the raw right-hand side of a
// `key = value` line: the quoted segment if quoted (comment after it ignored),
// otherwise the text up to an inline `#` comment, trimmed.
func tomlStringValue(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if q := raw[0]; q == '"' || q == '\'' {
		if end := strings.IndexByte(raw[1:], q); end >= 0 {
			return raw[1 : 1+end]
		}
		return "" // unterminated quote: treat as unset rather than guess
	}
	if i := strings.IndexByte(raw, '#'); i >= 0 {
		raw = raw[:i]
	}
	return strings.TrimSpace(raw)
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
	dirs := []string{
		"tasks/next-up", "tasks/ready-to-start", "tasks/in-progress",
		"tasks/completed", "tasks/deprecated", "tasks/deferred",
		"epics", "projects",
		"audits/open", "audits/closed", "audits/deferred",
	}
	var created []string
	for _, d := range dirs {
		p := filepath.Join(root, filepath.FromSlash(d))
		if isDir(p) {
			continue
		}
		if !dryRun {
			if err := os.MkdirAll(p, 0o755); err != nil {
				return created, fmt.Errorf("mkdir %s: %w", p, err)
			}
		}
		created = append(created, d)
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
