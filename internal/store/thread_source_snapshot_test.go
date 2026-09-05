package store

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestReadThreadsVersionsExactUnreadableSourceBytes(t *testing.T) {
	root := t.TempDir()
	threadID := testutil.TaskID("versioned-unreadable-thread")
	path := filepath.Join(root, domain.ThreadsDir, threadID+"-versioned-unreadable-thread.md")
	firstContent := "---\nid: [unterminated\n---\n# First body\n"
	secondContent := "---\nid: [unterminated\n---\n# Different body\n"
	testutil.Write(t, path, firstContent)
	fs := NewFS(root)

	first, err := fs.ReadThreads()
	if err != nil || len(first.Threads) != 0 || len(first.Problems) != 1 {
		t.Fatalf("first read=%+v err=%v", first, err)
	}
	problem := first.Problems[0]
	if problem.ThreadID != threadID || problem.ThreadSlug != "versioned-unreadable-thread" ||
		problem.Location != path || problem.SourceVersion != hashContent([]byte(firstContent)) {
		t.Fatalf("first problem = %+v", problem)
	}
	encoded, err := json.Marshal(problem)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), problem.SourceVersion) || strings.Contains(string(encoded), "SourceVersion") {
		t.Fatalf("opaque Thread source revision leaked through JSON: %s", encoded)
	}

	testutil.Write(t, path, secondContent)
	second, err := fs.ReadThreads()
	if err != nil || len(second.Problems) != 1 {
		t.Fatalf("second read=%+v err=%v", second, err)
	}
	if second.Problems[0].Message != problem.Message {
		t.Fatalf("fixture changed diagnostic:\nfirst: %s\nsecond: %s", problem.Message, second.Problems[0].Message)
	}
	if second.Problems[0].SourceVersion == problem.SourceVersion || second.Problems[0].SourceVersion != hashContent([]byte(secondContent)) {
		t.Fatalf("second source revision = %q", second.Problems[0].SourceVersion)
	}
	if err := verifyThreadSourceSnapshot(first, second); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("changed unreadable bytes error = %v, want ErrConflict", err)
	}

	testutil.Write(t, path, firstContent)
	restored, err := fs.ReadThreads()
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyThreadSourceSnapshot(first, restored); err != nil {
		t.Fatalf("byte-identical restored snapshot = %v", err)
	}
}

func TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed(t *testing.T) {
	firstID := testutil.TaskID("first-remote-thread-problem")
	secondID := testutil.TaskID("second-remote-thread-problem")
	problems := []core.ThreadReadProblem{
		{ThreadID: firstID, ThreadSlug: "first", Message: "bad remote row", SourceVersion: "opaque-a"},
		{ThreadID: secondID, ThreadSlug: "second", Message: "bad remote row", SourceVersion: "opaque-b"},
	}
	left := core.ThreadRead{Problems: append([]core.ThreadReadProblem(nil), problems...)}
	reordered := core.ThreadRead{Problems: []core.ThreadReadProblem{problems[1], problems[0]}}
	if err := verifyThreadSourceSnapshot(left, reordered); err != nil {
		t.Fatalf("reordered pathless problems = %v", err)
	}
	if !slices.Equal(left.Problems, problems) {
		t.Fatalf("snapshot comparison mutated caller problems: %+v", left.Problems)
	}

	changed := core.ThreadRead{Problems: append([]core.ThreadReadProblem(nil), problems...)}
	changed.Problems[0].SourceVersion = "opaque-changed"
	if err := verifyThreadSourceSnapshot(left, changed); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("changed opaque revision error = %v", err)
	}

	missing := core.ThreadRead{Problems: append([]core.ThreadReadProblem(nil), problems...)}
	missing.Problems[0].SourceVersion = ""
	if err := verifyThreadSourceSnapshot(missing, missing); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("two unversioned unreadable snapshots error = %v, want ErrConflict", err)
	}
}

func TestThreadSourceSnapshotRejectsRepresentationAndIdentityChanges(t *testing.T) {
	threadID := testutil.TaskID("thread-source-transition")
	thread := domain.Thread{
		ID: threadID, FilenameID: threadID, Slug: "thread-source-transition",
		Path: "threads/" + threadID + "-thread-source-transition.md", SourceVersion: "opaque-readable",
	}
	readable := core.ThreadRead{Threads: []domain.Thread{thread}}
	if err := verifyThreadSourceSnapshot(readable, readable); err != nil {
		t.Fatalf("identical readable snapshot = %v", err)
	}

	drifted := thread
	drifted.ID = testutil.TaskID("thread-source-drift")
	if err := verifyThreadSourceSnapshot(readable, core.ThreadRead{Threads: []domain.Thread{drifted}}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("identity drift error = %v", err)
	}

	unreadable := core.ThreadRead{Problems: []core.ThreadReadProblem{{
		ThreadID: threadID, ThreadSlug: thread.Slug, Location: thread.Path,
		Message: "row became unreadable", SourceVersion: "opaque-unreadable",
	}}}
	if err := verifyThreadSourceSnapshot(readable, unreadable); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("readable-to-unreadable error = %v", err)
	}
	if err := verifyThreadSourceSnapshot(unreadable, readable); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("unreadable-to-readable error = %v", err)
	}

	renamed := unreadable
	renamed.Problems = append([]core.ThreadReadProblem(nil), unreadable.Problems...)
	renamed.Problems[0].Location = "threads/" + threadID + "-renamed.md"
	if err := verifyThreadSourceSnapshot(unreadable, renamed); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("unreadable rename error = %v", err)
	}
	if err := verifyThreadSourceSnapshot(unreadable, core.ThreadRead{}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("unreadable removal error = %v", err)
	}
	added := unreadable
	added.Problems = append(append([]core.ThreadReadProblem(nil), unreadable.Problems...), core.ThreadReadProblem{
		ThreadID: testutil.TaskID("added-unreadable-thread"), Message: "another bad row", SourceVersion: "opaque-added",
	})
	if err := verifyThreadSourceSnapshot(unreadable, added); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("unreadable addition error = %v", err)
	}
}
