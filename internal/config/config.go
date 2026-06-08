// Package config locates the planning data within a single repo.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config records where the planning data lives (the "taskflow root": the dir
// holding tasks/, epics/, ...). One planning repo per product; no cross-product
// registry.
type Config struct {
	Root string
}

// Discover walks up from start to find the planning root: a dir containing a
// tasks/ directory, or one with a planning/tasks/ subdir. It terminates at a
// .git boundary or the filesystem root — never an infinite climb.
func Discover(start string) (*Config, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", start, err)
	}
	for {
		if isDir(filepath.Join(dir, "tasks")) {
			return &Config{Root: dir}, nil
		}
		if isDir(filepath.Join(dir, "planning", "tasks")) {
			return &Config{Root: filepath.Join(dir, "planning")}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir || isDir(filepath.Join(dir, ".git")) {
			return nil, fmt.Errorf(
				"not a taskflow planning repo (no tasks/ found from %s up) — run `tskflwctl init`", start)
		}
		dir = parent
	}
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

const defaultConfigTOML = `# tskflwctl planning config
taskflow_root = "."
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
