// Package testutil builds throwaway planning trees for tests across packages.
// It owns the single "mkdir + write a fixture file" implementation that the
// per-package helpers (store's writeTask, cli's mustWrite, core/tui's repo
// builders) previously each re-rolled — so the fixture convention lives in one
// place. It is test-only; nothing here is imported by non-test code.
package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// Write writes content to path, creating parent dirs. Fatal on error.
func Write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Repo is a temp planning tree with chainable writers for the entity dirs.
type Repo struct {
	t    *testing.T
	Root string
}

// NewRepo returns an empty planning tree rooted at t.TempDir().
func NewRepo(t *testing.T) *Repo {
	t.Helper()
	return &Repo{t: t, Root: t.TempDir()}
}

// Task writes tasks/<status>/<name>.
func (r *Repo) Task(status, name, content string) *Repo {
	Write(r.t, filepath.Join(r.Root, domain.TasksDir, status, name), content)
	return r
}

// Epic writes epics/<name>.
func (r *Repo) Epic(name, content string) *Repo {
	Write(r.t, filepath.Join(r.Root, domain.EpicsDir, name), content)
	return r
}

// Audit writes audits/<bucket>/<name>.
func (r *Repo) Audit(bucket, name, content string) *Repo {
	Write(r.t, filepath.Join(r.Root, domain.AuditsDir, bucket, name), content)
	return r
}

// File writes an arbitrary slash-relative path under the root.
func (r *Repo) File(rel, content string) *Repo {
	Write(r.t, filepath.Join(r.Root, filepath.FromSlash(rel)), content)
	return r
}
