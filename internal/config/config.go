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
			return &Config{Root: configuredRoot(dir)}, nil
		}
		if isDir(filepath.Join(dir, "tasks")) {
			return &Config{Root: dir}, nil
		}
		if isDir(filepath.Join(dir, "planning", "tasks")) {
			return &Config{Root: filepath.Join(dir, "planning")}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir || isDir(filepath.Join(dir, ".git")) {
			return nil, fmt.Errorf(
				"not a taskflow planning repo (no %s or tasks/ found from %s up) — run `tskflwctl init`",
				ConfigFile, start)
		}
		dir = parent
	}
}

// configuredRoot resolves the planning root from a dir holding ConfigFile,
// honoring its taskflow_root (default "."). The result is cleaned and kept
// within dir's tree.
func configuredRoot(dir string) string {
	rel := taskflowRoot(filepath.Join(dir, ConfigFile))
	return filepath.Join(dir, filepath.FromSlash(rel))
}

// taskflowRoot reads the taskflow_root value from a config file with a minimal
// line scan (no TOML dependency for one string key). Defaults to ".".
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
		if v = strings.Trim(strings.TrimSpace(v), `"'`); v != "" {
			return v
		}
	}
	return "."
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
func Init(root string) ([]string, error) {
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
		if err := os.MkdirAll(p, 0o755); err != nil {
			return created, fmt.Errorf("mkdir %s: %w", p, err)
		}
		created = append(created, d)
	}
	cfg := filepath.Join(root, ConfigFile)
	if !fileExists(cfg) {
		if err := os.WriteFile(cfg, []byte(defaultConfigTOML), 0o644); err != nil {
			return created, fmt.Errorf("write config: %w", err)
		}
		created = append(created, ConfigFile)
	}
	return created, nil
}
