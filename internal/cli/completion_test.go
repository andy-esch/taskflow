package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
)

// complete runs the hidden __complete driver in-process and returns the
// candidate slugs (dropping cobra's `:directive` line and any tab-descriptions).
func complete(t *testing.T, args ...string) []string {
	t.Helper()
	out := runRoot(t, append([]string{"__complete"}, args...)...)
	var got []string
	for _, ln := range strings.Split(out, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, ":") || strings.HasPrefix(ln, "Completion ended") {
			continue
		}
		if i := strings.IndexByte(ln, '\t'); i >= 0 {
			ln = ln[:i]
		}
		got = append(got, ln)
	}
	return got
}

func has(slugs []string, want string) bool {
	for _, s := range slugs {
		if s == want {
			return true
		}
	}
	return false
}

func TestComplete_TaskSlugs_IncludesMalformed(t *testing.T) {
	root := setupRepo(t) // alpha (ready-to-start), beta (in-progress)
	// A file whose frontmatter doesn't parse must still complete — you complete
	// it precisely to go fix it. (No YAML is parsed for completion.)
	mustWrite(t, filepath.Join(root, "tasks", "ready-to-start", "broken.md"), "tags: a,b,c NOT yaml\n")

	got := complete(t, "-C", root, "task", "show", "")
	for _, want := range []string{"alpha", "beta", "broken"} {
		if !has(got, want) {
			t.Errorf("completion missing %q: %v", want, got)
		}
	}
}

func TestComplete_TaskSlugs_PrefixFilters(t *testing.T) {
	root := setupRepo(t)
	got := complete(t, "-C", root, "task", "start", "al")
	if !has(got, "alpha") || has(got, "beta") {
		t.Errorf("prefix 'al' should yield alpha only: %v", got)
	}
}

func TestComplete_TaskSlugs_DropsAlreadyTyped(t *testing.T) {
	root := setupRepo(t)
	got := complete(t, "-C", root, "task", "move", "alpha", "")
	if has(got, "alpha") {
		t.Errorf("already-typed 'alpha' should not be re-suggested: %v", got)
	}
	if !has(got, "beta") {
		t.Errorf("expected beta still offered: %v", got)
	}
}

func TestComplete_AuditAndEpic(t *testing.T) {
	root := setupRepo(t)
	mustWrite(t, filepath.Join(root, "audits", "open", "aud-sec.md"), "---\narea: x\n---\n")
	mustWrite(t, filepath.Join(root, "epics", "17-pm.md"), "---\nstatus: in-progress\n---\n")

	if got := complete(t, "-C", root, "audit", "close", ""); !has(got, "aud-sec") {
		t.Errorf("audit completion missing aud-sec: %v", got)
	}
	if got := complete(t, "-C", root, "epic", "show", ""); !has(got, "17-pm") {
		t.Errorf("epic completion missing 17-pm: %v", got)
	}
}

func TestComplete_StatusAware_TaskTransitions(t *testing.T) {
	root := setupRepo(t) // alpha (ready-to-start), beta (in-progress)

	// `start` → in-progress: should NOT offer beta (already in-progress).
	if got := complete(t, "-C", root, "task", "start", ""); !has(got, "alpha") || has(got, "beta") {
		t.Errorf("start should offer alpha but not the in-progress beta: %v", got)
	}
	// `complete` → completed: neither is completed, so both are offered.
	if got := complete(t, "-C", root, "task", "complete", ""); !has(got, "alpha") || !has(got, "beta") {
		t.Errorf("complete should offer both: %v", got)
	}
}

func TestComplete_StatusAware_AuditBuckets(t *testing.T) {
	root := setupRepo(t)
	mustWrite(t, filepath.Join(root, "audits", "open", "o.md"), "---\narea: x\n---\n")
	mustWrite(t, filepath.Join(root, "audits", "closed", "c.md"), "---\narea: y\n---\n")

	// `close` → closed: should NOT offer the already-closed c.
	if got := complete(t, "-C", root, "audit", "close", ""); !has(got, "o") || has(got, "c") {
		t.Errorf("close should offer open o but not closed c: %v", got)
	}
	// `reopen` → open: should NOT offer the already-open o.
	if got := complete(t, "-C", root, "audit", "reopen", ""); !has(got, "c") || has(got, "o") {
		t.Errorf("reopen should offer closed c but not open o: %v", got)
	}
}

func TestComplete_OutsideRepo_Quiet(t *testing.T) {
	bare := t.TempDir() // no tasks/ anywhere — not a planning repo
	got := complete(t, "-C", bare, "task", "show", "")
	if len(got) != 0 {
		t.Errorf("expected no candidates outside a repo, got %v", got)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	testutil.Write(t, path, content)
}
