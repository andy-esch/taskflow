package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestTaskList_Quiet(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "list", "-q")
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Errorf("-q should list both slugs:\n%q", out)
	}
	// Every line is a bare id — no header, no whitespace/decoration (xargs-safe).
	for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
		if l == "" || strings.ContainsAny(l, " \t") {
			t.Errorf("-q line should be a bare id, got %q", l)
		}
	}
}

// TestOutput_NameEqualsQuiet pins -q as a pure alias for -o name.
func TestOutput_NameEqualsQuiet(t *testing.T) {
	root := setupRepo(t)
	q := runRoot(t, "-C", root, "task", "list", "-q")
	name := runRoot(t, "-C", root, "task", "list", "-o", "name")
	if q != name {
		t.Errorf("-q and -o name must be identical:\n -q:      %q\n -o name: %q", q, name)
	}
}

func TestTaskList_Table(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "list", "-o", "table")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatalf("-o table needs a header + ≥1 row:\n%q", out)
	}
	if lines[0] != "slug\tstatus\ttier\tpriority\tepic\tupdated\tdescription\trevisit_at\tid" {
		t.Errorf("-o table header wrong: %q", lines[0])
	}
	if cols := strings.Split(lines[1], "\t"); len(cols) != 9 {
		t.Errorf("-o table row should have 9 tab-separated columns, got %d: %q", len(cols), lines[1])
	}
}

// TestColumns_Projection: -c selects (and orders) the columns, and implies table.
func TestColumns_Projection(t *testing.T) {
	root := setupRepo(t)
	// Explicit -o table -c, and the bare -c (which implies -o table), must agree.
	for _, args := range [][]string{
		{"task", "list", "-o", "table", "-c", "status,slug"},
		{"task", "list", "-c", "status,slug"},
	} {
		out := runRoot(t, append([]string{"-C", root}, args...)...)
		lines := strings.Split(strings.TrimSpace(out), "\n")
		if lines[0] != "status\tslug" {
			t.Errorf("%v: header should be the projected columns in order, got %q", args, lines[0])
		}
		if cols := strings.Split(lines[1], "\t"); len(cols) != 2 {
			t.Errorf("%v: row should have 2 columns, got %d: %q", args, len(cols), lines[1])
		}
	}
}

// TestWriteError_JSONIsCompact pins that the --json error envelope is compact
// (single line) like every other --json envelope, and carries the sentinel-
// mapped code. (WriteError is what main calls on a fatal error under --json.)
func TestWriteError_JSONIsCompact(t *testing.T) {
	var buf bytes.Buffer
	WriteError(&buf, fmt.Errorf("%w: bad thing", domain.ErrValidation), true)
	out := buf.String()
	if strings.Count(out, "\n") != 1 {
		t.Errorf("JSON error envelope must be compact (one trailing newline): %q", out)
	}
	if !strings.HasPrefix(out, `{"schema_version"`) || !strings.Contains(out, `"code":"validation"`) {
		t.Errorf("compact error envelope shape wrong: %q", out)
	}
}

// TestColumns_JSONProjection: `--json -c` narrows each row to the selected
// columns (in order), keeping the schema_version envelope. The single biggest
// agent token win — field projection on the format agents actually parse.
func TestColumns_JSONProjection(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "list", "--json", "-c", "slug,status")
	var got struct {
		SchemaVersion string           `json:"schema_version"`
		Tasks         []map[string]any `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if got.SchemaVersion == "" {
		t.Errorf("projected envelope must carry schema_version:\n%s", out)
	}
	if len(got.Tasks) == 0 {
		t.Fatalf("expected projected rows:\n%s", out)
	}
	for _, row := range got.Tasks {
		if len(row) != 2 {
			t.Errorf("row should carry exactly the 2 selected fields, got %d: %v", len(row), row)
		}
		if _, ok := row["slug"]; !ok {
			t.Errorf("row missing selected field 'slug': %v", row)
		}
		if _, ok := row["status"]; !ok {
			t.Errorf("row missing selected field 'status': %v", row)
		}
		// A non-selected, normally-present field must be absent.
		if _, ok := row["description"]; ok {
			t.Errorf("row should NOT carry the unselected 'description' field: %v", row)
		}
	}
	// Key order follows -c (slug before status), which a plain map would lose.
	if i, j := strings.Index(out, "\"slug\""), strings.Index(out, "\"status\""); i < 0 || j < 0 || i > j {
		t.Errorf("projected keys should appear in -c order (slug before status):\n%s", out)
	}
}

func TestStableIDColumns_ProjectTaskAndAuditHandles(t *testing.T) {
	for _, tc := range []struct {
		name, noun, listKey string
	}{
		{name: "task", noun: "task", listKey: "tasks"},
		{name: "audit", noun: "audit", listKey: "audits"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := runRoot(t, "-C", fixtureRepo, tc.noun, "list", "--all", "--json", "-c", "id,slug")
			var envelope map[string]json.RawMessage
			if err := json.Unmarshal([]byte(out), &envelope); err != nil {
				t.Fatalf("invalid projected JSON: %v\n%s", err, out)
			}
			var rows []map[string]string
			if err := json.Unmarshal(envelope[tc.listKey], &rows); err != nil {
				t.Fatalf("invalid %s rows: %v\n%s", tc.listKey, err, out)
			}
			if len(rows) == 0 {
				t.Fatalf("expected at least one projected %s row", tc.name)
			}
			for _, row := range rows {
				if len(row) != 2 || row["id"] == "" || row["slug"] == "" || row["id"] == row["slug"] {
					t.Errorf("projected row must distinguish durable id from slug: %#v", row)
				}
			}
			if id, slug := strings.Index(out, `"id"`), strings.Index(out, `"slug"`); id < 0 || slug < 0 || id > slug {
				t.Errorf("projected keys should retain requested id,slug order:\n%s", out)
			}

			table := runRoot(t, "-C", fixtureRepo, tc.noun, "list", "--all", "-o", "table", "-c", "id,slug")
			if header := strings.SplitN(table, "\n", 2)[0]; header != "id\tslug" {
				t.Errorf("explicit table projection header = %q, want id\\tslug", header)
			}
		})
	}
}

func TestColumns_JSONProjectionUsesCanonicalWireSelectorAndKey(t *testing.T) {
	root := setupRepo(t) // neither fixture carries updated_at
	canonical := runRoot(t, "-C", root, "task", "list", "--json", "-c", "slug,updated_at")
	legacy := runRoot(t, "-C", root, "task", "list", "--json", "-c", "slug,updated")
	var got struct {
		Tasks []map[string]string `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(canonical), &got); err != nil {
		t.Fatal(err)
	}
	for _, row := range got.Tasks {
		if _, ok := row["updated_at"]; !ok {
			t.Errorf("projected row missing canonical updated_at key: %#v", row)
		}
		if row["updated_at"] != "" {
			t.Errorf("never-edited fixture invented updated_at: %#v", row)
		}
		if _, leaked := row["updated"]; leaked {
			t.Errorf("projected row leaked legacy updated key: %#v", row)
		}
	}
	got.Tasks = nil
	if err := json.Unmarshal([]byte(legacy), &got); err != nil {
		t.Fatal(err)
	}
	for _, row := range got.Tasks {
		if _, ok := row["updated"]; !ok {
			t.Errorf("legacy projection missing compatibility key: %#v", row)
		}
		if row["updated"] != "" {
			t.Errorf("legacy projection should still use raw values: %#v", row)
		}
		if _, leaked := row["updated_at"]; leaked {
			t.Errorf("legacy projection unexpectedly renamed its output key: %#v", row)
		}
	}

	canonicalTable := runRoot(t, "-C", root, "task", "list", "-c", "updated_at")
	legacyTable := runRoot(t, "-C", root, "task", "list", "-c", "updated")
	canonicalLines := strings.Split(strings.TrimSuffix(canonicalTable, "\n"), "\n")
	legacyLines := strings.Split(strings.TrimSuffix(legacyTable, "\n"), "\n")
	if header := canonicalLines[0]; header != "updated_at" {
		t.Errorf("explicit canonical table selection should echo its name, got %q", header)
	}
	if header := legacyLines[0]; header != "updated" {
		t.Errorf("legacy table selection should preserve its header, got %q", header)
	}
	for _, line := range canonicalLines[1:] {
		if line != "" {
			t.Errorf("canonical updated_at table cells should be empty for never-edited fixtures, got %q", line)
		}
	}
}

// TestColumns_CanonicalUpdatedAtUsesRawValueEndToEnd gives the command layer a
// real created-but-never-edited record. A fixture with neither date cannot tell
// the raw extractor from the legacy created-date fallback and would let the
// original projection defect survive outside render's unit tests.
func TestColumns_CanonicalUpdatedAtUsesRawValueEndToEnd(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "epic", "new", "Projection fixture", "--description", "projection fixture")
	runRoot(t, "-C", root, "task", "new", "Never edited task",
		"--epic", "01-projection-fixture", "--tags", "contract", "--description", "never edited")

	full := runRoot(t, "-C", root, "task", "list", "--json")
	var fullTasks struct {
		Tasks []struct {
			Created string `json:"created"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(full), &fullTasks); err != nil || len(fullTasks.Tasks) != 1 {
		t.Fatalf("full task list: err=%v output=%s", err, full)
	}
	created := fullTasks.Tasks[0].Created
	if created == "" {
		t.Fatal("task new should set created")
	}

	canonicalJSON := runRoot(t, "-C", root, "task", "list", "--json", "-c", "updated_at")
	legacyJSON := runRoot(t, "-C", root, "task", "list", "--json", "-c", "updated")
	for name, output := range map[string]string{"updated_at": canonicalJSON, "updated": legacyJSON} {
		var projected struct {
			Tasks []map[string]string `json:"tasks"`
		}
		if err := json.Unmarshal([]byte(output), &projected); err != nil || len(projected.Tasks) != 1 {
			t.Fatalf("%s task projection: err=%v output=%s", name, err, output)
		}
		if got := projected.Tasks[0][name]; got != "" {
			t.Fatalf("%s projection invented updated_at from created: %q", name, got)
		}
	}
	if got := runRoot(t, "-C", root, "task", "list", "-o", "table", "-c", "updated_at"); got != "updated_at\n\n" {
		t.Fatalf("canonical task table should carry the raw empty value, got %q", got)
	}
	if got := runRoot(t, "-C", root, "task", "list", "-o", "table", "-c", "updated"); got != "updated\n"+created+"\n" {
		t.Fatalf("legacy task table should retain the created fallback, got %q", got)
	}

	runRoot(t, "-C", root, "research", "new", "Never edited research",
		"--created", "2026-01-06", "--tags", "contract", "--description", "never edited")
	researchJSON := runRoot(t, "-C", root, "research", "list", "--json", "-c", "updated_at")
	var projectedResearch struct {
		Research []map[string]string `json:"research"`
	}
	if err := json.Unmarshal([]byte(researchJSON), &projectedResearch); err != nil || len(projectedResearch.Research) != 1 {
		t.Fatalf("research projection: err=%v output=%s", err, researchJSON)
	}
	if got := projectedResearch.Research[0]["updated_at"]; got != "" {
		t.Fatalf("canonical research projection invented updated_at from created: %q", got)
	}
	if got := runRoot(t, "-C", root, "research", "list", "-o", "csv", "-c", "updated_at"); got != "updated_at\n\n" {
		t.Fatalf("canonical research CSV should carry the raw empty value, got %q", got)
	}
	if got := runRoot(t, "-C", root, "research", "list", "-o", "table", "-c", "updated"); got != "updated\n2026-01-06\n" {
		t.Fatalf("legacy research table should retain the created fallback, got %q", got)
	}
}

// TestTable_EmptyIsHeaderOnly pins the porcelain contract end-to-end: an empty
// result still emits the header row (stable schema), where -q/human emit nothing.
func TestTable_EmptyIsHeaderOnly(t *testing.T) {
	root := setupRepo(t) // alpha/beta have no tags, so --tag filters to empty
	out := strings.TrimSpace(runRoot(t, "-C", root, "task", "list", "--tag", "zzz-none", "-o", "table"))
	if out != "slug\tstatus\ttier\tpriority\tepic\tupdated\tdescription\trevisit_at\tid" {
		t.Errorf("empty -o table should be header-only, got %q", out)
	}
	if q := runRoot(t, "-C", root, "task", "list", "--tag", "zzz-none", "-q"); q != "" {
		t.Errorf("empty -q should emit nothing, got %q", q)
	}
}

// TestOutput_JSONAlias pins --json (universal) as identical to -o json on list.
func TestOutput_JSONAlias(t *testing.T) {
	root := setupRepo(t)
	a := runRoot(t, "-C", root, "task", "list", "--json")
	b := runRoot(t, "-C", root, "task", "list", "-o", "json")
	if a != b {
		t.Errorf("--json and -o json must be identical:\n --json:  %q\n -o json: %q", a, b)
	}
}

func TestColumns_UnknownColumn(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "list", "-c", "slug,bogus"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("unknown column should error; output:\n%s", out.String())
	} else if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error should name the bad column: %v", err)
	}
}

// TestTable_ByteStableUnderColor pins the -o table contract: it ignores styling
// entirely, so it's byte-identical whether color is forced on or off (a script
// can rely on it). The default human table, by contrast, would emit ANSI here.
func TestTable_ByteStableUnderColor(t *testing.T) {
	root := setupRepo(t)
	on := runRoot(t, "-C", root, "task", "list", "-o", "table", "--color=always")
	off := runRoot(t, "-C", root, "task", "list", "-o", "table", "--color=never")
	if strings.Contains(on, "\x1b[") {
		t.Errorf("-o table must carry no ANSI even with --color=always:\n%q", on)
	}
	if on != off {
		t.Errorf("-o table must be byte-stable across color settings:\n on:  %q\n off: %q", on, off)
	}
}

func TestTaskList_CSV(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "list", "-o", "csv")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if lines[0] != "slug,status,tier,priority,epic,updated,description,revisit_at,id" {
		t.Errorf("-o csv header wrong: %q", lines[0])
	}
	if cols := strings.Split(lines[1], ","); len(cols) != 9 {
		t.Errorf("-o csv row should have 9 comma-separated columns, got %d: %q", len(cols), lines[1])
	}
	// -c projects csv too (csv is columnar, like table).
	proj := runRoot(t, "-C", root, "task", "list", "-o", "csv", "-c", "slug,status")
	if h := strings.SplitN(proj, "\n", 2)[0]; h != "slug,status" {
		t.Errorf("-o csv -c projection header wrong: %q", h)
	}
}

func TestAuditList_Table(t *testing.T) {
	root := setupAuditRepo(t)
	out := runRoot(t, "-C", root, "audit", "list", "-o", "table")
	if h := strings.SplitN(out, "\n", 2)[0]; h != "slug\tbucket\tarea\tdate\tfindings\topen\tid" {
		t.Errorf("audit -o table header wrong: %q", h)
	}
}

func TestTransition_FailureToStderr(t *testing.T) {
	root := setupRepo(t)
	var out, errOut bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &errOut)
	// alpha (ready-to-start) starts ok; ghost is not found → a partial failure.
	cmd.SetArgs([]string{"-C", root, "task", "start", "alpha", "ghost"})
	_ = cmd.Execute()
	if !strings.Contains(out.String(), "alpha -> in-progress") {
		t.Errorf("success confirmation should be on stdout:\n%q", out.String())
	}
	if strings.Contains(out.String(), "ghost") {
		t.Errorf("failure must not be on stdout (the data stream):\n%q", out.String())
	}
	if !strings.Contains(errOut.String(), "ghost") {
		t.Errorf("failure should be on stderr:\n%q", errOut.String())
	}
}

func TestList_ModeConflicts(t *testing.T) {
	root := setupRepo(t)
	// Every conflicting combination errors (validation, exit 11) and produces no
	// data — the format axis admits at most one selection, and -c needs a
	// projectable format (table/csv/json, not name/human). NB: `-c` WITH `--json`
	// is now a valid projection (see TestColumns_JSONProjection), not a conflict.
	for _, args := range [][]string{
		{"task", "list", "--json", "-o", "table"}, // --json vs explicit -o
		{"task", "list", "--json", "-q"},          // json alias vs name alias
		{"task", "list", "-q", "-o", "table"},     // name alias vs explicit -o
		{"task", "list", "-c", "slug", "-q"},      // -c needs columns, not name
		{"task", "list", "-o", "bogus"},           // unknown format
	} {
		var out bytes.Buffer
		cmd := NewRootCmd(strings.NewReader(""), &out, &out)
		cmd.SetArgs(append([]string{"-C", root}, args...))
		if err := cmd.Execute(); err == nil {
			t.Errorf("expected an error for %v", args)
		}
	}
}

// TestColumns_ConflictNamesEveryOffender guards determinism: when -c collides
// with more than one format, the error names them all in a stable order (the
// loop over the requested-format map would otherwise pick one at random).
func TestColumns_ConflictNamesEveryOffender(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "list", "-c", "slug", "--json", "-q"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected a conflict error")
	}
	for _, want := range []string{"--json", "-q/--quiet"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("conflict should name %q: %v", want, err)
		}
	}
}

// TestComplete_OutputFormats: `-o <TAB>` offers exactly the four formats.
func TestComplete_OutputFormats(t *testing.T) {
	root := setupRepo(t)
	got := complete(t, "-C", root, "task", "list", "-o", "")
	for _, want := range []string{"human", "json", "name", "table", "csv"} {
		if !has(got, want) {
			t.Errorf("output completion missing %q: %v", want, got)
		}
	}
	if has(got, "alpha") {
		t.Errorf("output completion must not leak slugs: %v", got)
	}
}

// TestComplete_Columns: `-c` completes column names, prefixes the chosen ones,
// and drops a column already in the list (the dedup nicety the known-set buys).
func TestComplete_Columns(t *testing.T) {
	root := setupRepo(t)
	if got := complete(t, "-C", root, "task", "list", "-c", "sl"); !has(got, "slug") {
		t.Errorf("`-c sl` should complete to slug: %v", got)
	}
	got := complete(t, "-C", root, "task", "list", "-c", "slug,sta")
	if !has(got, "slug,status") {
		t.Errorf("`-c slug,sta` should complete to slug,status: %v", got)
	}
	if has(got, "slug") || has(got, "slug,slug") {
		t.Errorf("an already-chosen column must not be re-offered: %v", got)
	}
	for _, noun := range []string{"task", "audit", "thread"} {
		if got := complete(t, "-C", root, noun, "list", "-c", "i"); !has(got, "id") {
			t.Errorf("%s stable id column should be discoverable through completion: %v", noun, got)
		}
	}
	for _, tc := range []struct {
		args []string
		bad  string
	}{
		{[]string{"task", "list", "-c", "updated,"}, "updated,updated_at"},
		{[]string{"research", "list", "-c", "updated,"}, "updated,updated_at"},
		{[]string{"audit", "list", "-c", "open,"}, "open,open_findings"},
	} {
		if got := complete(t, append([]string{"-C", root}, tc.args...)...); has(got, tc.bad) {
			t.Errorf("legacy alias must suppress its canonical duplicate %q: %v", tc.bad, got)
		}
	}
}
