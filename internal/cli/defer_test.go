package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/cli/prompt"
	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/store"
)

// readTaskFile returns the raw contents of a task file at tasks/<status>/<name>.
func readTaskFile(t *testing.T, root, status, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "tasks", status, name))
	if err != nil {
		t.Fatalf("read %s/%s: %v", status, name, err)
	}
	return string(b)
}

// TestTaskDefer_Until pins the snooze path: `task defer <slug> --until <date>`
// moves the task to deferred/ AND records revisit_at; a bare defer (no --until)
// is unchanged (no revisit_at written).
func TestTaskDefer_Until(t *testing.T) {
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "defer", "alpha", "--until", "2026-09-01")

	// Moved out of ready-to-start, into deferred/.
	notExist(t, filepath.Join(root, "tasks", "ready-to-start", "alpha.md"))
	got := readTaskFile(t, root, "deferred", "alpha.md")
	if !strings.Contains(got, `revisit_at: "2026-09-01"`) {
		t.Errorf("deferred task should carry revisit_at:\n%s", got)
	}
	if !strings.Contains(got, "status: deferred") {
		t.Errorf("deferred task frontmatter status should be deferred:\n%s", got)
	}
}

// TestTaskDefer_NoUntilUnchanged pins the indefinite defer: without --until the
// task moves to deferred/ with no revisit_at field (exactly as before).
func TestTaskDefer_NoUntilUnchanged(t *testing.T) {
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "defer", "alpha")
	got := readTaskFile(t, root, "deferred", "alpha.md")
	if strings.Contains(got, "revisit_at") {
		t.Errorf("a bare defer must not write revisit_at:\n%s", got)
	}
}

// TestTaskDefer_BadDateExit11 pins the up-front validation: a malformed --until
// is a loud validation error (exit 11) BEFORE anything moves — the task stays put.
func TestTaskDefer_BadDateExit11(t *testing.T) {
	root := setupRepo(t)
	_, err := runRootRC(t, "-C", root, "task", "defer", "alpha", "--until", "next-week")
	if err == nil {
		t.Fatal("a bad --until date should fail")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit 11 for a bad date, got %d (%v)", ExitCode(err), err)
	}
	// Nothing moved: alpha is still in ready-to-start, not deferred.
	if _, statErr := os.Stat(filepath.Join(root, "tasks", "ready-to-start", "alpha.md")); statErr != nil {
		t.Errorf("a bad date must not move the task: %v", statErr)
	}
	notExist(t, filepath.Join(root, "tasks", "deferred", "alpha.md"))
}

// TestTaskDefer_UntilDryRun pins the preview: --dry-run previews the snooze
// without writing — the file stays in its original bucket.
func TestTaskDefer_UntilDryRun(t *testing.T) {
	root := setupRepo(t)
	out, err := runRootRC(t, "-C", root, "task", "defer", "alpha", "--until", "2026-09-01", "--dry-run")
	if err != nil {
		t.Fatalf("dry-run defer failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "would move") {
		t.Errorf("dry-run should preview the move, got:\n%s", out)
	}
	// The preview must actually confirm the would-be snooze date — not be
	// indistinguishable from a bare `defer --dry-run`.
	if !strings.Contains(out, "2026-09-01") {
		t.Errorf("dry-run should preview the revisit date, got:\n%s", out)
	}
	// Nothing written: still in ready-to-start, not deferred.
	if _, statErr := os.Stat(filepath.Join(root, "tasks", "ready-to-start", "alpha.md")); statErr != nil {
		t.Errorf("dry-run must not move the task: %v", statErr)
	}
	notExist(t, filepath.Join(root, "tasks", "deferred", "alpha.md"))
}

// TestTaskDefer_UntilDryRunJSON pins the would-be revisit_at in the --json move
// report on a dry-run preview (nothing written, but the date is surfaced).
func TestTaskDefer_UntilDryRunJSON(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "--json", "task", "defer", "alpha", "--until", "2026-09-01", "--dry-run")
	if !strings.Contains(out, `"revisit_at":"2026-09-01"`) {
		t.Errorf("dry-run --json should carry the would-be revisit_at:\n%s", out)
	}
	notExist(t, filepath.Join(root, "tasks", "deferred", "alpha.md"))
}

// TestTaskDefer_MultipleSlugsUntil pins the batch-snooze contract: every slug
// passed to a single `defer ... --until` gets the same revisit_at.
func TestTaskDefer_MultipleSlugsUntil(t *testing.T) {
	root := setupRepo(t) // alpha (ready-to-start) + beta (in-progress)
	runRoot(t, "-C", root, "task", "defer", "alpha", "beta", "--until", "2026-09-01")
	for _, name := range []string{"alpha", "beta"} {
		got := readTaskFile(t, root, "deferred", name+".md")
		if !strings.Contains(got, `revisit_at: "2026-09-01"`) {
			t.Errorf("%s should carry the batch revisit date:\n%s", name, got)
		}
	}
}

// TestTaskDefer_RedeferKeepsDate pins the load-bearing from==to no-op: re-deferring
// an already-deferred task without --until KEEPS the existing revisit_at instead of
// wiping it (a bare re-defer is a Move no-op, and until=="" skips SetFields).
func TestTaskDefer_RedeferKeepsDate(t *testing.T) {
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "defer", "alpha", "--until", "2026-09-01")
	runRoot(t, "-C", root, "task", "defer", "alpha") // bare re-defer (gate closed in tests → no prompt)
	got := readTaskFile(t, root, "deferred", "alpha.md")
	if !strings.Contains(got, `revisit_at: "2026-09-01"`) {
		t.Errorf("bare re-defer must keep the existing revisit_at:\n%s", got)
	}
}

// TestTaskDefer_UntilJSONRealRun pins that a REAL (non-dry-run) --json defer
// surfaces the recorded revisit_at, sourced from the written task (not just the
// dry-run preview path).
func TestTaskDefer_UntilJSONRealRun(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "--json", "task", "defer", "alpha", "--until", "2026-09-01")
	if !strings.Contains(out, `"revisit_at":"2026-09-01"`) {
		t.Errorf("real-run --json defer should carry revisit_at:\n%s", out)
	}
	if got := readTaskFile(t, root, "deferred", "alpha.md"); !strings.Contains(got, `revisit_at: "2026-09-01"`) {
		t.Errorf("real run should persist revisit_at:\n%s", got)
	}
}

// TestTaskDefer_InteractiveRelativeDate pins the relative-answer path end to end:
// a "1w" typed at the prompt is computed against the command's clock (time.Now())
// and persisted. The expected date is captured around Execute so a midnight tick
// between the two clock reads can't flake it.
func TestTaskDefer_InteractiveRelativeDate(t *testing.T) {
	root := setupRepo(t)
	f := &prompt.Fake{TextAnswers: []string{"1w"}}
	var out bytes.Buffer
	cmd := newDeferCmd(deferApp(root, f, &out))
	cmd.SetArgs([]string{"alpha"})
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	before := time.Now().AddDate(0, 0, 7).Format(time.DateOnly)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("interactive relative defer: %v\n%s", err, out.String())
	}
	after := time.Now().AddDate(0, 0, 7).Format(time.DateOnly)

	got := readTaskFile(t, root, "deferred", "alpha.md")
	if !strings.Contains(got, `revisit_at: "`+before+`"`) && !strings.Contains(got, `revisit_at: "`+after+`"`) {
		t.Errorf("relative '1w' should persist %s (or %s across a midnight tick):\n%s", before, after, got)
	}
}

// TestTaskDefer_RevisitClearedOnResume pins that revisit_at is dropped when a task
// leaves deferred (`task next`/`task ready` to resume) — a stale snooze date must
// not ride along onto a now-active task.
func TestTaskDefer_RevisitClearedOnResume(t *testing.T) {
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "defer", "alpha", "--until", "2026-09-01")
	if got := readTaskFile(t, root, "deferred", "alpha.md"); !strings.Contains(got, "revisit_at") {
		t.Fatalf("precondition: deferred alpha should carry revisit_at:\n%s", got)
	}
	runRoot(t, "-C", root, "task", "next", "alpha") // deferred -> next-up
	got := readTaskFile(t, root, "next-up", "alpha.md")
	if strings.Contains(got, "revisit_at") {
		t.Errorf("leaving deferred should clear revisit_at:\n%s", got)
	}
}

// TestTaskTransition_DeprecatedAliases pins back-compat: the old promote/demote
// verbs still move tasks (hidden + deprecation-warned) so existing scripts/muscle
// memory don't break after the rename to next/ready.
func TestTaskTransition_DeprecatedAliases(t *testing.T) {
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "promote", "alpha") // alias for `next`
	if _, err := os.Stat(filepath.Join(root, "tasks", "next-up", "alpha.md")); err != nil {
		t.Errorf("`task promote` alias should still move to next-up: %v", err)
	}
	runRoot(t, "-C", root, "task", "demote", "alpha") // alias for `ready`
	if _, err := os.Stat(filepath.Join(root, "tasks", "ready-to-start", "alpha.md")); err != nil {
		t.Errorf("`task demote` alias should still move to ready-to-start: %v", err)
	}
}

// deferApp builds an App wired for an interactive (TTY-equivalent) defer: a real
// FS-backed service over root, the gate open, and a scripted prompter. It mirrors
// the hand-built App pattern the epic/edit picker tests use.
func deferApp(root string, p prompt.Prompter, out *bytes.Buffer) *App {
	return &App{
		Out: out, ErrOut: out, In: strings.NewReader(""),
		Style:  render.NewStyle(false),
		Gate:   prompt.NewGate(true),
		Prompt: p,
		Svc:    core.NewService(store.NewFS(root)),
	}
}

// TestTaskDefer_InteractivePromptsForDate pins the new snooze UX: a defer on a TTY
// without --until brings up a separate revisit-date prompt and records what the
// user enters (here an absolute date).
func TestTaskDefer_InteractivePromptsForDate(t *testing.T) {
	root := setupRepo(t)
	f := &prompt.Fake{TextAnswers: []string{"2026-09-01"}}
	var out bytes.Buffer
	cmd := newDeferCmd(deferApp(root, f, &out))
	cmd.SetArgs([]string{"alpha"})
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("interactive defer: %v\n%s", err, out.String())
	}
	if len(f.Asked) != 1 || !strings.Contains(f.Asked[0], "Revisit date") {
		t.Errorf("expected one revisit-date prompt, got %v", f.Asked)
	}
	got := readTaskFile(t, root, "deferred", "alpha.md")
	if !strings.Contains(got, `revisit_at: "2026-09-01"`) {
		t.Errorf("interactive defer should record the prompted date:\n%s", got)
	}
}

// TestTaskDefer_InteractiveBlankSkipsDate pins that leaving the prompt blank parks
// the task indefinitely (no revisit_at), so the snooze stays opt-in.
func TestTaskDefer_InteractiveBlankSkipsDate(t *testing.T) {
	root := setupRepo(t)
	f := &prompt.Fake{TextAnswers: []string{"  "}} // blank/whitespace = skip
	var out bytes.Buffer
	cmd := newDeferCmd(deferApp(root, f, &out))
	cmd.SetArgs([]string{"alpha"})
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("interactive defer (blank): %v\n%s", err, out.String())
	}
	if got := readTaskFile(t, root, "deferred", "alpha.md"); strings.Contains(got, "revisit_at") {
		t.Errorf("a blank revisit prompt must not write revisit_at:\n%s", got)
	}
}

// TestTaskDefer_FlagSkipsInteractivePrompt pins that an explicit --until bypasses
// the prompt entirely (the flag-twin contract): the date is taken from the flag,
// not asked for.
func TestTaskDefer_FlagSkipsInteractivePrompt(t *testing.T) {
	root := setupRepo(t)
	f := &prompt.Fake{} // empty: any prompt call would error
	var out bytes.Buffer
	cmd := newDeferCmd(deferApp(root, f, &out))
	cmd.SetArgs([]string{"alpha", "--until", "2026-09-01"})
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("defer --until on a TTY should not prompt: %v\n%s", err, out.String())
	}
	if len(f.Asked) != 0 {
		t.Errorf("--until should bypass the revisit prompt, got prompts %v", f.Asked)
	}
	if got := readTaskFile(t, root, "deferred", "alpha.md"); !strings.Contains(got, `revisit_at: "2026-09-01"`) {
		t.Errorf("--until value should be recorded:\n%s", got)
	}
}

// TestTaskSet_RevisitAt pins that revisit_at rides the generic `task set` path for
// free: --set writes the field (date-validated) and --unset clears it.
func TestTaskSet_RevisitAt(t *testing.T) {
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "set", "alpha", "--set", "revisit_at=2026-09-01")
	got := readTaskFile(t, root, "ready-to-start", "alpha.md")
	if !strings.Contains(got, `revisit_at: "2026-09-01"`) {
		t.Errorf("task set --set revisit_at should write the field:\n%s", got)
	}

	// A bad date is rejected (the field routes through ValidateDate).
	if _, err := runRootRC(t, "-C", root, "task", "set", "alpha", "--set", "revisit_at=someday"); err == nil {
		t.Error("task set revisit_at=<bad date> should fail validation")
	} else if ExitCode(err) != 11 {
		t.Errorf("want exit 11 for a bad revisit_at, got %d", ExitCode(err))
	}

	// --unset clears it.
	runRoot(t, "-C", root, "task", "set", "alpha", "--unset", "revisit_at")
	got = readTaskFile(t, root, "ready-to-start", "alpha.md")
	if strings.Contains(got, "revisit_at") {
		t.Errorf("task set --unset revisit_at should remove the field:\n%s", got)
	}
}

// writeDeferred writes a deferred task fixture (optionally with a revisit_at) into
// root's deferred bucket — for the `task list --revisit-due` filter tests.
func writeDeferred(t *testing.T, root, name, revisitAt string) {
	t.Helper()
	fm := "---\nstatus: deferred\ndescription: " + name + "\ntags: [seed]\n"
	if revisitAt != "" {
		fm += "revisit_at: \"" + revisitAt + "\"\n"
	}
	fm += "---\n# " + name + "\n"
	dir := filepath.Join(root, "tasks", "deferred")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".md"), []byte(fm), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestTaskList_RevisitDue pins the focused triage query: --revisit-due lists only
// deferred tasks whose revisit date has arrived (past date = always due; future =
// never; no date = never), excludes active tasks, and feeds -q/--json. Past/future
// dates keep it robust against the wall clock without injecting one.
func TestTaskList_RevisitDue(t *testing.T) {
	root := setupRepo(t)                            // alpha (ready-to-start) + beta (in-progress)
	writeDeferred(t, root, "overdue", "2020-01-01") // due
	writeDeferred(t, root, "later", "2099-01-01")   // not due
	writeDeferred(t, root, "parked", "")            // indefinite, not due

	// -q emits just the due slugs (ready for `| xargs task next`).
	out := runRoot(t, "-C", root, "task", "list", "--revisit-due", "-q")
	if !strings.Contains(out, "overdue") {
		t.Errorf("--revisit-due should list the overdue task:\n%s", out)
	}
	for _, excluded := range []string{"later", "parked", "alpha", "beta"} {
		if strings.Contains(out, excluded) {
			t.Errorf("--revisit-due must exclude %q (not deferred-and-due):\n%s", excluded, out)
		}
	}

	// --json returns the standard tasks envelope, narrowed to the due task.
	js := runRoot(t, "-C", root, "--json", "task", "list", "--revisit-due")
	if !strings.Contains(js, `"slug":"overdue"`) {
		t.Errorf("--revisit-due --json should carry the due task:\n%s", js)
	}
	if strings.Contains(js, `"slug":"later"`) || strings.Contains(js, `"slug":"parked"`) {
		t.Errorf("--revisit-due --json should exclude not-due deferred tasks:\n%s", js)
	}
}
