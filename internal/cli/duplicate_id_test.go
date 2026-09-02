package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDuplicateStableID_ExitsAmbiguousNamingTheFiles is the whole chain the store
// guard's conflict/failure split exists for, seen from where a user meets it: two
// files claiming one stable id must exit 13 (ambiguous) naming both, not exit 14
// with "changed on disk; retry" after four backoffs.
//
// Exit code is the load-bearing assertion — a script or agent branches on it, and
// 14 tells it to try again forever at something no retry can fix.
func TestDuplicateStableID_ExitsAmbiguousNamingTheFiles(t *testing.T) {
	root := t.TempDir()
	tasks := filepath.Join(root, "tasks")
	if err := os.MkdirAll(tasks, 0o755); err != nil {
		t.Fatal(err)
	}
	const dup = "6fjangd7kvc1"
	front := "---\nid: " + dup + "\nstatus: ready-to-start\nepic: e1\ndescription: d\ntags: [a]\n---\n"
	for name, body := range map[string]string{dup + "-alpha.md": "# A\n", dup + "-beta.md": "# B\n"} {
		if err := os.WriteFile(filepath.Join(tasks, name), []byte(front+body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	_, err := runRootStreams(t, "-C", root, "task", "set", "alpha", "--priority", "low")
	if err == nil {
		t.Fatal("writing through a duplicate stable id must fail")
	}
	if got := ExitCode(err); got != 13 {
		t.Errorf("exit code = %d, want 13 (ambiguous); 14 would tell a caller to retry", got)
	}
	msg := err.Error()
	if strings.Contains(msg, "retry") {
		t.Errorf("a duplicate id must not be presented as retryable: %s", msg)
	}
	for _, want := range []string{"alpha", "beta"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the error should name the colliding file %q: %s", want, msg)
		}
	}
}
