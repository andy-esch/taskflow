package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func seedFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func frontmatterID(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fm, _ := splitFrontmatter(b)
	var m struct {
		ID string `yaml:"id"`
	}
	if err := yaml.Unmarshal(fm, &m); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return m.ID
}

func changesHave(changes []string, want string) bool {
	for _, c := range changes {
		if c == want {
			return true
		}
	}
	return false
}

const idlessTask = "---\nstatus: ready-to-start\nepic: e1\ntier: 2\npriority: high\neffort: 2h\ncreated: 2026-01-05\ntags: [a]\n---\n# T\n"

// TestFixFrontmatter_BackfillsMissingTaskID: a task with no id gets a valid one,
// timestamped from its created date.
func TestFixFrontmatter_BackfillsMissingTaskID(t *testing.T) {
	root := t.TempDir()
	p, out := testutil.TaskFixture(root, "ready-to-start", "t.md", idlessTask)
	seedFile(t, p, out)

	results, err := NewFS(root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !changesHave(results[0].Changes, "id: assigned (was missing)") {
		t.Fatalf("expected one id-assigned change, got %+v", results)
	}
	got := frontmatterID(t, p)
	if !id.Valid(got) {
		t.Errorf("backfilled id %q is not a valid id", got)
	}
	// The id's timestamp derives from created, not from now.
	if d := id.Time(got).UTC().Format("2006-01-02"); d != "2026-01-05" {
		t.Errorf("backfilled id time = %s, want 2026-01-05 (the created date)", d)
	}
}

// TestFixFrontmatter_BackfillsMissingAuditID: audits derive the id from their slug
// date (they carry no `created`).
func TestFixFrontmatter_BackfillsMissingAuditID(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, domain.AuditsDir, "open", "2026-01-02-x.md")
	seedFile(t, p, "---\narea: x\ndate: 2026-01-02\n---\n#### H1. t  · **Status:** open\n")

	if _, err := NewFS(root).FixFrontmatter(false); err != nil {
		t.Fatal(err)
	}
	got := frontmatterID(t, p)
	if !id.Valid(got) {
		t.Errorf("audit id %q is not valid", got)
	}
	if d := id.Time(got).UTC().Format("2006-01-02"); d != "2026-01-02" {
		t.Errorf("audit id time = %s, want 2026-01-02 (the slug date)", d)
	}
}

// TestFixFrontmatter_KeepsExistingID: a file that already has an id is untouched.
func TestFixFrontmatter_KeepsExistingID(t *testing.T) {
	root := t.TempDir()
	p, out := testutil.TaskFixture(root, "ready-to-start", "t.md", "---\nid: 6fjangd7kvh0\nstatus: ready-to-start\nepic: e1\ntier: 2\npriority: high\neffort: 2h\ncreated: 2026-01-05\ntags: [a]\n---\n# T\n")
	seedFile(t, p, out)

	results, err := NewFS(root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("a task that already has an id needs no fix, got %+v", results)
	}
	if got := frontmatterID(t, p); got != "6fjangd7kvh0" {
		t.Errorf("existing id must be preserved, got %q", got)
	}
}

// TestFixFrontmatter_SkipsBackfillWithoutDate: with no date to derive a timestamp
// from, the fix leaves the id missing (the re-lint after --fix re-flags it) rather
// than invent one.
func TestFixFrontmatter_SkipsBackfillWithoutDate(t *testing.T) {
	root := t.TempDir()
	p, out := testutil.TaskFixture(root, "ready-to-start", "t.md", "---\nstatus: ready-to-start\nepic: e1\ntier: 2\npriority: high\neffort: 2h\ntags: [a]\n---\n# T\n")
	seedFile(t, p, out)

	results, err := NewFS(root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if changesHave(r.Changes, "id: assigned (was missing)") {
			t.Errorf("a dateless task must not be backfilled: %+v", r)
		}
	}
	if got := frontmatterID(t, p); got != "" {
		t.Errorf("expected no id assigned, got %q", got)
	}
}

// TestFixFrontmatter_DistinctIDsForSameDate: two id-less tasks that share a created
// date still get distinct ids — the run-level dedup covers the 17-bit-tail collision
// NewAt alone could produce.
func TestFixFrontmatter_DistinctIDsForSameDate(t *testing.T) {
	root := t.TempDir()
	pa, outA := testutil.TaskFixture(root, "ready-to-start", "a.md", idlessTask)
	pb, outB := testutil.TaskFixture(root, "ready-to-start", "b.md", idlessTask)
	seedFile(t, pa, outA)
	seedFile(t, pb, outB)

	if _, err := NewFS(root).FixFrontmatter(false); err != nil {
		t.Fatal(err)
	}
	a := frontmatterID(t, pa)
	b := frontmatterID(t, pb)
	if a == "" || b == "" {
		t.Fatalf("both tasks should get ids, got %q and %q", a, b)
	}
	if a == b {
		t.Errorf("same-date tasks must get distinct ids, both = %s", a)
	}
}

// TestFixFrontmatter_BackfillsFromLifecycleStamp: an archived task carrying only a
// completed_at (no created/updated_at) still gets an id, dated from that stamp —
// the fallback that predated tasks like taskflow-00 would otherwise miss.
func TestFixFrontmatter_BackfillsFromLifecycleStamp(t *testing.T) {
	root := t.TempDir()
	p, out := testutil.TaskFixture(root, "completed", "old.md", "---\nstatus: completed\nepic: e1\ntier: 1\npriority: high\ncompleted_at: 2026-02-04\ntags: [x]\n---\n# Old\n")
	seedFile(t, p, out)

	if _, err := NewFS(root).FixFrontmatter(false); err != nil {
		t.Fatal(err)
	}
	got := frontmatterID(t, p)
	if !id.Valid(got) {
		t.Errorf("archived task id %q is not valid", got)
	}
	if d := id.Time(got).UTC().Format("2006-01-02"); d != "2026-02-04" {
		t.Errorf("id time = %s, want 2026-02-04 (from completed_at)", d)
	}
}

func TestDateFromFilename(t *testing.T) {
	cases := []struct{ name, want string }{
		{"2025-10-19-slug.md", "2025-10-19"},
		{"2026-01-05-x.md", "2026-01-05"},
		{"2025-10-19.md", "2025-10-19"}, // date is the whole stem
		{"refactor-dispatcher.md", ""},  // no date prefix
		{"pants-6d.md", ""},             // no date prefix
		{"2025-13-01-bad-month.md", ""}, // month out of range
		{"2025-10-32-bad-day.md", ""},   // day out of range
		{"2025-1-05-unpadded.md", ""},   // not zero-padded → not YYYY-MM-DD
		{"short.md", ""},                // shorter than 10 chars
		{"", ""},                        // empty
		{"2025-10-19", "2025-10-19"},    // no extension, exactly 10 chars
	}
	for _, c := range cases {
		if got := dateFromFilename(c.name); got != c.want {
			t.Errorf("dateFromFilename(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestMintUniqueID_RetriesPastCollisions(t *testing.T) {
	seen := map[string]bool{"AAA": true} // AAA already taken
	calls := 0
	gen := func(int64) string {
		calls++
		if calls < 3 {
			return "AAA" // collide twice, then yield a fresh one
		}
		return "BBB"
	}
	got, ok := mintUniqueID(0, seen, gen)
	if !ok || got != "BBB" {
		t.Fatalf("want BBB after retrying past collisions, got %q ok=%v", got, ok)
	}
	if !seen["BBB"] {
		t.Error("the assigned id must be recorded in seen so later files avoid it")
	}
	if calls != 3 {
		t.Errorf("want 3 generator calls (2 collisions + 1 hit), got %d", calls)
	}
}

func TestMintUniqueID_GivesUpOnPathologicalGenerator(t *testing.T) {
	got, ok := mintUniqueID(0, map[string]bool{"X": true}, func(int64) string { return "X" })
	if ok || got != "" {
		t.Errorf("a generator that always collides must give up, got %q ok=%v", got, ok)
	}
}

func TestFirstDateMillis_Preference(t *testing.T) {
	if _, ok := firstDateMillis("", "", ""); ok {
		t.Error("no dates → not ok")
	}
	if _, ok := firstDateMillis("not-a-date", ""); ok {
		t.Error("an unparseable date → not ok")
	}
	// The first PARSEABLE candidate wins (created preferred over updated_at).
	ms, ok := firstDateMillis("", "2026-01-02", "2026-03-04")
	if !ok {
		t.Fatal("a valid date should parse")
	}
	want, _ := time.Parse("2006-01-02", "2026-01-02")
	if ms != want.UnixMilli() {
		t.Errorf("got %d, want %d (first parseable candidate)", ms, want.UnixMilli())
	}
}
