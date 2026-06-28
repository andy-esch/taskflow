package render

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// decodeStrict pins each JSON envelope's shape: DisallowUnknownFields means a
// renamed/added key fails the test instead of silently passing a loose
// substring check — the schema_version contract made executable.
func decodeStrict(t *testing.T, data []byte, into any) {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(into); err != nil {
		t.Fatalf("envelope shape changed (strict decode failed): %v\n%s", err, data)
	}
}

var sampleTasks = []domain.Task{
	{Slug: "alpha", Status: domain.StatusInProgress, Declared: domain.StatusInProgress,
		Epic: "e1", Description: "first task", Tier: 2, Priority: "high",
		Created: "2026-06-01", Updated: "2026-06-10", Tags: []string{"go", "cli"}},
	{Slug: "beta", Status: domain.StatusReadyToStart, Declared: domain.StatusCompleted}, // misfiled
}

func TestTasksJSON_Envelope(t *testing.T) {
	var out bytes.Buffer
	problems := []domain.FileProblem{{Path: "bad.md", Message: "broken"}}
	if err := TasksJSON(&out, sampleTasks, problems); err != nil {
		t.Fatal(err)
	}
	var got struct {
		SchemaVersion string `json:"schema_version"`
		Tasks         []struct {
			Slug        string   `json:"slug"`
			Status      string   `json:"status"`
			Epic        string   `json:"epic,omitempty"`
			Description string   `json:"description,omitempty"`
			Effort      string   `json:"effort,omitempty"`
			Tier        int      `json:"tier,omitempty"`
			Priority    string   `json:"priority,omitempty"`
			Autonomy    int      `json:"autonomy_level,omitempty"`
			Created     string   `json:"created,omitempty"`
			Updated     string   `json:"updated_at,omitempty"`
			Tags        []string `json:"tags,omitempty"`
			Misfiled    bool     `json:"misfiled,omitempty"`
			Declared    string   `json:"declared_status,omitempty"`
		} `json:"tasks"`
		Unreadable []struct {
			Path    string `json:"path"`
			Message string `json:"message"`
		} `json:"unreadable,omitempty"`
	}
	decodeStrict(t, out.Bytes(), &got)
	if got.SchemaVersion != SchemaVersion {
		t.Errorf("schema_version = %q, want %q", got.SchemaVersion, SchemaVersion)
	}
	if len(got.Tasks) != 2 || got.Tasks[0].Slug != "alpha" || got.Tasks[0].Tier != 2 {
		t.Errorf("tasks payload wrong: %+v", got.Tasks)
	}
	// The misfiled signal (status ≠ folder) must be machine-readable — agents
	// are exactly the consumers who should detect drift (schema 1.1).
	if !got.Tasks[1].Misfiled || got.Tasks[1].Declared != "completed" {
		t.Errorf("misfiled task must carry misfiled+declared_status: %+v", got.Tasks[1])
	}
	if len(got.Unreadable) != 1 || got.Unreadable[0].Path != "bad.md" {
		t.Errorf("unreadable files must be included: %+v", got.Unreadable)
	}
}

func TestTasksHuman_TableAndMisfiledFlag(t *testing.T) {
	var out bytes.Buffer
	if err := TasksHuman(&out, NewStyle(false), sampleTasks); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, want := range []string{"alpha", "beta", "first task", "2 tasks", "⚠ 1 misfiled"} {
		if !strings.Contains(s, want) {
			t.Errorf("human output missing %q:\n%s", want, s)
		}
	}
	// Empty input writes nothing (no headers over zero rows).
	out.Reset()
	if err := TasksHuman(&out, NewStyle(false), nil); err != nil || out.Len() != 0 {
		t.Errorf("empty input should write nothing, got %q (%v)", out.String(), err)
	}
}

func TestLintJSON_Envelope(t *testing.T) {
	var out bytes.Buffer
	results := []core.LintResult{{Slug: "alpha", Issues: []domain.Issue{{Field: "tags", Message: "missing"}}}}
	problems := []domain.FileProblem{{Path: "bad.md", Message: "unterminated"}}
	if err := LintJSON(&out, results, problems); err != nil {
		t.Fatal(err)
	}
	var got struct {
		SchemaVersion string `json:"schema_version"`
		Unreadable    []struct {
			Path    string `json:"path"`
			Message string `json:"message"`
		} `json:"unreadable"`
		Issues []struct {
			Slug   string `json:"slug"`
			Issues []struct {
				Field   string `json:"field"`
				Message string `json:"message"`
			} `json:"issues"`
		} `json:"issues"`
	}
	decodeStrict(t, out.Bytes(), &got)
	if got.SchemaVersion != SchemaVersion || len(got.Issues) != 1 || got.Issues[0].Issues[0].Field != "tags" {
		t.Errorf("lint payload wrong:\n%s", out.String())
	}
	if len(got.Unreadable) != 1 {
		t.Errorf("unreadable files must be included:\n%s", out.String())
	}
}

func TestEpicsJSONAndHuman(t *testing.T) {
	epics := []core.EpicSummary{
		{Epic: domain.Epic{ID: "01-x", Status: "active", Description: "an epic"}, Total: 4, Done: 1},
	}
	var out bytes.Buffer
	if err := EpicsJSON(&out, epics, nil); err != nil {
		t.Fatal(err)
	}
	var got struct {
		SchemaVersion string `json:"schema_version"`
		Epics         []struct {
			ID          string   `json:"id"`
			Status      string   `json:"status,omitempty"`
			Description string   `json:"description,omitempty"`
			Priority    string   `json:"priority,omitempty"`
			Created     string   `json:"created,omitempty"`
			Tags        []string `json:"tags,omitempty"`
			Total       int      `json:"total"`
			Done        int      `json:"done"`
			Open        int      `json:"open"`
			Percent     int      `json:"percent"`
			Deprecated  int      `json:"deprecated"`
			Liveness    string   `json:"liveness"`
		} `json:"epics"`
		Unreadable []any `json:"unreadable,omitempty"`
	}
	decodeStrict(t, out.Bytes(), &got)
	if len(got.Epics) != 1 || got.Epics[0].Percent != 25 {
		t.Errorf("epics payload wrong:\n%s", out.String())
	}

	out.Reset()
	if err := EpicsHuman(&out, NewStyle(false), epics); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"01-x", "1/4", "25%"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("epics human output missing %q:\n%s", want, out.String())
		}
	}
}

// TestEpicShowHuman_Tree: epic show renders a status-grouped tree (lipgloss/v2),
// ANSI-free under no-color, with the deprecated footnote and the body intact.
func TestEpicShowHuman_Tree(t *testing.T) {
	var out bytes.Buffer
	tasks := []domain.Task{
		{Slug: "alpha", Status: domain.StatusCompleted},
		{Slug: "beta", Status: domain.StatusCompleted},
		{Slug: "gamma", Status: domain.StatusReadyToStart},
		{Slug: "delta", Status: domain.StatusDeprecated},
	}
	es := core.EpicSummary{Epic: domain.Epic{ID: "e1", Status: "active"}, Done: 2, Total: 3, Deprecated: 1}
	if err := EpicShowHuman(&out, NewStyle(false), es, tasks, "# body"); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	// progress line rolls up non-deprecated tasks (2 of 3 done = 66%); delta excluded.
	for _, want := range []string{"├──", "completed", "ready-to-start", "alpha", "gamma", "delta", "1 deprecated", "2/3", "66%", "# body"} {
		if !strings.Contains(s, want) {
			t.Errorf("epic show tree missing %q:\n%s", want, s)
		}
	}
	if strings.Contains(s, "\x1b[") {
		t.Errorf("no-color epic show must be ANSI-free:\n%q", s)
	}
}

func TestAuditsJSONAndHuman(t *testing.T) {
	audits := []domain.Audit{{Slug: "2026-06-01-x", Bucket: domain.AuditOpen, Area: "store", Date: "2026-06-01", Findings: 5, OpenFindings: 2}}
	var out bytes.Buffer
	if err := AuditsJSON(&out, audits, nil); err != nil {
		t.Fatal(err)
	}
	var got struct {
		SchemaVersion string `json:"schema_version"`
		Audits        []struct {
			Slug               string `json:"slug"`
			Bucket             string `json:"bucket"`
			Area               string `json:"area,omitempty"`
			Date               string `json:"date,omitempty"`
			Findings           int    `json:"findings"`
			OpenFindings       int    `json:"open_findings"`
			InProgressFindings int    `json:"in_progress_findings"`
			DoneFindings       int    `json:"done_findings"`
			DroppedFindings    int    `json:"dropped_findings"`
		} `json:"audits"`
		Unreadable []any `json:"unreadable,omitempty"`
	}
	decodeStrict(t, out.Bytes(), &got)
	if len(got.Audits) != 1 || got.Audits[0].OpenFindings != 2 {
		t.Errorf("audits payload wrong:\n%s", out.String())
	}

	out.Reset()
	if err := AuditsHuman(&out, NewStyle(false), audits); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "2026-06-01-x") || !strings.Contains(out.String(), "store") {
		t.Errorf("audits human output wrong:\n%s", out.String())
	}
}

// TestAuditShowHuman_FindingTree: audit show renders meta + a status-grouped
// finding tree (lifecycle order, glyph-coded) + the body, mirroring epic show.
// A finding with no status must land in a trailing group, not vanish.
func TestAuditShowHuman_FindingTree(t *testing.T) {
	a := domain.Audit{Slug: "2026-06-01-x", Bucket: domain.AuditOpen, Area: "store", Date: "2026-06-01", Findings: 3, OpenFindings: 1, DoneFindings: 2}
	findings := []domain.Finding{
		{Code: "M1", Title: "done deal", Status: "fixed"},
		{Code: "H1", Title: "still open", Status: "open"},
		{Code: "L9", Title: "mystery"}, // missing status → grouped under (no status), not dropped
	}
	var out bytes.Buffer
	if err := AuditShowHuman(&out, NewStyle(false), a, findings, "# body"); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, want := range []string{"2026-06-01-x", "66%", "2/3", "├──", "open", "H1", "still open", "fixed", "M1", "(no status)", "L9", "mystery", "# body"} {
		if !strings.Contains(s, want) {
			t.Errorf("audit show tree missing %q:\n%s", want, s)
		}
	}
	// lifecycle order: the open group precedes the fixed group (despite input order).
	if strings.Index(s, "still open") > strings.Index(s, "done deal") {
		t.Errorf("open group should render before fixed:\n%s", s)
	}
	if strings.Contains(s, "\x1b[") {
		t.Errorf("no-color audit show must be ANSI-free:\n%q", s)
	}
}

// TestEpicShowHuman_FitsWidth pins the width-fit: at a narrow Style width, the
// metadata values AND the task tree are truncated so no line overflows the
// terminal. Piped output (width 0) stays full — covered by the other show tests.
func TestEpicShowHuman_FitsWidth(t *testing.T) {
	st := NewStyle(false).WithWidth(40)
	tasks := []domain.Task{{Slug: "a-very-long-task-slug-that-would-overflow-a-narrow-terminal", Status: domain.StatusReadyToStart}}
	epic := domain.Epic{ID: "01-x", Status: "active", Description: "A deliberately long epic description that must be truncated to the terminal width"}
	var out bytes.Buffer
	if err := EpicShowHuman(&out, st, core.EpicSummary{Epic: epic, Total: 1}, tasks, "# body"); err != nil {
		t.Fatal(err)
	}
	for _, ln := range strings.Split(strings.TrimRight(out.String(), "\n"), "\n") {
		if visibleWidth(ln) > 40 {
			t.Errorf("line exceeds width 40 (%d): %q", visibleWidth(ln), ln)
		}
	}
	if !strings.Contains(out.String(), "…") {
		t.Errorf("expected a truncation ellipsis at width 40:\n%s", out.String())
	}
}

func TestSummaryOutputs(t *testing.T) {
	s := core.Summary{
		Counts: []core.StatusCount{
			{Status: domain.StatusInProgress, Count: 2},
			{Status: domain.StatusCompleted, Count: 5},
		},
		InProgress: []domain.Task{{Slug: "alpha", Status: domain.StatusInProgress, Declared: domain.StatusInProgress}},
		Epics:      []core.EpicSummary{{Epic: domain.Epic{ID: "01-x"}, Total: 2, Done: 1}},
		OpenAudits: []domain.Audit{{Slug: "2026-06-01-audit-x", Bucket: domain.AuditOpen, Area: "store", Findings: 4, OpenFindings: 1, DoneFindings: 3}},
		Misfiled:   1,
	}
	var out bytes.Buffer
	if err := SummaryJSON(&out, s); err != nil {
		t.Fatal(err)
	}
	var got struct {
		SchemaVersion string `json:"schema_version"`
		OpenAudits    []struct {
			Slug         string `json:"slug"`
			OpenFindings int    `json:"open_findings"`
		} `json:"open_audits"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("summary json invalid: %v", err)
	}
	if got.SchemaVersion != SchemaVersion {
		t.Errorf("summary missing schema_version:\n%s", out.String())
	}
	if len(got.OpenAudits) != 1 || got.OpenAudits[0].Slug != "2026-06-01-audit-x" || got.OpenAudits[0].OpenFindings != 1 {
		t.Errorf("summary open_audits wrong:\n%s", out.String())
	}

	out.Reset()
	if err := SummaryHuman(&out, NewStyle(false), s); err != nil {
		t.Fatal(err)
	}
	// open audits surface in their own dashboard section with the rollup (3/4 resolved).
	for _, want := range []string{"in-progress", "alpha", "01-x", "Open audits", "2026-06-01-audit-x", "3/4"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("summary human output missing %q:\n%s", want, out.String())
		}
	}
}

// TestSummaryHuman_NoAudits pins the self-hiding contract: with no open audits the
// dashboard renders no Audits section and the JSON omits open_audits entirely.
func TestSummaryHuman_NoAudits(t *testing.T) {
	s := core.Summary{Epics: []core.EpicSummary{{Epic: domain.Epic{ID: "01-x"}, Total: 2, Done: 1}}}
	var out bytes.Buffer
	if err := SummaryJSON(&out, s); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "open_audits") {
		t.Errorf("open_audits must be omitted when there are none:\n%s", out.String())
	}
	out.Reset()
	if err := SummaryHuman(&out, NewStyle(false), s); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "Open audits") {
		t.Errorf("no Audits section expected when there are none:\n%s", out.String())
	}
}

// TestSummary_RevisitDueNudge pins the snooze surface: a non-zero RevisitDue
// renders the ↻ nudge (no emoji) in the human dashboard and carries revisit_due in
// the JSON envelope; zero renders no nudge.
func TestSummary_RevisitDueNudge(t *testing.T) {
	s := core.Summary{RevisitDue: 2}
	var out bytes.Buffer
	if err := SummaryHuman(&out, NewStyle(false), s); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "↻ 2 deferred due to revisit") {
		t.Errorf("expected revisit nudge in dashboard:\n%s", out.String())
	}
	out.Reset()
	if err := SummaryJSON(&out, s); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"revisit_due":2`) {
		t.Errorf("expected revisit_due:2 in summary json:\n%s", out.String())
	}

	// Zero: no nudge in the dashboard (the field is still always present in JSON).
	out.Reset()
	if err := SummaryHuman(&out, NewStyle(false), core.Summary{}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "deferred due to revisit") {
		t.Errorf("no revisit nudge expected when RevisitDue is 0:\n%s", out.String())
	}
}

// TestMoves_RevisitAtReported pins that a per-item revisit_at (set by the defer
// decorator on a real run AND on a --dry-run preview) shows in the human line and
// rides the JSON move report — so a snooze is confirmed, not just the move.
func TestMoves_RevisitAtReported(t *testing.T) {
	results := []MoveResult{{Slug: "alpha", To: "deferred", RevisitAt: "2026-09-01"}}

	var human bytes.Buffer
	MovesHuman(&human, &human, NewStyle(false), results, true)
	if got := human.String(); !strings.Contains(got, "would move alpha -> deferred") || !strings.Contains(got, "revisit 2026-09-01") {
		t.Errorf("dry-run human move line should confirm the revisit date:\n%s", got)
	}

	var j bytes.Buffer
	if err := MovesJSON(&j, results, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(j.String(), `"revisit_at":"2026-09-01"`) {
		t.Errorf("move JSON should carry revisit_at:\n%s", j.String())
	}

	// A move with no revisit date (any other transition) omits it entirely.
	var plain bytes.Buffer
	MovesHuman(&plain, &plain, NewStyle(false), []MoveResult{{Slug: "beta", To: "next-up"}}, false)
	if strings.Contains(plain.String(), "revisit") {
		t.Errorf("a dateless move must not mention a revisit:\n%s", plain.String())
	}
}

func TestFixOutputs(t *testing.T) {
	results := []domain.FixResult{{Path: "tasks/ready-to-start/a.md", Changes: []string{"tags: normalized to a YAML list"}}}
	var out bytes.Buffer
	if err := FixJSON(&out, results, nil, nil, true); err != nil {
		t.Fatal(err)
	}
	var got struct {
		SchemaVersion string `json:"schema_version"`
		DryRun        bool   `json:"dry_run"`
		Fixed         []struct {
			Path    string   `json:"path"`
			Changes []string `json:"changes"`
		} `json:"fixed"`
		Unreadable []struct {
			Path    string `json:"path"`
			Message string `json:"message"`
		} `json:"unreadable"`
		Remaining []struct {
			Slug   string         `json:"slug"`
			Issues []domain.Issue `json:"issues"`
		} `json:"remaining"`
	}
	decodeStrict(t, out.Bytes(), &got)
	if !got.DryRun || len(got.Fixed) != 1 {
		t.Errorf("fix payload wrong:\n%s", out.String())
	}
	if got.Unreadable == nil {
		t.Errorf("unreadable should be present (empty, not null) on a dry-run:\n%s", out.String())
	}
	if got.Remaining == nil {
		t.Errorf("remaining should be present (empty, not null) on a dry-run:\n%s", out.String())
	}

	out.Reset()
	FixHuman(&out, NewStyle(false), results, nil, false)
	if !strings.Contains(out.String(), "a.md") {
		t.Errorf("fix human output missing the path:\n%s", out.String())
	}
	out.Reset()
	FixHuman(&out, NewStyle(false), nil, nil, false)
	if out.Len() == 0 {
		t.Error("zero fixes should still print a confirmation")
	}

	// Leftover lint findings the pass couldn't repair surface after the fixed list.
	out.Reset()
	FixHuman(&out, NewStyle(false), results, []core.LintResult{{Slug: "01-e", Issues: []domain.Issue{{Field: "priority", Message: "missing"}}}}, false)
	if !strings.Contains(out.String(), "could not auto-repair") || !strings.Contains(out.String(), "01-e") {
		t.Errorf("fix human output should surface leftover lint findings:\n%s", out.String())
	}
}

func TestProblemsHuman(t *testing.T) {
	var out bytes.Buffer
	ProblemsHuman(&out, NewStyle(false), []domain.FileProblem{{Path: "x.md", Message: "unterminated frontmatter"}})
	if !strings.Contains(out.String(), "x.md") || !strings.Contains(out.String(), "unterminated") {
		t.Errorf("problems output wrong:\n%s", out.String())
	}
}

// TestCreatedSlugNote pins the surfaced-slug UX: a title whose slug diverges
// beyond the obvious (filename-hostile chars dropped) gets a "→ slug: <slug>"
// line so the derivation isn't silent, while an everyday title (only lowercased +
// space→hyphen) prints nothing.
func TestCreatedSlugNote(t *testing.T) {
	var out bytes.Buffer
	CreatedSlugNote(&out, NewStyle(false), "Wire OAuth: PKCE + refresh", "wire-oauth-pkce-refresh")
	if got := out.String(); got != "→ slug: wire-oauth-pkce-refresh\n" {
		t.Errorf("diverging title should surface the slug, got %q", got)
	}
	// An everyday title — just case + spaces — is no surprise, so it's silent. The
	// apostrophe and trailing-dot cases are silent too: Slugify drops apostrophes and
	// trims trailing '.'/'-', so the note (whose naiveSlug mirrors those) must NOT
	// over-fire on "don't" or "… backoff.".
	for _, clean := range []string{
		"Add retry backoff", "add-retry-backoff", "Multi-Entity Navigation",
		"Don't break the build", "Fix the parser's edge case",
		"Add retry backoff.", "Tidy up the config-",
	} {
		out.Reset()
		CreatedSlugNote(&out, NewStyle(false), clean, domain.Slugify(clean))
		if out.Len() != 0 {
			t.Errorf("a no-surprise title (%q) should print nothing, got %q", clean, out.String())
		}
	}
	// A genuinely-diverging title (a character turned into a word-break) still fires.
	out.Reset()
	CreatedSlugNote(&out, NewStyle(false), "Refactor: split the dispatcher", domain.Slugify("Refactor: split the dispatcher"))
	if out.Len() == 0 {
		t.Error("a title with a dropped colon should surface the slug")
	}
}
