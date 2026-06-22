package cli

import (
	"bytes"
	"strings"
	"testing"
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
	if lines[0] != "slug\tstatus\ttier\tpriority\tepic\tupdated\tdescription" {
		t.Errorf("-o table header wrong: %q", lines[0])
	}
	if cols := strings.Split(lines[1], "\t"); len(cols) != 7 {
		t.Errorf("-o table row should have 7 tab-separated columns, got %d: %q", len(cols), lines[1])
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

// TestTable_EmptyIsHeaderOnly pins the porcelain contract end-to-end: an empty
// result still emits the header row (stable schema), where -q/human emit nothing.
func TestTable_EmptyIsHeaderOnly(t *testing.T) {
	root := setupRepo(t) // alpha/beta have no tags, so --tag filters to empty
	out := strings.TrimSpace(runRoot(t, "-C", root, "task", "list", "--tag", "zzz-none", "-o", "table"))
	if out != "slug\tstatus\ttier\tpriority\tepic\tupdated\tdescription" {
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
	if lines[0] != "slug,status,tier,priority,epic,updated,description" {
		t.Errorf("-o csv header wrong: %q", lines[0])
	}
	if cols := strings.Split(lines[1], ","); len(cols) != 7 {
		t.Errorf("-o csv row should have 7 comma-separated columns, got %d: %q", len(cols), lines[1])
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
	if h := strings.SplitN(out, "\n", 2)[0]; h != "slug\tbucket\tarea\tdate\tfindings\topen" {
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
	// data — the format axis admits at most one selection, and -c needs table.
	for _, args := range [][]string{
		{"task", "list", "--json", "-o", "table"},    // --json vs explicit -o
		{"task", "list", "--json", "-q"},             // json alias vs name alias
		{"task", "list", "-q", "-o", "table"},        // name alias vs explicit -o
		{"task", "list", "-c", "slug", "-o", "json"}, // -c needs table, not json
		{"task", "list", "-c", "slug", "-q"},         // -c needs table, not name
		{"task", "list", "-o", "bogus"},              // unknown format
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
}
