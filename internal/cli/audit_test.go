package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func setupAuditRepo(t *testing.T) string {
	t.Helper()
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
	write("tasks/.gitkeep", "") // so Discover anchors here
	audit := func(bucket, name, content string) {
		p, out := testutil.AuditFixture(root, bucket, name, content)
		testutil.Write(t, p, out)
	}
	audit("open", "o.md", "---\nid: "+testutil.TaskID("o")+"\nbucket: open\narea: dispatcher\n---\n#### H1. t  · **Status:** open\n")
	audit("closed", "c.md", "---\nid: "+testutil.TaskID("c")+"\nbucket: closed\narea: web\n---\n#### M1. t  · **Status:** fixed\n")
	return root
}

// TestAuditAppend_JSON pins the `audit_mutation` --json envelope (the contract the
// schema_version 1.20 bump is for): a parseable envelope with the reloaded audit,
// dry_run=false, and the echoed resulting body.
func TestAuditAppend_JSON(t *testing.T) {
	root := setupAuditRepo(t)
	out := runRoot(t, "-C", root, "--json", "audit", "append", "o", "--body", "#### M9. new  · **Status:** open")
	var env struct {
		SchemaVersion string `json:"schema_version"`
		DryRun        bool   `json:"dry_run"`
		Body          string `json:"body"`
		Audit         struct {
			Slug   string `json:"slug"`
			Bucket string `json:"bucket"`
		} `json:"audit"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("audit append --json is not a parseable envelope: %v\n%s", err, out)
	}
	if env.SchemaVersion == "" || env.Audit.Slug != "o" || env.Audit.Bucket != "open" {
		t.Errorf("audit append --json envelope wrong:\n%s", out)
	}
	if env.DryRun {
		t.Error("a real append should report dry_run=false")
	}
	if !strings.Contains(env.Body, "#### M9. new") {
		t.Errorf("append --json should echo the resulting body:\n%s", out)
	}
}

// --dry-run previews an audit append without writing.
func TestAuditAppend_DryRun_NoWrite(t *testing.T) {
	root := setupAuditRepo(t)
	p := filepath.Join(root, "audits", testutil.TaskID("o")+"-o.md")
	before, _ := os.ReadFile(p)
	runRoot(t, "-C", root, "--dry-run", "audit", "append", "o", "--body", "#### NOPE.  · **Status:** open")
	if after, _ := os.ReadFile(p); !bytes.Equal(before, after) {
		t.Error("--dry-run audit append must not write")
	}
}

func TestAuditAppendKeepsFreshManagedCandidateSectionValid(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "audit", "new", "append-narrative", "--date", "2026-09-20")
	runRoot(t, "-C", root, "audit", "append", "2026-09-20-append-narrative",
		"--body", "## Progress\n\nNarrative update.")
	runRoot(t, "-C", root, "audit", "lint", "2026-09-20-append-narrative")

	b, err := os.ReadFile(auditPath(t, root, "2026-09-20-append-narrative"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if strings.Index(got, "## Progress") > strings.Index(got, "## Candidate tasks") {
		t.Fatalf("audit append put narrative inside/after the managed projection:\n%s", got)
	}
}

// Empty append input is a clean validation error, not an empty write.
func TestAuditAppend_Empty_Errors(t *testing.T) {
	root := setupAuditRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "audit", "append", "o", "--body", "   "})
	if err := cmd.Execute(); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("empty audit append should wrap ErrValidation (exit 11), got %v", err)
	}
}

// Passing both --body and --body-file to `audit append` is a usage error, not a
// silent precedence pick — mirroring `task append`/`task new`.
func TestAuditAppend_BodyAndBodyFile_Exclusive(t *testing.T) {
	root := setupAuditRepo(t)
	if _, err := runRootRC(t, "-C", root, "audit", "append", "o", "--body", "x", "--body-file", "-"); err == nil {
		t.Fatal("`audit append --body … --body-file -` should be rejected (mutually exclusive)")
	}
}

// `audit edit --dry-run` is rejected (it's interactive, no preview) — a safety flag
// must never be a silent no-op.
func TestAuditEdit_RejectsDryRun(t *testing.T) {
	root := setupAuditRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "--dry-run", "audit", "edit", "o"})
	if err := cmd.Execute(); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("`audit edit --dry-run` should be rejected with ErrValidation, got %v", err)
	}
}

func TestAuditList_DefaultsToOpen(t *testing.T) {
	root := setupAuditRepo(t)
	out := runRoot(t, "-C", root, "audit", "list", "--json")

	var got struct {
		Audits []struct {
			Slug   string `json:"slug"`
			Bucket string `json:"bucket"`
		} `json:"audits"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("bad json: %v\n%s", err, out)
	}
	if len(got.Audits) != 1 || got.Audits[0].Slug != "o" || got.Audits[0].Bucket != "open" {
		t.Errorf("default should be open only: %+v", got.Audits)
	}
}

func TestAuditList_All(t *testing.T) {
	root := setupAuditRepo(t)
	out := runRoot(t, "-C", root, "audit", "list", "--all")
	if !strings.Contains(out, "o") || !strings.Contains(out, "c") {
		t.Errorf("--all should show both buckets:\n%s", out)
	}
}

// Closing an audit is an IN-PLACE frontmatter edit under the flat layout: the file
// path never changes, only its `bucket:` flips open→closed.
func TestAuditClose_ChangesBucketInPlace(t *testing.T) {
	root := setupAuditRepo(t)
	// A clean audit (no open findings) closes fine — `o` carries an open finding
	// and is covered by TestAuditClose_RejectsOpenFindings below.
	p := filepath.Join(root, "audits", testutil.TaskID("clean")+"-clean.md")
	mustWrite(t, p,
		"---\nid: "+testutil.TaskID("clean")+"\nbucket: open\narea: clean\n---\n#### H1. t  · **Status:** fixed\n")
	out := runRoot(t, "-C", root, "audit", "close", "clean")
	if !strings.Contains(out, "clean -> closed") {
		t.Errorf("unexpected output: %q", out)
	}
	if _, err := os.Stat(p); err != nil {
		t.Errorf("close must be in-place — the file must stay at its original flat path: %v", err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "bucket: closed") {
		t.Errorf("close must flip the frontmatter bucket to closed:\n%s", b)
	}
}

// M4 (2026-06-22 audit): closing/deferring an audit that still has open findings
// must be refused (the bucket↔state invariant `audit lint` enforces), with the
// audit left in its original bucket.
func TestAuditClose_RejectsOpenFindings(t *testing.T) {
	root := setupAuditRepo(t) // `o` has H1 open
	if _, err := runRootRC(t, "-C", root, "audit", "close", "o"); err == nil {
		t.Fatal("closing an audit with open findings must be rejected")
	}
	p := filepath.Join(root, "audits", testutil.TaskID("o")+"-o.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("a rejected close must leave the audit file untouched: %v", err)
	}
	if !strings.Contains(string(b), "bucket: open") {
		t.Errorf("a rejected close must leave the frontmatter bucket open:\n%s", b)
	}
}

// TestAuditList_ConflictingFlagsError pins the mutual exclusion: --closed
// --deferred (or --all with either) must error, not silently prefer one.
func TestAuditList_ConflictingFlagsError(t *testing.T) {
	root := setupRepo(t)
	for _, args := range [][]string{
		{"audit", "list", "--closed", "--deferred"},
		{"audit", "list", "--all", "--closed"},
	} {
		var out bytes.Buffer
		cmd := NewRootCmd(strings.NewReader(""), &out, &out)
		cmd.SetArgs(append([]string{"-C", root}, args...))
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		if err := cmd.Execute(); err == nil {
			t.Errorf("%v should error (mutually exclusive flags)", args)
		}
	}
}

// An id-led audit with NO frontmatter bucket still lists (raw bucket) but is flagged by lint.
func TestAuditLint_FlagsMissingBucket(t *testing.T) {
	root := setupAuditRepo(t)
	nb := filepath.Join(root, "audits", testutil.TaskID("2026-06-17-nb")+"-2026-06-17-nb.md")
	if err := os.WriteFile(nb,
		[]byte("---\nid: 6fjangd7kvcb\narea: x\ndate: 2026-06-17\n---\n# x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runRootRC(t, "-C", root, "audit", "lint")
	if err == nil {
		t.Error("audit lint must exit non-zero on a missing frontmatter bucket")
	}
	if !strings.Contains(out, "frontmatter bucket missing or unrecognized") {
		t.Errorf("expected the missing-bucket flag:\n%s", out)
	}
}

// `audit finding --note` writes the `**Resolution:**` paragraph, so the block lands inside
// the right finding by construction rather than by careful typing — the reason the status
// write exists, applied to the sentence beside it. Status and note given together must be
// ONE write: two would leave a window where the finding claims `fixed` with last round's
// explanation.
func TestAuditFinding_StatusAndNoteInOneWrite(t *testing.T) {
	root := setupAuditRepo(t)
	p := filepath.Join(root, "audits", testutil.TaskID("o")+"-o.md")

	if _, err := runRootRC(t, "-C", root, "audit", "finding", "o", "H1",
		"--status", "fixed 2026-08-24", "--note", "Widened the regex; regression test added."); err != nil {
		t.Fatalf("audit finding: %v", err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, "**Status:** fixed 2026-08-24") {
		t.Errorf("status not stamped:\n%s", got)
	}
	if !strings.Contains(got, "**Resolution:** Widened the regex; regression test added.") {
		t.Errorf("note not written:\n%s", got)
	}

	// Re-noting replaces; it must not stack a second label under the first.
	if _, err := runRootRC(t, "-C", root, "audit", "finding", "o", "H1", "--note", "Superseded explanation."); err != nil {
		t.Fatalf("re-note: %v", err)
	}
	b, _ = os.ReadFile(p)
	if n := strings.Count(string(b), "**Resolution:**"); n != 1 {
		t.Errorf("want exactly one resolution block after re-noting, got %d:\n%s", n, b)
	}

	// An empty --note removes it, leaving the status alone.
	if _, err := runRootRC(t, "-C", root, "audit", "finding", "o", "H1", "--note", ""); err != nil {
		t.Fatalf("clear note: %v", err)
	}
	b, _ = os.ReadFile(p)
	if strings.Contains(string(b), "**Resolution:**") {
		t.Errorf("note should have been removed:\n%s", b)
	}
	if !strings.Contains(string(b), "**Status:** fixed 2026-08-24") {
		t.Errorf("clearing the note must not disturb the status:\n%s", b)
	}
}

// A managed candidate row is the finding's optional projection, not a second source of
// truth. One command can create it while changing status, later status changes refresh it,
// and an empty value removes it. All three operations use the same atomic audit-body write.
func TestAuditFinding_ManagedCandidateLifecycle(t *testing.T) {
	root := setupAuditRepo(t)
	p := filepath.Join(root, "audits", testutil.TaskID("o")+"-o.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	managed := strings.TrimRight(string(b), "\n") + "\n\n## Candidate tasks\n\n" +
		domain.CandidateTasksMarkerComment() + "\n"
	if err := os.WriteFile(p, []byte(managed), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := runRootRC(t, "-C", root, "audit", "finding", "o", "H1",
		"--status", "in-progress", "--candidate", "Harden the candidate projection"); err != nil {
		t.Fatalf("combined status + candidate edit: %v", err)
	}
	b, _ = os.ReadFile(p)
	if got := string(b); !strings.Contains(got, "**Status:** in-progress") ||
		!strings.Contains(got, "- ● H1 · in-progress — Harden the candidate projection") {
		t.Fatalf("finding and candidate did not land together:\n%s", got)
	}

	if _, err := runRootRC(t, "-C", root, "audit", "finding", "o", "H1", "--status", "fixed"); err != nil {
		t.Fatalf("status-only candidate sync: %v", err)
	}
	b, _ = os.ReadFile(p)
	if !strings.Contains(string(b), "- ✔ H1 · fixed — Harden the candidate projection") {
		t.Fatalf("status-only edit left the candidate stale:\n%s", b)
	}

	if _, err := runRootRC(t, "-C", root, "audit", "finding", "o", "H1", "--candidate", ""); err != nil {
		t.Fatalf("remove candidate: %v", err)
	}
	b, _ = os.ReadFile(p)
	if strings.Contains(string(b), " H1 · fixed —") {
		t.Fatalf("empty --candidate should remove the row:\n%s", b)
	}
}

func TestAuditFindingNewCreatesCanonicalBlockAndCandidate(t *testing.T) {
	root := setupAuditRepo(t)
	p := filepath.Join(root, "audits", testutil.TaskID("o")+"-o.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	managed := strings.TrimRight(string(b), "\n") + "\n\n## Candidate tasks\n\n" +
		domain.CandidateTasksMarkerComment() + "\n"
	if err := os.WriteFile(p, []byte(managed), 0o644); err != nil {
		t.Fatal(err)
	}

	out := runRoot(t, "-C", root, "audit", "finding", "new", "o", "A second finding",
		"--band", "H", "--file", "internal/a.go:12", "--component", "core",
		"--effort", "s", "--urgency", "soon", "--body", "Evidence from the failing case.",
		"--recommendation", "Add the focused guard.", "--candidate", "Create the repair task")
	if !strings.Contains(out, "created H2 in o") {
		t.Fatalf("human receipt should expose allocated identity:\n%s", out)
	}
	after, _ := os.ReadFile(p)
	for _, want := range []string{
		"#### H2. A second finding · **Status:** open",
		"**File:** internal/a.go:12 | **Component:** core",
		"**Effort:** S · **Urgency:** soon",
		"- ○ H2 · open — Create the repair task",
	} {
		if !strings.Contains(string(after), want) {
			t.Errorf("created audit missing %q:\n%s", want, after)
		}
	}
	if strings.Index(string(after), "#### H2.") > strings.Index(string(after), "## Candidate tasks") {
		t.Fatalf("finding should precede candidate section:\n%s", after)
	}
}

func TestAuditFindingNewJSONDryRunIsCompactAndDoesNotWrite(t *testing.T) {
	root := setupAuditRepo(t)
	p := filepath.Join(root, "audits", testutil.TaskID("o")+"-o.md")
	before, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	out := runRoot(t, "-C", root, "--json", "--dry-run", "audit", "finding", "new", "o", "Preview finding", "--band", "M")
	var env struct {
		SchemaVersion string `json:"schema_version"`
		DryRun        bool   `json:"dry_run"`
		Audit         struct {
			Slug string `json:"slug"`
		} `json:"audit"`
		Finding struct {
			Audit  string `json:"audit"`
			Code   string `json:"code"`
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"finding"`
		Body json.RawMessage `json:"body"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("finding creation JSON: %v\n%s", err, out)
	}
	if env.SchemaVersion == "" || !env.DryRun || env.Audit.Slug != "o" || env.Finding.Audit != "o" ||
		env.Finding.Code != "M1" || env.Finding.Title != "Preview finding" || env.Finding.Status != "open" {
		t.Fatalf("unexpected finding creation receipt: %+v", env)
	}
	if env.Body != nil {
		t.Fatalf("compact finding receipt must not echo the full audit body: %s", env.Body)
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Fatal("dry-run finding creation wrote the audit")
	}
}

func TestAuditFindingNewCandidateRefusalLeavesLegacyAuditUntouched(t *testing.T) {
	root := setupAuditRepo(t)
	p := filepath.Join(root, "audits", testutil.TaskID("o")+"-o.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	legacy := strings.TrimRight(string(b), "\n") + "\n\n## Candidate tasks\n\n- legacy prose\n"
	if err := os.WriteFile(p, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runRootRC(t, "-C", root, "audit", "finding", "new", "o", "Must not land",
		"--band", "H", "--candidate", "Needs managed linkage"); !errors.Is(err, domain.ErrValidation) || ExitCode(err) != 11 {
		t.Fatalf("legacy candidate creation should fail validation, got %v", err)
	}
	after, _ := os.ReadFile(p)
	if string(after) != legacy {
		t.Fatalf("failed finding creation partially wrote the audit:\n%s", after)
	}
}

// Candidate edits travel through TransformAuditBody, whose normalization boundary must
// restore the audit's original line-ending convention after the domain transform. Exercise
// that at the CLI boundary so future candidate-specific writes cannot accidentally bypass it.
func TestAuditFinding_ManagedCandidatePreservesCRLFAndDryRun(t *testing.T) {
	root := setupAuditRepo(t)
	p := filepath.Join(root, "audits", testutil.TaskID("o")+"-o.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	managedLF := strings.TrimRight(string(b), "\n") + "\n\n## Candidate tasks\n\n" +
		domain.CandidateTasksMarkerComment() + "\n"
	managedCRLF := strings.ReplaceAll(managedLF, "\n", "\r\n")
	if err := os.WriteFile(p, []byte(managedCRLF), 0o644); err != nil {
		t.Fatal(err)
	}

	before, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runRootRC(t, "-C", root, "--dry-run", "audit", "finding", "o", "H1",
		"--candidate", "Preserve the audit's line endings"); err != nil {
		t.Fatalf("dry-run candidate edit: %v", err)
	}
	afterDryRun, _ := os.ReadFile(p)
	if !bytes.Equal(before, afterDryRun) {
		t.Fatal("dry-run candidate edit changed the audit")
	}

	if _, err := runRootRC(t, "-C", root, "audit", "finding", "o", "H1",
		"--status", "in-progress", "--candidate", "Preserve the audit's line endings"); err != nil {
		t.Fatalf("real candidate edit: %v", err)
	}
	after, _ := os.ReadFile(p)
	if bareLF := bytes.Count(after, []byte("\n")) - bytes.Count(after, []byte("\r\n")); bareLF != 0 {
		t.Fatalf("candidate edit introduced %d bare LF line endings:\n%q", bareLF, after)
	}
	if !bytes.Contains(after, []byte("- "+domain.FindingStatusGlyph("in-progress")+" H1 · in-progress — Preserve the audit's line endings")) {
		t.Fatalf("candidate row did not land in CRLF audit:\n%s", after)
	}
}

func TestAuditFinding_CandidateRefusesLegacyWithoutWriting(t *testing.T) {
	root := setupAuditRepo(t)
	p := filepath.Join(root, "audits", testutil.TaskID("o")+"-o.md")
	original, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	legacy := strings.TrimRight(string(original), "\n") + "\n\n## Candidate tasks\n\n- ⏳ H1 — handwritten legacy row\n"
	if err := os.WriteFile(p, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	before := []byte(legacy)
	if _, err := runRootRC(t, "-C", root, "audit", "finding", "o", "H1",
		"--status", "fixed", "--candidate", "Do not guess at this legacy section"); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "legacy") {
		t.Fatalf("legacy candidate edit should explain its refusal, got %v", err)
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Fatalf("failed combined edit must be atomic; file changed:\n%s", after)
	}
}

func TestAuditNewFindingAndManagedCandidateRoundTrip(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "audit", "new", "candidate-round-trip", "--date", "2026-09-19")
	runRoot(t, "-C", root, "audit", "finding", "new", "2026-09-19-candidate-round-trip",
		"Exercise the managed projection", "--band", "H", "--body", "Evidence.",
		"--candidate", "Create the focused implementation task")
	runRoot(t, "-C", root, "audit", "lint", "2026-09-19-candidate-round-trip")

	b, err := os.ReadFile(auditPath(t, root, "2026-09-19-candidate-round-trip"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if strings.Index(got, "#### H1. Exercise the managed projection") >
		strings.Index(got, "- ○ H1 · open — Create the focused implementation task") {
		t.Fatalf("finding should remain before its trailing candidate projection:\n%s", got)
	}
}

// Neither flag is a usage error, not a silent no-op: a call that changes nothing is almost
// certainly a mistyped flag name.
func TestAuditFinding_RequiresAFlag(t *testing.T) {
	root := setupAuditRepo(t)
	if _, err := runRootRC(t, "-C", root, "audit", "finding", "o", "H1"); err == nil {
		t.Error("audit finding with no requested field must be rejected")
	}
}

// A note carrying a newline could open a heading or a fence mid-finding, restructuring the
// document — the same corruption class as a desynchronised offset, arriving via content.
func TestAuditFinding_RejectsNewlineInNote(t *testing.T) {
	root := setupAuditRepo(t)
	if _, err := runRootRC(t, "-C", root, "audit", "finding", "o", "H1",
		"--note", "bad\n#### H9. injected"); err == nil {
		t.Error("a newline in --note must be rejected")
	}
}

// --pr is sugar for one canonical decoration. `(PR #12)`, `PR 12`, and `pull/12` read the
// same to a human and differently to grep; the corpus already spells its dates two ways.
func TestAuditFinding_PRSugar(t *testing.T) {
	root := setupAuditRepo(t)
	p := filepath.Join(root, "audits", testutil.TaskID("o")+"-o.md")

	if _, err := runRootRC(t, "-C", root, "audit", "finding", "o", "H1", "--status", "fixed 2026-08-24", "--pr", "12"); err != nil {
		t.Fatalf("audit finding --pr: %v", err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "**Status:** fixed 2026-08-24 (PR #12)") {
		t.Errorf("PR decoration not appended:\n%s", b)
	}
	for _, args := range [][]string{
		{"audit", "finding", "o", "H1", "--pr", "12"},                              // nothing to decorate
		{"audit", "finding", "o", "H1", "--status", "fixed", "--pr", "0"},          // not a PR number
		{"audit", "finding", "o", "H1", "--status", "fixed (PR #9)", "--pr", "12"}, // two references
	} {
		if _, err := runRootRC(t, append([]string{"-C", root}, args...)...); err == nil {
			t.Errorf("%v should be rejected", args)
		}
	}
}
