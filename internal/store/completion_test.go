package store

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestCompletionSourceRetainsMalformedRecordsAndIndependentIdentity(t *testing.T) {
	root := t.TempDir()
	sourceID, declaredID := testutil.TaskID("completion-source"), testutil.TaskID("completion-declaration")
	for _, test := range []struct {
		kind core.EntityKind
		dir  string
	}{
		{core.EntityTask, domain.TasksDir}, {core.EntityThread, domain.ThreadsDir},
		{core.EntityAudit, domain.AuditsDir}, {core.EntityResearch, domain.ResearchDir},
	} {
		t.Run(string(test.kind), func(t *testing.T) {
			dir := filepath.Join(root, test.dir)
			testutil.Write(t, filepath.Join(dir, sourceID+"-broken.md"), "not frontmatter\n")
			testutil.Write(t, filepath.Join(dir, "README.md"), "not a record\n")
			testutil.Write(t, filepath.Join(dir, "stray.md"), "not an id-led record\n")
			for _, state := range []bool{false, true} {
				got, err := NewFS(root).ReadCompletionCandidates(test.kind, state)
				want := []core.CompletionCandidate{{ID: sourceID, Slug: "broken", Reference: sourceID, SlugIsReference: true}}
				if err != nil || !slices.Equal(got, want) {
					t.Fatalf("state=%t candidates=%+v err=%v", state, got, err)
				}
			}
		})
	}
	// A parsed declaration cannot replace the independent filename identity.
	driftID := testutil.TaskID("completion-drift-source")
	path := filepath.Join(root, domain.TasksDir, driftID+"-drift.md")
	testutil.Write(t, path, "---\nid: "+declaredID+"\nstatus: in-progress\n---\n# Drift\n")
	got, err := NewFS(root).ReadCompletionCandidates(core.EntityTask, true)
	if err != nil || len(got) != 2 {
		t.Fatalf("drift candidates=%+v err=%v", got, err)
	}
	for _, candidate := range got {
		if candidate.Slug == "drift" && (candidate.ID != driftID || candidate.State != "in-progress") {
			t.Fatalf("declaration replaced completion source: %+v", candidate)
		}
	}
}

func TestCompletionSourceSkipsSymlinks(t *testing.T) {
	root := t.TempDir()
	sourceID := testutil.TaskID("completion-link")
	target := filepath.Join(t.TempDir(), "outside.md")
	testutil.Write(t, target, "malformed but should not be scanned\n")
	dir := filepath.Join(root, domain.TasksDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(dir, sourceID+"-symlink.md")); err != nil {
		t.Skipf("symlinks unsupported here: %v", err)
	}
	for _, withState := range []bool{false, true} {
		got, err := NewFS(root).ReadCompletionCandidates(core.EntityTask, withState)
		if err != nil || len(got) != 0 {
			t.Fatalf("withState=%t followed symlink: candidates=%+v err=%v", withState, got, err)
		}
	}
}

func TestCompletionSourcePlainQueriesDoNotParseState(t *testing.T) {
	root := t.TempDir()
	sourceID := testutil.TaskID("state-completion")
	testutil.Write(t, filepath.Join(root, domain.TasksDir, sourceID+"-task.md"), "---\nstatus: in-progress\n---\n")
	testutil.Write(t, filepath.Join(root, domain.AuditsDir, sourceID+"-audit.md"), "---\nbucket: closed\n---\n")
	testutil.Write(t, filepath.Join(root, domain.EpicsDir, "17-project.md"), "malformed but addressable\n")
	for _, test := range []struct {
		kind  core.EntityKind
		state string
	}{
		{core.EntityTask, "in-progress"}, {core.EntityAudit, "closed"},
	} {
		for _, withState := range []bool{false, true} {
			got, err := NewFS(root).ReadCompletionCandidates(test.kind, withState)
			want := ""
			if withState {
				want = test.state
			}
			if err != nil || len(got) != 1 || got[0].State != want {
				t.Fatalf("kind=%s withState=%t candidates=%+v err=%v", test.kind, withState, got, err)
			}
		}
	}
	epic, err := NewFS(root).ReadCompletionCandidates(core.EntityEpic, false)
	if err != nil || !slices.Equal(epic, []core.CompletionCandidate{{ID: "17-project", Slug: "17-project", Reference: "17-project", SlugIsReference: true}}) {
		t.Fatalf("epic candidates=%+v err=%v", epic, err)
	}
	missing, err := NewFS(t.TempDir()).ReadCompletionCandidates(core.EntityTask, false)
	if err != nil || len(missing) != 0 {
		t.Fatalf("missing directory candidates=%+v err=%v", missing, err)
	}
}

func TestCompletionSourceReferencesMatchActualResolver(t *testing.T) {
	for _, test := range []struct {
		kind core.EntityKind
		dir  string
	}{
		{core.EntityTask, domain.TasksDir}, {core.EntityThread, domain.ThreadsDir},
		{core.EntityAudit, domain.AuditsDir}, {core.EntityResearch, domain.ResearchDir},
	} {
		t.Run(string(test.kind), func(t *testing.T) {
			root := t.TempDir()
			firstID := testutil.TaskID("resolver-first")
			duplicateID := testutil.TaskID("resolver-duplicate-id")
			files := []struct{ id, slug string }{
				{firstID, "dup"}, {testutil.TaskID("resolver-second"), "dup"},
				{testutil.TaskID("resolver-upper"), "Case"}, {testutil.TaskID("resolver-lower"), "case"},
				{testutil.TaskID("resolver-shadow"), firstID}, {testutil.TaskID("resolver-unique"), "unique"},
				{testutil.TaskID("resolver-invalid-query"), "bad..query"},
				{duplicateID, "one"}, {duplicateID, "two"},
			}
			paths := make(map[string]string)
			for _, file := range files {
				path := filepath.Join(root, test.dir, file.id+"-"+file.slug+".md")
				testutil.Write(t, path, "damaged metadata\n")
				paths[file.id] = path
			}
			fs := NewFS(root)
			var resolve func(string) (string, error)
			switch test.kind {
			case core.EntityTask:
				resolve = fs.ResolveTaskPath
			case core.EntityThread:
				resolve = fs.ResolveThreadPath
			case core.EntityAudit:
				resolve = fs.ResolveAuditPath
			case core.EntityResearch:
				resolve = fs.ResolveResearchPath
			}
			for _, withState := range []bool{false, true} {
				got, err := fs.ReadCompletionCandidates(test.kind, withState)
				if err != nil || len(got) != len(files)-2 {
					t.Fatalf("candidates=%+v err=%v", got, err)
				}
				for _, candidate := range got {
					path, err := resolve(candidate.Reference)
					if candidate.ID == duplicateID || err != nil || path != paths[candidate.ID] {
						t.Fatalf("reference selected another source: candidate=%+v path=%q err=%v", candidate, path, err)
					}
					wantAlias := candidate.Slug == "unique"
					if candidate.SlugIsReference != wantAlias {
						t.Fatalf("alias safety=%t want=%t for %+v", candidate.SlugIsReference, wantAlias, candidate)
					}
					if candidate.SlugIsReference {
						path, err := resolve(candidate.Slug)
						if err != nil || path != paths[candidate.ID] {
							t.Fatalf("certified alias selected another source: candidate=%+v path=%q err=%v", candidate, path, err)
						}
					}
				}
			}
		})
	}
}

func TestCompletionSourceEpicCarveoutMatchesResolver(t *testing.T) {
	root := t.TempDir()
	for _, stem := range []string{"README", "readme", "ReadMe", "17-project", "legacy", "bad..query"} {
		testutil.Write(t, filepath.Join(root, domain.EpicsDir, stem+".md"), "damaged metadata\n")
	}
	fs := NewFS(root)
	got, err := fs.ReadCompletionCandidates(core.EntityEpic, false)
	if err != nil || len(got) != 2 {
		t.Fatalf("epic candidates=%+v err=%v", got, err)
	}
	for _, candidate := range got {
		path, err := fs.ResolveEpicPath(candidate.Reference)
		if err != nil || path != filepath.Join(root, domain.EpicsDir, candidate.ID+".md") {
			t.Fatalf("epic reference=%+v path=%q err=%v", candidate, path, err)
		}
	}
}
