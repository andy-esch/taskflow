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
		{Epic: domain.Epic{ID: "01-x", Status: "in-progress", Description: "an epic"}, Total: 4, Done: 1},
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
			Percent     int      `json:"percent"`
			Deprecated  int      `json:"deprecated"`
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
	if err := EpicShowHuman(&out, NewStyle(false), domain.Epic{ID: "e1", Status: "planning"}, tasks, "# body"); err != nil {
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
			Slug         string `json:"slug"`
			Bucket       string `json:"bucket"`
			Area         string `json:"area,omitempty"`
			Date         string `json:"date,omitempty"`
			Findings     int    `json:"findings"`
			OpenFindings int    `json:"open_findings"`
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

func TestSummaryOutputs(t *testing.T) {
	s := core.Summary{
		Counts: []core.StatusCount{
			{Status: domain.StatusInProgress, Count: 2},
			{Status: domain.StatusCompleted, Count: 5},
		},
		InProgress: []domain.Task{{Slug: "alpha", Status: domain.StatusInProgress, Declared: domain.StatusInProgress}},
		Epics:      []core.EpicSummary{{Epic: domain.Epic{ID: "01-x"}, Total: 2, Done: 1}},
		OpenAudits: []domain.Audit{{Slug: "2026-06-01-audit-x", Bucket: domain.AuditOpen, Area: "store", Findings: 4, OpenFindings: 1}},
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

func TestFixOutputs(t *testing.T) {
	results := []domain.FixResult{{Path: "tasks/ready-to-start/a.md", Changes: []string{"tags: normalized to a YAML list"}}}
	var out bytes.Buffer
	if err := FixJSON(&out, results, nil, true); err != nil {
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
	}
	decodeStrict(t, out.Bytes(), &got)
	if !got.DryRun || len(got.Fixed) != 1 {
		t.Errorf("fix payload wrong:\n%s", out.String())
	}
	if got.Unreadable == nil {
		t.Errorf("unreadable should be present (empty, not null) on a dry-run:\n%s", out.String())
	}

	out.Reset()
	FixHuman(&out, NewStyle(false), results, false)
	if !strings.Contains(out.String(), "a.md") {
		t.Errorf("fix human output missing the path:\n%s", out.String())
	}
	out.Reset()
	FixHuman(&out, NewStyle(false), nil, false)
	if out.Len() == 0 {
		t.Error("zero fixes should still print a confirmation")
	}
}

func TestProblemsHuman(t *testing.T) {
	var out bytes.Buffer
	ProblemsHuman(&out, NewStyle(false), []domain.FileProblem{{Path: "x.md", Message: "unterminated frontmatter"}})
	if !strings.Contains(out.String(), "x.md") || !strings.Contains(out.String(), "unterminated") {
		t.Errorf("problems output wrong:\n%s", out.String())
	}
}
