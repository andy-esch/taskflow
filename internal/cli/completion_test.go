package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/userconfig"
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
	mustWrite(t, filepath.Join(root, "tasks", testutil.TaskID("broken")+"-broken.md"), "tags: a,b,c NOT yaml\n")

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
	auditPath, auditContent := testutil.AuditFixture(root, "open", "aud-sec.md", "---\narea: x\n---\n")
	mustWrite(t, auditPath, auditContent)
	mustWrite(t, filepath.Join(root, "epics", "17-pm.md"), "---\nstatus: active\n---\n")

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

func TestComplete_StatusAware_DamagedDuplicateRemainsAddressable(t *testing.T) {
	root := setupRepo(t) // beta is in-progress
	brokenID := testutil.TaskID("damaged-beta")
	mustWrite(t, filepath.Join(root, domain.TasksDir, brokenID+"-beta.md"), "damaged metadata\n")
	got := complete(t, "-C", root, "task", "start", "be")
	// Counting only eligible records would offer the ambiguous slug. Filtering
	// by slug rather than observed source would hide the damaged sibling entirely.
	if want := []string{brokenID}; !slices.Equal(got, want) {
		t.Fatalf("damaged duplicate completion=%v want=%v", got, want)
	}
	// Execute the emitted selector: a validation error proves it addressed the
	// damaged source, unlike the old id-slug suggestion's not-found failure.
	_, err := runRootStreams(t, "-C", root, "task", "show", got[0], "--json")
	if ExitCode(err) != 11 {
		t.Fatalf("damaged selector %q did not reach the record: %v", got[0], err)
	}
}

func TestCompleteTaskReferencesCanBeStartedAndKeepSiblings(t *testing.T) {
	root := freshRepo(t)
	first, second := "6fjangd7kva1", "6fjangd7kvb2"
	for _, id := range []string{first, second} {
		mustWrite(t, filepath.Join(root, domain.TasksDir, id+"-dup.md"), "---\n"+
			"schema: 1\nid: "+id+"\nstatus: ready-to-start\ndescription: Duplicate alias\n"+
			"tags: [fixture]\n---\n# Duplicate\n")
	}
	got := complete(t, "-C", root, "task", "start", "du")
	if !slices.Equal(got, []string{first, second}) {
		t.Fatalf("duplicate selectors=%v", got)
	}
	for _, ref := range got {
		out := runRoot(t, "-C", root, "task", "show", ref, "--json")
		var envelope struct{ Task struct{ ID string } }
		if err := json.Unmarshal([]byte(out), &envelope); err != nil || envelope.Task.ID != ref {
			// The fixture's explicit IDs are independent of completion's policy.
			t.Fatalf("selector=%q show=%q err=%v", ref, out, err)
		}
		runRoot(t, "-C", root, "task", "start", ref, "--dry-run", "--json")
	}
	if got := complete(t, "-C", root, "task", "start", first, "du"); !slices.Equal(got, []string{second}) {
		t.Fatalf("typed one record hid its sibling: %v", got)
	}
	runRoot(t, "-C", root, "task", "start", first, second, "--dry-run", "--json")
}

func TestCompleteFlatEntitiesResolveAliasCollisionsToObservedSource(t *testing.T) {
	for _, test := range []struct{ kind, dir string }{
		{"task", domain.TasksDir}, {"thread", domain.ThreadsDir},
		{"audit", domain.AuditsDir}, {"research", domain.ResearchDir},
	} {
		t.Run(test.kind, func(t *testing.T) {
			root := freshRepo(t)
			files := []struct{ id, slug string }{
				{"6fjangd7kva1", "dup"}, {"6fjangd7kvb2", "dup"},
				{"6fjangd7kvc3", "Case"}, {"6fjangd7kvd4", "case"},
				{"6fjangd7kve5", "6fjangd7kva1"}, {"6fjangd7kvf6", "unique"},
			}
			want := make(map[string]string)
			for _, file := range files {
				status := "ready-to-start"
				if test.kind == "thread" {
					status = "unstarted"
				}
				mustWrite(t, filepath.Join(root, test.dir, file.id+"-"+file.slug+".md"), "---\n"+
					"schema: 1\nid: "+file.id+"\nstatus: "+status+"\nbucket: open\n"+
					"description: Fixture\ngoal: Fixture\narea: fixture\ndate: \"2026-01-01\"\n"+
					"created: \"2026-01-01\"\ntasks: []\ntags: [fixture]\n---\n# Fixture\n")
				selector := file.id
				if file.slug == "unique" {
					selector = file.slug
				}
				want[selector] = file.id
			}
			got := complete(t, "-C", root, test.kind, "show", "")
			if len(got) != len(want) {
				t.Fatalf("suggestions=%v want selectors=%v", got, want)
			}
			for _, ref := range got {
				wantID, exists := want[ref]
				if !exists {
					t.Fatalf("unresolvable/unsafe suggestion %q", ref)
				}
				out := runRoot(t, "-C", root, test.kind, "show", ref, "--json")
				var envelope struct {
					Task, Audit, Research struct{ ID string }
					View                  struct{ Thread struct{ ID string } }
				}
				if err := json.Unmarshal([]byte(out), &envelope); err != nil {
					t.Fatalf("show %q: %v\n%s", ref, err, out)
				}
				gotID := map[string]string{"task": envelope.Task.ID, "audit": envelope.Audit.ID,
					"research": envelope.Research.ID, "thread": envelope.View.Thread.ID}[test.kind]
				if gotID != wantID {
					t.Fatalf("selector=%q addressed=%q want=%q", ref, gotID, wantID)
				}
			}
		})
	}
}

func TestComplete_StatusAware_AuditBuckets(t *testing.T) {
	root := setupRepo(t)
	openPath, openContent := testutil.AuditFixture(root, "open", "o.md", "---\narea: x\n---\n")
	mustWrite(t, openPath, openContent)
	closedPath, closedContent := testutil.AuditFixture(root, "closed", "c.md", "---\narea: y\n---\n")
	mustWrite(t, closedPath, closedContent)

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

// TestComplete_SpaceIDsOutsideRepo pins that registry labels are shell-addressable even
// when cwd has no planning tree. Prefix filtering and malformed-registry tolerance both
// stay independent of repo discovery.
func TestComplete_SpaceIDsOutsideRepo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(userconfig.DirEnv, dir)
	for _, s := range []userconfig.Space{
		{ID: "taskflow", Path: t.TempDir()},
		{ID: "desirelines", Path: t.TempDir()},
	} {
		if _, _, err := userconfig.AddSpace(s, false); err != nil {
			t.Fatal(err)
		}
	}
	bare := t.TempDir()
	if got := complete(t, "-C", bare, "space", "forget", "ta"); !has(got, "taskflow") || has(got, "desirelines") {
		t.Errorf("space completion should prefix-filter registry ids outside a repo: %v", got)
	}

	path, err := userconfig.SpacesPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("[[space]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := complete(t, "-C", bare, "space", "forget", ""); len(got) != 0 {
		t.Errorf("malformed registry should degrade completion to silence, got %v", got)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	testutil.Write(t, path, content)
}

// TestComplete_Templates pins the template discovery surface (no repo needed).
func TestComplete_Templates(t *testing.T) {
	if got := complete(t, "template", "show", ""); !has(got, "audit") || !has(got, "task") {
		t.Errorf("template show <kind> completion: %v", got)
	}
	if got := complete(t, "template", "show", "audit", ""); !has(got, "default") || !has(got, "security") {
		t.Errorf("template show audit <name> completion: %v", got)
	}
	if got := complete(t, "audit", "new", "--template", ""); !has(got, "security") {
		t.Errorf("audit new --template completion: %v", got)
	}
	if got := complete(t, "template", "list", "--kind", ""); !has(got, "audit") {
		t.Errorf("template list --kind completion: %v", got)
	}
}

func TestComplete_Threads(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, domain.ThreadsDir, "6g3q4rtmv4ak-delivery.md"), "---\nid: 6g3q4rtmv4ak\nstatus: unstarted\ndescription: delivery\ngoal: ship\ncreated: \"2026-08-29\"\ntasks: []\n---\n# Delivery\n")
	if got := complete(t, "-C", root, "thread", "show", "del"); !has(got, "delivery") {
		t.Errorf("thread ref completion: %v", got)
	}
	if got := complete(t, "-C", root, "thread", "list", "--status", "in"); !has(got, "in-progress") {
		t.Errorf("Thread status completion: %v", got)
	}
	if got := complete(t, "-C", root, "thread", "list", "-c", "graph_"); !has(got, "graph_health") {
		t.Errorf("Thread column completion: %v", got)
	}
}
