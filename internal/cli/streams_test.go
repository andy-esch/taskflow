package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
)

// repoWithUnreadableFile builds a tree holding one good task and one file whose
// frontmatter cannot be parsed — the shape that makes a command produce payload
// AND diagnostics in the same run, which is the only shape that can prove which
// stream each went to.
func repoWithUnreadableFile(t *testing.T) (root, goodID, goodSlug string) {
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
	goodID = testutil.TaskID("good")
	write("tasks/"+goodID+"-good.md", "---\nid: "+goodID+"\nstatus: ready-to-start\ndescription: ok\n---\n# Good\n")
	write("tasks/"+testutil.TaskID("bad")+"-bad.md", "---\nstatus: ready-to-start\ntags: a,b,c\n---\n# Bad\n")
	return root, goodID, "good"
}

// TestStreamDiscipline_NameModeStdoutCarriesOnlyNames pins the contract problems.go
// documents: diagnostics go to stderr so that a script capturing stdout gets only
// the payload. `-o name` is the sharpest case — its entire output contract is "one
// name per line", so a diagnostic landing on stdout is not cosmetic, it is a
// corrupted result for `| xargs`.
//
// This is the assertion the package could not make before: with one buffer behind
// both streams, moving render.ProblemsHuman from app.ErrOut to app.Out left every
// test green.
func TestStreamDiscipline_NameModeStdoutCarriesOnlyNames(t *testing.T) {
	root, _, goodSlug := repoWithUnreadableFile(t)

	res, err := runRootStreams(t, "-C", root, "task", "list", "-o", "name")
	if err == nil {
		t.Fatal("an unreadable file should still make the command exit non-zero")
	}
	if got := ExitCode(err); got != 11 {
		t.Errorf("exit code = %d, want 11", got)
	}

	// stdout: exactly the names, nothing else.
	for _, line := range strings.Split(strings.TrimSpace(res.Out), "\n") {
		if line != goodSlug {
			t.Errorf("stdout must carry names alone, got line %q:\n%s", line, res.Out)
		}
	}
	// stderr: the diagnostic, naming the file and the offending field.
	if !strings.Contains(res.Err, "bad.md") || !strings.Contains(res.Err, "tags") {
		t.Errorf("the unreadable file should be reported on stderr:\n%s", res.Err)
	}
	// Both were produced — otherwise the test proves nothing about routing.
	if res.Out == "" || res.Err == "" {
		t.Fatalf("setup: expected output on both streams, got out=%q err=%q", res.Out, res.Err)
	}
}

// TestStreamDiscipline_JSONStdoutStaysParseable covers the other consumer, whose
// contract is enforced differently: JSON mode returns before ProblemsHuman is
// reached and embeds the problems in the envelope instead (renderList's comment
// states this), so the risk here is not a misrouted diagnostic but ANY stray
// write to stdout. Unmarshalling res.Out is the total assertion for that —
// nothing else may be on the stream.
func TestStreamDiscipline_JSONStdoutStaysParseable(t *testing.T) {
	root, goodID, _ := repoWithUnreadableFile(t)

	res, err := runRootStreams(t, "-C", root, "task", "list", "--json")
	if err == nil {
		t.Fatal("an unreadable file should still make the command exit non-zero")
	}

	var envelope struct {
		SchemaVersion string           `json:"schema_version"`
		Tasks         []map[string]any `json:"tasks"`
	}
	if jsonErr := json.Unmarshal([]byte(res.Out), &envelope); jsonErr != nil {
		t.Fatalf("stdout must be the JSON envelope alone: %v\nstdout:\n%s", jsonErr, res.Out)
	}
	if envelope.SchemaVersion == "" {
		t.Errorf("envelope lost its schema_version:\n%s", res.Out)
	}
	if len(envelope.Tasks) != 1 || envelope.Tasks[0]["id"] != goodID {
		t.Errorf("the readable task should still be in the payload:\n%s", res.Out)
	}
}

// TestStreamDiscipline_HumanListSeparatesPayloadFromDiagnostics covers the
// default human mode, where the merged view is what a person sees but the split
// still has to hold underneath.
func TestStreamDiscipline_HumanListSeparatesPayloadFromDiagnostics(t *testing.T) {
	root, _, _ := repoWithUnreadableFile(t)

	res, _ := runRootStreams(t, "-C", root, "task", "list")

	if !strings.Contains(res.Out, "good") {
		t.Errorf("the readable task belongs on stdout:\n%s", res.Out)
	}
	if strings.Contains(res.Out, "bad.md") {
		t.Errorf("the file diagnostic must not be on stdout:\n%s", res.Out)
	}
	if !strings.Contains(res.Err, "bad.md") {
		t.Errorf("the file diagnostic belongs on stderr:\n%s", res.Err)
	}
	// The merged view still shows both, so tests that don't care keep working.
	if !strings.Contains(res.Merged, "good") || !strings.Contains(res.Merged, "bad.md") {
		t.Errorf("the merged view should still carry both:\n%s", res.Merged)
	}
}
