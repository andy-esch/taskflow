package store

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// renameRepo seeds a scratch tree: an epic, task A (id-old.md) with an H1, and task B that
// links to A via a relative-path markdown link (display text == A's slug).
func renameRepo(t *testing.T) (root, aPath, bPath string) {
	t.Helper()
	root = t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("epics/01-e.md", "---\nstatus: active\npriority: high\ndescription: e\n---\n# E\n")
	aPath = filepath.Join(root, "tasks", "6fjangd7kva1-old.md")
	write("tasks/6fjangd7kva1-old.md", "---\nid: 6fjangd7kva1\nstatus: ready-to-start\nepic: 01-e\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\n---\n# Old Title\n\nbody.\n")
	bPath = filepath.Join(root, "tasks", "6fjangd7kvb2-b.md")
	write("tasks/6fjangd7kvb2-b.md", "---\nid: 6fjangd7kvb2\nstatus: ready-to-start\nepic: 01-e\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\n---\n# B\n\nSee [old](6fjangd7kva1-old.md#h) for context.\n")
	return root, aPath, bPath
}

func TestRenameTask_RenamesAndCascades(t *testing.T) {
	root, aPath, bPath := renameRepo(t)
	newPath := filepath.Join(root, "tasks", "6fjangd7kva1-shiny-new-title.md")

	result, err := NewFS(root).RenameTask("old", "Shiny New Title", false)
	if err != nil {
		t.Fatal(err)
	}
	task, cascade := result.Task, result.PlannedLinks
	if task.Slug != "shiny-new-title" || task.ID != "6fjangd7kva1" {
		t.Errorf("renamed task = %q id=%q, want shiny-new-title / same id", task.Slug, task.ID)
	}
	if cascade != 1 {
		t.Errorf("cascade count = %d, want 1 (task B's inbound link)", cascade)
	}
	if !result.Changed || !result.Committed || !result.Complete || !result.DestinationWritten || !result.SourceRemoved ||
		result.PlannedDocuments != 2 || result.AppliedDocuments != 2 || result.AppliedLinks != cascade {
		t.Errorf("rename result does not describe the complete write set: %+v", result)
	}
	// The file is renamed (id kept), the old name is gone.
	if _, err := os.Stat(aPath); !os.IsNotExist(err) {
		t.Error("old file should be removed")
	}
	got, _ := os.ReadFile(newPath)
	if !contains(got, "# Shiny New Title") {
		t.Errorf("body H1 not re-titled:\n%s", got)
	}
	// Task B's inbound link is repointed — filename, display text, AND the #anchor preserved.
	b, _ := os.ReadFile(bPath)
	if !contains(b, "[shiny-new-title](6fjangd7kva1-shiny-new-title.md#h)") {
		t.Errorf("inbound link not cascaded (filename/display/anchor):\n%s", b)
	}
}

func TestRenameTask_PreservesSourceFileMode(t *testing.T) {
	root, oldPath, _ := renameRepo(t)
	if err := os.Chmod(oldPath, 0o600); err != nil {
		t.Fatal(err)
	}
	newPath := filepath.Join(root, "tasks", "6fjangd7kva1-new-title.md")

	result, err := NewFS(root).RenameTask("old", "New title", false)
	if err != nil || !result.Complete {
		t.Fatalf("restricted-mode rename = %+v, %v", result, err)
	}
	info, err := os.Stat(newPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("renamed task mode = %#o, want source mode 0600", got)
	}
}

// TestRenameTask_CrossDirSameNameLeftAlone: a link is repointed only if it RESOLVES to the
// renamed file — a same-basename file in a different directory (and links to it) are left
// untouched (the cross-dir-collision fix).
func TestRenameTask_CrossDirSameNameLeftAlone(t *testing.T) {
	root, _, _ := renameRepo(t)
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A DIFFERENT file sharing task A's basename, and a task C linking to THAT one.
	write("research/6fjangd7kva1-old.md", "# a research doc\n")
	cPath := "tasks/6fjangd7kvc3-c.md"
	write(cPath, "---\nid: 6fjangd7kvc3\nstatus: ready-to-start\nepic: 01-e\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\n---\n# C\n\nElsewhere: [old](../research/6fjangd7kva1-old.md).\n")

	result, err := NewFS(root).RenameTask("old", "New Title", false)
	if err != nil {
		t.Fatal(err)
	}
	cascade := result.PlannedLinks
	// Only task B's link (to tasks/…old.md) cascades; task C's link to research/…old.md doesn't.
	if cascade != 1 {
		t.Errorf("only a link to the ACTUAL renamed file should cascade, got %d", cascade)
	}
	c, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(cPath)))
	if !contains(c, "../research/6fjangd7kva1-old.md") {
		t.Errorf("a same-named file in another dir must be left untouched:\n%s", c)
	}
	if _, err := os.Stat(filepath.Join(root, "research", "6fjangd7kva1-old.md")); err != nil {
		t.Error("the cross-dir same-name file must not be renamed/removed")
	}
}

func TestRenameTask_DryRunTouchesNothing(t *testing.T) {
	root, aPath, bPath := renameRepo(t)
	aBefore, _ := os.ReadFile(aPath)
	bBefore, _ := os.ReadFile(bPath)

	result, err := NewFS(root).RenameTask("old", "New", true)
	if err != nil {
		t.Fatal(err)
	}
	task, cascade := result.Task, result.PlannedLinks
	if task.Slug != "new" || cascade != 1 {
		t.Errorf("dry run should still report the would-be result: slug=%q cascade=%d", task.Slug, cascade)
	}
	if !result.DryRun || result.Committed || result.Complete || result.AppliedDocuments != 0 || result.PlannedDocuments != 2 {
		t.Errorf("dry-run result must describe a non-durable preview: %+v", result)
	}
	if _, err := os.Stat(aPath); err != nil {
		t.Error("dry run must not remove the old file")
	}
	if a, _ := os.ReadFile(aPath); !equal(a, aBefore) {
		t.Error("dry run modified the target file")
	}
	if b, _ := os.ReadFile(bPath); !equal(b, bBefore) {
		t.Error("dry run modified an inbound-link file")
	}
}

func TestRenameTask_SameSlugDoesNotCascadeOrRemoveSource(t *testing.T) {
	root, oldPath, bPath := renameRepo(t)
	bBefore, _ := os.ReadFile(bPath)

	result, err := NewFS(root).RenameTask("old", "Old", false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Task.Slug != "old" || result.PlannedLinks != 0 || result.AppliedLinks != 0 ||
		!result.Committed || !result.Complete || !result.DestinationWritten || result.SourceRemoved ||
		result.PlannedDocuments != 1 || result.AppliedDocuments != 1 {
		t.Fatalf("same-slug rename result = %+v", result)
	}
	if body, _ := os.ReadFile(oldPath); !contains(body, "# Old\n") {
		t.Fatalf("same-slug rename did not update the title in place:\n%s", body)
	}
	if body, _ := os.ReadFile(bPath); !equal(body, bBefore) {
		t.Fatalf("same-slug rename changed an inbound link unnecessarily:\n%s", body)
	}
}

func TestRenameTask_ExactRetitleNoOpDoesNotWrite(t *testing.T) {
	root, oldPath, bPath := renameRepo(t)
	body, err := os.ReadFile(oldPath)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.Replace(body, []byte("# Old Title\n"), []byte("# Old\n"), 1)
	if err := os.WriteFile(oldPath, body, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(oldPath, 0o600); err != nil {
		t.Fatal(err)
	}
	bBefore, _ := os.ReadFile(bPath)

	result, err := NewFS(root).RenameTask("old", "Old", false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Changed || result.Committed || !result.Complete || result.DestinationWritten || result.SourceRemoved ||
		result.PlannedDocuments != 0 || result.AppliedDocuments != 0 || result.PlannedLinks != 0 {
		t.Fatalf("exact rename no-op result = %+v", result)
	}
	if got, _ := os.ReadFile(oldPath); !equal(got, body) {
		t.Fatalf("exact rename no-op rewrote the source:\n%s", got)
	}
	if info, err := os.Stat(oldPath); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("exact rename no-op changed source mode: info=%v err=%v", info, err)
	}
	if got, _ := os.ReadFile(bPath); !equal(got, bBefore) {
		t.Fatalf("exact rename no-op rewrote an inbound document:\n%s", got)
	}
}

func TestRenameTask_EmptyTitleRejected(t *testing.T) {
	root, _, _ := renameRepo(t)
	if _, err := NewFS(root).RenameTask("old", "…", false); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("a title that slugifies to empty must be ErrValidation, got %v", err)
	}
}

// TestRenameTask_TargetCollisionRefused: renaming onto a filename that already exists must
// fail loud (ErrConflict) rather than silently clobber it — the write loop would otherwise
// overwrite the pre-existing same-id file. Nothing on disk changes.
func TestRenameTask_TargetCollisionRefused(t *testing.T) {
	root, aPath, _ := renameRepo(t)
	// A stray file occupying task A's would-be target name (same id, different slug).
	takenPath := filepath.Join(root, "tasks", "6fjangd7kva1-taken.md")
	if err := os.WriteFile(takenPath, []byte("# do not clobber\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	aBefore, _ := os.ReadFile(aPath)

	if _, err := NewFS(root).RenameTask("old", "Taken", false); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("rename onto an existing target must be ErrConflict, got %v", err)
	}
	// The source is untouched and the pre-existing target is intact.
	if a, _ := os.ReadFile(aPath); !equal(a, aBefore) {
		t.Error("source file must be unchanged on a refused rename")
	}
	if got, _ := os.ReadFile(takenPath); !contains(got, "do not clobber") {
		t.Error("pre-existing target file must not be overwritten")
	}
}

// TestRenameTask_RefStyleAndFencedExamples: the cascade repoints a reference-style link to
// the renamed file but leaves an inline example link inside a fenced code block untouched.
func TestRenameTask_RefStyleAndFencedExamples(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("epics/01-e.md", "---\nstatus: active\npriority: high\ndescription: e\n---\n# E\n")
	write("tasks/6fjangd7kva1-old.md", "---\nid: 6fjangd7kva1\nstatus: ready-to-start\nepic: 01-e\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\n---\n# Old Title\n\nbody.\n")
	bPath := filepath.Join(root, "tasks", "6fjangd7kvb2-b.md")
	write("tasks/6fjangd7kvb2-b.md", "---\nid: 6fjangd7kvb2\nstatus: ready-to-start\nepic: 01-e\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\n---\n# B\n\n"+
		"Ref: see the [old task][a].\n\n[a]: 6fjangd7kva1-old.md\n\nExample:\n\n```\n[old](6fjangd7kva1-old.md)\n```\n")

	result, err := NewFS(root).RenameTask("old", "Renamed", false)
	if err != nil {
		t.Fatal(err)
	}
	cascade := result.PlannedLinks
	if cascade != 1 {
		t.Errorf("only the reference-style link should cascade (fenced example skipped), got %d", cascade)
	}
	b, _ := os.ReadFile(bPath)
	if !contains(b, "[a]: 6fjangd7kva1-renamed.md") {
		t.Errorf("reference-style link target not repointed:\n%s", b)
	}
	if !contains(b, "[old](6fjangd7kva1-old.md)") {
		t.Errorf("an inline example inside a code fence must be left untouched:\n%s", b)
	}
}

func TestRenameTask_ConcurrentRenamesCommitOneSourceSnapshot(t *testing.T) {
	root, oldPath, _ := renameRepo(t)
	ready := make(chan struct{}, 2)
	release := make(chan struct{})
	testHookBeforeTaskRenameLock = func() {
		ready <- struct{}{}
		<-release
	}
	defer func() { testHookBeforeTaskRenameLock = nil }()

	type outcome struct {
		result core.TaskRenameMutationResult
		err    error
	}
	outcomes := make(chan outcome, 2)
	for _, title := range []string{"Alpha title", "Beta title"} {
		title := title
		go func() {
			result, err := NewFS(root).RenameTask("old", title, false)
			outcomes <- outcome{result: result, err: err}
		}()
	}
	<-ready
	<-ready
	close(release)

	var successes, conflicts int
	for range 2 {
		outcome := <-outcomes
		switch {
		case outcome.err == nil:
			successes++
			if !outcome.result.Complete || !outcome.result.Committed {
				t.Errorf("successful concurrent rename lacks complete receipt: %+v", outcome.result)
			}
		case errors.Is(outcome.err, domain.ErrConflict):
			conflicts++
			if outcome.result.Committed {
				t.Errorf("losing concurrent rename reported a durable write: %+v", outcome.result)
			}
		default:
			t.Fatalf("unexpected concurrent rename error: %v", outcome.err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent renames successes=%d conflicts=%d, want one each", successes, conflicts)
	}
	matches, err := filepath.Glob(filepath.Join(root, "tasks", "6fjangd7kva1-*.md"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("concurrent rename left %d stable-id owners: %v (err=%v)", len(matches), matches, err)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("the winning rename did not retire the old path: %v", err)
	}
}

func TestRenameTask_ReplansAfterCooperatingCascadeDocumentWrite(t *testing.T) {
	root, _, bPath := renameRepo(t)
	ready := make(chan struct{})
	release := make(chan struct{})
	testHookBeforeTaskRenameLock = func() {
		close(ready)
		<-release
	}
	defer func() { testHookBeforeTaskRenameLock = nil }()

	type outcome struct {
		result core.TaskRenameMutationResult
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		result, err := NewFS(root).RenameTask("old", "New title", false)
		done <- outcome{result: result, err: err}
	}()
	<-ready
	if _, err := NewFS(root).SetFields("b", map[string]any{"priority": "low"}, false); err != nil {
		t.Fatal(err)
	}
	close(release)
	got := <-done
	if got.err != nil || !got.result.Complete {
		t.Fatalf("rename after cooperating write = %+v, %v", got.result, got.err)
	}

	task, body, err := NewFS(root).GetTask("b")
	if err != nil {
		t.Fatal(err)
	}
	if task.Priority != "low" {
		t.Fatalf("rename discarded the cooperating priority write: %+v", task)
	}
	if !strings.Contains(body, "6fjangd7kva1-new-title.md") {
		t.Fatalf("rename did not cascade through the refreshed document:\n%s", body)
	}
	if _, err := os.Stat(bPath); err != nil {
		t.Fatalf("cascade document moved unexpectedly: %v", err)
	}
}

func TestRenameTask_RejectsSourceChangedWhileWaitingBeforeCascadeWrites(t *testing.T) {
	root, oldPath, bPath := renameRepo(t)
	bBefore, _ := os.ReadFile(bPath)
	ready := make(chan struct{})
	release := make(chan struct{})
	testHookBeforeTaskRenameLock = func() {
		close(ready)
		<-release
	}
	defer func() { testHookBeforeTaskRenameLock = nil }()

	type renameOutcome struct {
		result core.TaskRenameMutationResult
		err    error
	}
	done := make(chan renameOutcome, 1)
	go func() {
		result, err := NewFS(root).RenameTask("old", "New title", false)
		done <- renameOutcome{result: result, err: err}
	}()
	<-ready
	if _, err := NewFS(root).SetFields("old", map[string]any{"priority": "low"}, false); err != nil {
		t.Fatal(err)
	}
	close(release)
	got := <-done
	if !errors.Is(got.err, domain.ErrConflict) || got.result.Committed || got.result.AppliedDocuments != 0 {
		t.Fatalf("rename over a changed source = %+v, %v; want an uncommitted conflict", got.result, got.err)
	}
	task, _, err := NewFS(root).GetTask("old")
	if err != nil || task.Priority != "low" {
		t.Fatalf("guarded source edit was not preserved: %+v, %v", task, err)
	}
	if body, _ := os.ReadFile(bPath); !equal(body, bBefore) {
		t.Fatalf("stale source intent wrote a needless cascade prefix:\n%s", body)
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Fatalf("stale source intent removed the source: %v", err)
	}
}

func TestRenameTask_RechecksTargetAfterRepositoryGuard(t *testing.T) {
	root, oldPath, _ := renameRepo(t)
	target := filepath.Join(root, "tasks", "6fjangd7kva1-new-title.md")
	oldBefore, _ := os.ReadFile(oldPath)
	testHookBeforeTaskRenameLock = func() {
		if err := os.WriteFile(target, []byte("# raw target\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	defer func() { testHookBeforeTaskRenameLock = nil }()

	result, err := NewFS(root).RenameTask("old", "New title", false)
	if !errors.Is(err, domain.ErrConflict) || result.Committed {
		t.Fatalf("post-lock target collision = %+v, %v; want uncommitted conflict", result, err)
	}
	if body, _ := os.ReadFile(oldPath); !equal(body, oldBefore) {
		t.Fatal("target collision changed the rename source")
	}
	if body, _ := os.ReadFile(target); string(body) != "# raw target\n" {
		t.Fatalf("target collision clobbered the raw target:\n%s", body)
	}
}

func TestRenameTask_ExclusiveDestinationCreateRejectsPostCASTarget(t *testing.T) {
	root, oldPath, bPath := renameRepo(t)
	if err := os.Remove(bPath); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "tasks", "6fjangd7kva1-new-title.md")
	const raced = "# raw target after source CAS\n"
	testHookBeforeTaskRenameDestinationCreate = func(path string) {
		if path != target {
			t.Fatalf("destination-create hook path = %s, want %s", path, target)
		}
		if err := os.WriteFile(path, []byte(raced), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	defer func() { testHookBeforeTaskRenameDestinationCreate = nil }()

	result, err := NewFS(root).RenameTask("old", "New title", false)
	if !errors.Is(err, domain.ErrConflict) || result.Committed || result.DestinationWritten {
		t.Fatalf("post-CAS target race = %+v, %v; want an uncommitted conflict", result, err)
	}
	if body, _ := os.ReadFile(target); string(body) != raced {
		t.Fatalf("exclusive destination create clobbered the raced target:\n%s", body)
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Fatalf("exclusive destination conflict removed the source: %v", err)
	}
}

func TestRenameTask_ExclusiveDestinationCreatePreservesDanglingSymlink(t *testing.T) {
	root, oldPath, bPath := renameRepo(t)
	if err := os.Remove(bPath); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "tasks", "6fjangd7kva1-new-title.md")
	if err := os.Symlink(filepath.Join(root, "missing-target"), target); err != nil {
		t.Fatal(err)
	}

	result, err := NewFS(root).RenameTask("old", "New title", false)
	if !errors.Is(err, domain.ErrConflict) || result.Committed || result.DestinationWritten {
		t.Fatalf("dangling target symlink = %+v, %v; want an uncommitted conflict", result, err)
	}
	info, err := os.Lstat(target)
	if err != nil {
		t.Fatalf("exclusive destination create removed the dangling symlink: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("exclusive destination create replaced the dangling symlink: mode=%v", info.Mode())
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Fatalf("dangling target conflict removed the source: %v", err)
	}
}

func TestRenameTask_CASCatchesRawCascadeDocumentEdit(t *testing.T) {
	root, oldPath, bPath := renameRepo(t)
	const raced = "---\nid: 6fjangd7kvb2\nstatus: ready-to-start\nepic: 01-e\npriority: low\n---\n# B raced\n\nSee [old](6fjangd7kva1-old.md#h).\n"
	mutated := false
	testHookBeforeTaskRenameWrite = func(path string) {
		if path == bPath && !mutated {
			mutated = true
			if err := os.WriteFile(path, []byte(raced), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	defer func() { testHookBeforeTaskRenameWrite = nil }()

	result, err := NewFS(root).RenameTask("old", "New title", false)
	if !errors.Is(err, domain.ErrConflict) || result.Committed {
		t.Fatalf("raw cascade race = %+v, %v; want uncommitted conflict", result, err)
	}
	if body, _ := os.ReadFile(bPath); string(body) != raced {
		t.Fatalf("rename clobbered the raw cascade edit:\n%s", body)
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Fatalf("conflicted rename removed its source: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "tasks", "6fjangd7kva1-new-title.md")); !os.IsNotExist(err) {
		t.Fatalf("conflicted rename created its destination: %v", err)
	}
}

func TestRenameTask_PreDestinationCASCatchesRawSourceEdit(t *testing.T) {
	root, oldPath, bPath := renameRepo(t)
	target := filepath.Join(root, "tasks", "6fjangd7kva1-new-title.md")
	const raced = "---\nid: 6fjangd7kva1\nstatus: ready-to-start\nepic: 01-e\n---\n# Raw source edit before destination\n"
	mutated := false
	testHookBeforeTaskRenameWrite = func(path string) {
		if path == oldPath && !mutated {
			mutated = true
			if err := os.WriteFile(path, []byte(raced), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	defer func() { testHookBeforeTaskRenameWrite = nil }()

	result, err := NewFS(root).RenameTask("old", "New title", false)
	if !errors.Is(err, domain.ErrConflict) || !result.Committed || result.DestinationWritten ||
		result.AppliedDocuments != 1 || result.AppliedLinks != 1 {
		t.Fatalf("raw pre-destination source race = %+v, %v", result, err)
	}
	if body, _ := os.ReadFile(oldPath); string(body) != raced {
		t.Fatalf("pre-destination CAS discarded the raw source edit:\n%s", body)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("pre-destination CAS allowed a stale destination: %v", err)
	}
	if body, _ := os.ReadFile(bPath); !contains(body, "6fjangd7kva1-new-title.md") {
		t.Fatalf("receipt did not match its durable cascade prefix:\n%s", body)
	}
}

func TestRenameTask_PartialCascadeReceiptIsResumable(t *testing.T) {
	root, oldPath, bPath := renameRepo(t)
	cPath := filepath.Join(root, "tasks", "6fjangd7kvc3-c.md")
	if err := os.WriteFile(cPath, []byte("---\nid: 6fjangd7kvc3\nstatus: ready-to-start\nepic: 01-e\n---\n# C\n\nSee [old](6fjangd7kva1-old.md).\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testHookAfterTaskRenameWrite = func(path string) error {
		if path == bPath {
			return errors.New("injected cascade interruption")
		}
		return nil
	}

	receipt, err := core.NewService(NewFS(root)).RenameTask("old", "New title", false)
	var partial *core.TaskRenameFailure
	if !errors.As(err, &partial) || !receipt.Committed || receipt.Complete || receipt.DestinationWritten ||
		receipt.AppliedDocuments != 1 || receipt.PlannedDocuments != 3 || receipt.AppliedLinks != 1 || receipt.PlannedLinks != 2 {
		t.Fatalf("partial cascade receipt = %+v, err=%v", receipt, err)
	}
	if partial.Receipt.Remedy == "" || !strings.Contains(err.Error(), "rerun") {
		t.Fatalf("partial cascade lacks recovery guidance: %+v, %v", partial.Receipt, err)
	}
	if body, _ := os.ReadFile(bPath); !contains(body, "6fjangd7kva1-new-title.md") {
		t.Fatalf("durable prefix was not reported accurately:\n%s", body)
	}
	if body, _ := os.ReadFile(cPath); !contains(body, "6fjangd7kva1-old.md") {
		t.Fatalf("unapplied suffix changed unexpectedly:\n%s", body)
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Fatalf("resumable prefix removed the source: %v", err)
	}

	testHookAfterTaskRenameWrite = nil
	completed, err := core.NewService(NewFS(root)).RenameTask("6fjangd7kva1", "New title", false)
	if err != nil || !completed.Complete || !completed.Committed {
		t.Fatalf("retry did not converge: %+v, %v", completed, err)
	}
	for _, path := range []string{bPath, cPath} {
		if body, _ := os.ReadFile(path); !contains(body, "6fjangd7kva1-new-title.md") {
			t.Errorf("retry left an old link in %s:\n%s", path, body)
		}
	}
}

func TestRenameTask_DestinationWrittenCleanupFailureRequiresInspection(t *testing.T) {
	root, oldPath, _ := renameRepo(t)
	newPath := filepath.Join(root, "tasks", "6fjangd7kva1-new-title.md")
	testHookBeforeTaskRenameSourceRemove = func(_, _ string) error {
		return errors.New("injected source cleanup failure")
	}
	defer func() { testHookBeforeTaskRenameSourceRemove = nil }()

	receipt, err := core.NewService(NewFS(root)).RenameTask("old", "New title", false)
	var partial *core.TaskRenameFailure
	if !errors.As(err, &partial) || !receipt.Committed || receipt.Complete || !receipt.DestinationWritten || receipt.SourceRemoved {
		t.Fatalf("destination-written receipt = %+v, err=%v", receipt, err)
	}
	if !strings.Contains(receipt.Remedy, "old and destination") || !strings.Contains(err.Error(), "before retrying") {
		t.Fatalf("cleanup failure does not require inspection: %+v, %v", receipt, err)
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Fatalf("cleanup failure did not retain old source: %v", err)
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("cleanup failure did not report the durable destination: %v", err)
	}
}

func TestRenameTask_SourceRemovalCASCatchesRawEdit(t *testing.T) {
	root, oldPath, _ := renameRepo(t)
	newPath := filepath.Join(root, "tasks", "6fjangd7kva1-new-title.md")
	const raced = "---\nid: 6fjangd7kva1\nstatus: ready-to-start\nepic: 01-e\n---\n# Raw edit after destination write\n"
	testHookBeforeTaskRenameSourceRemove = func(_, _ string) error {
		return os.WriteFile(oldPath, []byte(raced), 0o644)
	}
	defer func() { testHookBeforeTaskRenameSourceRemove = nil }()

	receipt, err := core.NewService(NewFS(root)).RenameTask("old", "New title", false)
	var partial *core.TaskRenameFailure
	if !errors.As(err, &partial) || !errors.Is(err, domain.ErrConflict) ||
		!receipt.Committed || receipt.Complete || !receipt.DestinationWritten || receipt.SourceRemoved {
		t.Fatalf("source-removal CAS receipt = %+v, err=%v", receipt, err)
	}
	if body, _ := os.ReadFile(oldPath); string(body) != raced {
		t.Fatalf("source-removal CAS discarded the raw edit:\n%s", body)
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("source-removal CAS did not retain the durable destination: %v", err)
	}
}

func TestRenameTask_CompleteUnlockFailureIsNotRetryable(t *testing.T) {
	root, oldPath, _ := renameRepo(t)
	testHookRepositoryUnlockError = func() error { return errors.New("injected unlock failure") }
	defer func() { testHookRepositoryUnlockError = nil }()

	receipt, err := core.NewService(NewFS(root)).RenameTask("old", "New title", false)
	var committed *core.TaskRenameFailure
	if !errors.As(err, &committed) || !receipt.Committed || !receipt.Complete {
		t.Fatalf("completed unlock failure = %+v, %v", receipt, err)
	}
	if !strings.Contains(receipt.Remedy, "already complete") {
		t.Fatalf("completed failure suggests an unsafe retry: %+v", receipt)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("completed rename retained its old path: %v", err)
	}
}

func TestRenameTask_SuccessHasNoRecoveryRemedy(t *testing.T) {
	root, _, _ := renameRepo(t)
	receipt, err := core.NewService(NewFS(root)).RenameTask("old", "New title", false)
	if err != nil || !receipt.Complete || receipt.Remedy != "" {
		t.Fatalf("successful rename receipt = %+v, %v; want no recovery remedy", receipt, err)
	}
}

func contains(b []byte, sub string) bool { return strings.Contains(string(b), sub) }
func equal(a, b []byte) bool             { return bytes.Equal(a, b) }
