package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/testutil"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestCreateTask_OrderQuotingClobber(t *testing.T) {
	fs := NewFS(t.TempDir())
	task := domain.Task{
		Slug: "demo", ID: "0abcdef12345", Status: domain.StatusReadyToStart, Epic: "e1",
		Description: "has a colon: yes", Effort: "Unknown", Tier: 3,
		Priority: "medium", Autonomy: 3, Tags: []string{"a", "b"}, Created: "2026-06-08",
	}

	got, err := fs.CreateTask(task, "\n# Demo\n", false)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(got.Path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)

	// Canonical order: status before epic before description.
	si, ei, di := strings.Index(s, "status:"), strings.Index(s, "epic:"), strings.Index(s, "description:")
	if si < 0 || si >= ei || ei >= di {
		t.Errorf("frontmatter not in canonical order:\n%s", s)
	}
	// A colon in the description must be quoted (the pm non-conformant-YAML trap);
	// yaml.v3 uses single quotes, which round-trip fine.
	if !strings.Contains(s, "'has a colon: yes'") {
		t.Errorf("description with colon not quoted:\n%s", s)
	}
	// And it must re-parse — proving valid YAML was written.
	if _, _, err := fs.GetTask("demo"); err != nil {
		t.Errorf("created task does not re-parse: %v", err)
	}
	// Clobber refused with a conflict (not a generic validation error).
	if _, err := fs.CreateTask(task, "x", false); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("clobber should be ErrConflict, got %v", err)
	}
}

func TestCreateTaskCreatesMissingPlanningRootBeforeLocking(t *testing.T) {
	root := filepath.Join(t.TempDir(), "new-planning-root")
	fs := NewFS(root)
	task := domain.Task{
		Slug: "first", ID: "0abcdef12345", Status: domain.StatusReadyToStart, Epic: "e1",
		Description: "first task", Effort: "Unknown", Tier: 3,
		Priority: "medium", Autonomy: 3, Tags: []string{"a"}, Created: "2026-08-27",
	}
	got, err := fs.CreateTask(task, "# First\n", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(got.Path); err != nil {
		t.Fatalf("created task path: %v", err)
	}
}

func TestCreateEntityFileSerializesPreparationWithWrite(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "entities")
	firstPrepared := make(chan struct{})
	releaseFirst := make(chan struct{})
	secondPrepared := make(chan struct{})
	results := make(chan error, 2)

	go func() {
		_, err := NewFS(root).createEntityFile(false, func() (entityFileCreation, error) {
			close(firstPrepared)
			<-releaseFirst
			return entityFileCreation{
				dir: dir, path: filepath.Join(dir, "first.md"), content: []byte("first"), kind: "test entity", name: "first",
			}, nil
		})
		results <- err
	}()
	<-firstPrepared

	go func() {
		_, err := NewFS(root).createEntityFile(false, func() (entityFileCreation, error) {
			close(secondPrepared)
			return entityFileCreation{
				dir: dir, path: filepath.Join(dir, "second.md"), content: []byte("second"), kind: "test entity", name: "second",
			}, nil
		})
		results <- err
	}()

	select {
	case <-secondPrepared:
		t.Fatal("a concurrent entity prepared while the first create still held the repository guard")
	case <-time.After(50 * time.Millisecond):
	}
	close(releaseFirst)
	for range 2 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
}

func TestCreateTaskRefusesDuplicateIDAcrossDifferentSlugs(t *testing.T) {
	root := t.TempDir()
	const shared = "6g7s6hr3qnfq"
	newTask := func(slug string) domain.Task {
		return domain.Task{ID: shared, Slug: slug, Status: domain.StatusReadyToStart, Created: "2026-09-07"}
	}
	if _, err := NewFS(root).CreateTask(newTask("alpha"), "# Alpha\n", false); err != nil {
		t.Fatal(err)
	}
	_, err := NewFS(root).CreateTask(newTask("beta"), "# Beta\n", false)
	if !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "alpha") {
		t.Fatalf("duplicate task id error = %v, want conflict naming the existing owner", err)
	}
	if matches, _ := filepath.Glob(filepath.Join(root, domain.TasksDir, "*beta*")); len(matches) != 0 {
		t.Fatalf("refused create wrote %v", matches)
	}
}

func TestCreateTaskTreatsUnreadableFilenameIdentityAsOwned(t *testing.T) {
	root := t.TempDir()
	const shared = "6g7s6hr3qnfq"
	dir := filepath.Join(root, domain.TasksDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	owner := filepath.Join(dir, shared+"-broken-owner.md")
	if err := os.WriteFile(owner, []byte("# missing frontmatter\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := NewFS(root).CreateTask(domain.Task{
		ID: shared, Slug: "replacement", Status: domain.StatusReadyToStart, Created: "2026-09-07",
	}, "# Replacement\n", false)
	if !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "broken-owner") {
		t.Fatalf("unreadable owner collision = %v, want conflict naming the filename owner", err)
	}
}

func TestCreateTask_IDRoundTrips(t *testing.T) {
	fs := NewFS(t.TempDir())
	// Alphanumeric and all-digit ids: the latter must survive YAML as a string, not
	// be coerced to an int (which would drop the leading zero).
	for _, wantID := range []string{"0abcdef12345", "012345678901"} {
		task := domain.Task{
			Slug: "t-" + wantID, ID: wantID, Status: domain.StatusReadyToStart, Epic: "e1",
			Description: "d", Effort: "Unknown", Tier: 3, Priority: "medium",
			Autonomy: 3, Tags: []string{"a"}, Created: "2026-07-02",
		}
		got, err := fs.CreateTask(task, "\n# x\n", false)
		if err != nil {
			t.Fatalf("%s: %v", wantID, err)
		}
		b, err := os.ReadFile(got.Path)
		if err != nil {
			t.Fatal(err)
		}
		// id is written in canonical position: after schema, before status.
		sci, ii, sti := strings.Index(string(b), "schema:"), strings.Index(string(b), "id:"), strings.Index(string(b), "status:")
		if sci < 0 || sci >= ii || ii >= sti {
			t.Errorf("%s: id not in canonical position (schema<id<status):\n%s", wantID, b)
		}
		reparsed, _, err := fs.GetTask("t-" + wantID)
		if err != nil {
			t.Fatalf("%s: re-parse: %v", wantID, err)
		}
		if reparsed.ID != wantID {
			t.Errorf("id did not round-trip: got %q want %q", reparsed.ID, wantID)
		}
	}
}

func TestCreateAudit_OpenBucketOrderClobber(t *testing.T) {
	fs := NewFS(t.TempDir())
	a := domain.Audit{ID: "0abcdef45678", Slug: "2026-06-16-dispatcher", Area: "dispatcher", Date: "2026-06-16"}

	got, err := fs.CreateAudit(a, "\n# Audit\n", false)
	if err != nil {
		t.Fatal(err)
	}
	// Flat layout: the audit lives directly under audits/ (bucket is frontmatter, open).
	if base := filepath.Base(filepath.Dir(got.Path)); base != "audits" {
		t.Errorf("audit created under %q/, want audits/", base)
	}
	if got.Bucket != domain.AuditOpen {
		t.Errorf("created audit bucket = %q, want open", got.Bucket)
	}
	b, err := os.ReadFile(got.Path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	// Canonical frontmatter order: area before date.
	if ai, di := strings.Index(s, "area:"), strings.Index(s, "date:"); ai < 0 || ai >= di {
		t.Errorf("frontmatter not in canonical order (area before date):\n%s", s)
	}
	// And it re-parses through the store.
	if _, _, err := fs.GetAudit("2026-06-16-dispatcher"); err != nil {
		t.Errorf("created audit does not re-parse: %v", err)
	}
	// Clobber refused with a conflict.
	if _, err := fs.CreateAudit(a, "x", false); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("clobber should be ErrConflict, got %v", err)
	}
}

func TestCreateAudit_IDRoundTrips(t *testing.T) {
	fs := NewFS(t.TempDir())
	const wantID = "0abcdef12345"
	a := domain.Audit{Slug: "2026-07-02-x", ID: wantID, Area: "x", Date: "2026-07-02"}
	got, err := fs.CreateAudit(a, "\n# x\n", false)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(got.Path)
	if err != nil {
		t.Fatal(err)
	}
	// id in canonical position: after schema, before area.
	sci, ii, ai := strings.Index(string(b), "schema:"), strings.Index(string(b), "id:"), strings.Index(string(b), "area:")
	if sci < 0 || sci >= ii || ii >= ai {
		t.Errorf("id not in canonical position (schema<id<area):\n%s", b)
	}
	reparsed, _, err := fs.GetAudit("2026-07-02-x")
	if err != nil {
		t.Fatal(err)
	}
	if reparsed.ID != wantID {
		t.Errorf("audit id did not round-trip: got %q want %q", reparsed.ID, wantID)
	}
}

func TestCreateAudit_RefusesDuplicateIDAcrossDifferentSlugs(t *testing.T) {
	root := t.TempDir()
	fs := NewFS(root)
	const shared = "6g7s4k845fsb"
	alpha := domain.Audit{ID: shared, Slug: "2026-09-07-alpha", Area: "alpha", Date: "2026-09-07"}
	beta := domain.Audit{ID: shared, Slug: "2026-09-07-beta", Area: "beta", Date: "2026-09-07"}
	if _, err := fs.CreateAudit(alpha, "# Alpha\n", false); err != nil {
		t.Fatal(err)
	}

	_, err := fs.CreateAudit(beta, "# Beta\n", false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate audit id must be ErrConflict, got %v", err)
	}
	if !strings.Contains(err.Error(), "2026-09-07-alpha") {
		t.Errorf("error should name the existing owner: %v", err)
	}
	if matches, _ := filepath.Glob(filepath.Join(root, domain.AuditsDir, "*beta*")); len(matches) != 0 {
		t.Errorf("refused create wrote %v", matches)
	}

	if _, err := fs.CreateAudit(beta, "# Beta\n", true); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("dry-run must refuse the same duplicate id, got %v", err)
	}
}

func TestCreateAudit_SerializesDuplicateIDCheckWithCreate(t *testing.T) {
	root := t.TempDir()
	fs := NewFS(root)
	const shared = "6g7s4k845fsc"
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, slug := range []string{"2026-09-07-alpha", "2026-09-07-beta"} {
		audit := domain.Audit{ID: shared, Slug: slug, Area: slug, Date: "2026-09-07"}
		go func() {
			<-start
			_, err := fs.CreateAudit(audit, "# Audit\n", false)
			results <- err
		}()
	}
	close(start)

	var successes, conflicts int
	for range 2 {
		err := <-results
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrConflict):
			conflicts++
		default:
			t.Fatalf("unexpected create result: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d, want 1 each", successes, conflicts)
	}
	matches, err := filepath.Glob(filepath.Join(root, domain.AuditsDir, shared+"-*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("created %d audits for one stable id: %v", len(matches), matches)
	}
}

func TestCreateEpic_AutoNumber(t *testing.T) {
	fs := NewFS(t.TempDir())
	// First epic → 01; with an existing 04-... the next is 05.
	first, err := fs.CreateEpic("alpha", domain.Epic{Status: "active", Description: "d", Priority: "medium", Created: "2026-06-08"}, "\n# Alpha\n", false)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != "01-alpha" {
		t.Errorf("first epic id = %q, want 01-alpha", first.ID)
	}
	if err := os.WriteFile(fs.epicsDir+"/04-beta.md", []byte("---\nstatus: active\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	next, err := fs.CreateEpic("gamma", domain.Epic{Status: "active", Description: "d", Priority: "medium", Created: "2026-06-08"}, "\n# G\n", false)
	if err != nil {
		t.Fatal(err)
	}
	if next.ID != "05-gamma" {
		t.Errorf("next epic id = %q, want 05-gamma", next.ID)
	}
}

func TestCreateEpicSerializesNumberAllocation(t *testing.T) {
	root := t.TempDir()
	type result struct {
		epic domain.Epic
		err  error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	for _, slug := range []string{"alpha", "beta"} {
		slug := slug
		go func() {
			<-start
			epic, err := NewFS(root).CreateEpic(slug, domain.Epic{
				Status: "active", Description: slug, Priority: "medium", Created: "2026-09-07",
			}, "# "+slug+"\n", false)
			results <- result{epic: epic, err: err}
		}()
	}
	close(start)
	first, second := <-results, <-results
	if first.err != nil || second.err != nil {
		t.Fatalf("concurrent epic creates failed: %v, %v", first.err, second.err)
	}
	firstNum, secondNum := epicNum(first.epic.ID), epicNum(second.epic.ID)
	if firstNum == secondNum || firstNum+secondNum != 3 {
		t.Fatalf("concurrent epic ids = %q, %q; want distinct allocations 1 and 2", first.epic.ID, second.epic.ID)
	}
}

// The creates build `<id>-<slug>.md` straight from the id, so an illegal one would write a
// file the scanner then refuses to parse. Catch it at the write, not at the next lint.
func TestCreateRejectsAnInvalidEntityID(t *testing.T) {
	r := testutil.NewRepo(t)
	fs := NewFS(r.Root)
	if _, err := fs.CreateTask(domain.Task{ID: "6fbj87000lt6", Slug: "s"}, "# T\n", false); err == nil {
		t.Error("CreateTask accepted an invalid id")
	} else if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("CreateTask error = %v; want a validation failure", err)
	}
	if _, err := fs.CreateAudit(domain.Audit{ID: "6fbj87000ut6", Slug: "s", Area: "a", Date: "2026-01-01"}, "# A\n", false); err == nil {
		t.Error("CreateAudit accepted an invalid id")
	}
	if _, err := fs.CreateResearch(domain.Research{ID: "nope", Slug: "s"}, "# R\n", false); err == nil {
		t.Error("CreateResearch accepted an invalid id")
	}
	// Nothing was written.
	for _, dir := range []string{"tasks", "audits", "research"} {
		entries, _ := os.ReadDir(filepath.Join(r.Root, dir))
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".md") {
				t.Errorf("%s/%s was created despite an invalid id", dir, e.Name())
			}
		}
	}
}

func TestCreateTaskRejectsLifecycleOwnedAndInvalidStatuses(t *testing.T) {
	for _, status := range []domain.Status{
		domain.StatusInProgress, domain.StatusCompleted, domain.StatusDeferred,
		domain.StatusDeprecated, domain.Status("invented"), "",
	} {
		t.Run(string(status), func(t *testing.T) {
			fs := NewFS(t.TempDir())
			_, err := fs.CreateTask(domain.Task{
				ID: testutil.TaskID("ordinary-create-" + string(status)), Slug: "task",
				Status: status, Tags: []string{"test"}, Created: "2026-08-29",
			}, "# Task\n", false)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("CreateTask status %q error = %v, want validation", status, err)
			}
		})
	}
}
