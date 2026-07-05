// Package testutil builds throwaway planning trees for tests across packages.
// It owns the single "mkdir + write a fixture file" implementation that the
// per-package helpers (store's writeTask, cli's mustWrite, core/tui's repo
// builders) previously each re-rolled — so the fixture convention lives in one
// place. It is test-only; nothing here is imported by non-test code.
package testutil

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

// idAlphabet mirrors internal/id's Crockford base32 (lowercase); a fixture id must
// pass id.Valid to be recognized as an entity under the flat layout.
const idAlphabet = "0123456789abcdefghjkmnpqrstvwxyz"

// TaskID derives a stable, valid 12-char id from a seed (a slug), so flat-layout
// task fixtures get a deterministic id-led filename (tasks/<id>-<slug>.md) without
// threading real minted ids through every test.
func TaskID(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	b := make([]byte, id.Length)
	for i := range b {
		b[i] = idAlphabet[int(sum[i])%len(idAlphabet)]
	}
	return string(b)
}

// TaskFixture is the flatten migration shim: it maps the old dir-as-status fixture
// convention (a `status` dir + a `name.md`) onto the flat layout — a stable id-led
// path tasks/<id>-<slug>.md, and, when the frontmatter block declares no status,
// injects the dir status into it (status is authoritative in frontmatter now,
// ADR-0003 §4). A fence-less body is left untouched (it stays a loud FileProblem).
func TaskFixture(root, status, name, content string) (path, out string) {
	slug := strings.TrimSuffix(name, ".md")
	if strings.HasPrefix(content, "---\n") && !strings.Contains(content, "status:") {
		content = "---\nstatus: " + status + "\n" + content[len("---\n"):]
	}
	return filepath.Join(root, domain.TasksDir, TaskID(slug)+"-"+name), content
}

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

// Task writes a flat task fixture tasks/<id>-<name> (see TaskFixture); status is
// injected into the frontmatter when the content declares none.
func (r *Repo) Task(status, name, content string) *Repo {
	path, out := TaskFixture(r.Root, status, name, content)
	Write(r.t, path, out)
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
