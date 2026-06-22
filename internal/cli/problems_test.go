package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskList_ReportsBadFileButShowsGood(t *testing.T) {
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
	write("tasks/ready-to-start/good.md", "---\nstatus: ready-to-start\ndescription: ok\n---\n# Good\n")
	write("tasks/ready-to-start/bad.md", "---\nstatus: ready-to-start\ntags: a,b,c\n---\n# Bad\n")

	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "list"})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected a non-zero result for the unreadable file")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit 11, got %d", ExitCode(err))
	}
	s := out.String()
	if !bytes.Contains([]byte(s), []byte("good")) {
		t.Errorf("the good task should still be listed:\n%s", s)
	}
	if !bytes.Contains([]byte(s), []byte("tags")) || !bytes.Contains([]byte(s), []byte("bad.md")) {
		t.Errorf("the bad file should be reported with guidance:\n%s", s)
	}
}
