package cli

import (
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
)

// Selection must establish identity without requiring a typed read of the broken
// value: task set validates the repaired document, not the original document.
func TestTaskSetRepairsMalformedTypedFields(t *testing.T) {
	for _, tc := range []struct {
		field, broken, repaired string
		args                    []string
	}{
		{"tier", "tier: three", "tier: 2", []string{"--tier", "2"}},
		{"tags", "tags: {a: b}", "tags: [go, core]", []string{"--tags", "go,core"}},
		{"priority", "priority: [x]", "priority: high", []string{"--priority", "high"}},
	} {
		for _, dry := range []bool{false, true} {
			t.Run(tc.field+"/"+map[bool]string{false: "write", true: "preview"}[dry], func(t *testing.T) {
				root := setupRepo(t)
				path := filepath.Join(root, "tasks", testutil.TaskID("alpha")+"-alpha.md")
				original := "---\nid: " + testutil.TaskID("alpha") + "\nstatus: ready-to-start\ndescription: alpha\ntags: [seed]\n---\n# Alpha\n"
				if tc.field == "tags" {
					original = strings.Replace(original, "tags: [seed]", tc.broken, 1)
				} else {
					original = strings.Replace(original, "\n---\n", "\n"+tc.broken+"\n---\n", 1)
				}
				testutil.Write(t, path, original)
				before := testutil.SnapshotTree(t, root)
				args := append([]string{"task", "set", "ALP"}, tc.args...)
				if dry {
					args = append(args, "--dry-run")
				}
				out, errOut, err := runIn(t, root, args...)
				if err != nil {
					t.Fatalf("field repair failed: %v\nstdout=%s\nstderr=%s", err, out, errOut)
				}
				if dry {
					if !strings.Contains(out, "would update") || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
						t.Fatalf("preview did not validate the repair without effects: %s", out)
					}
					return
				}
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(content), tc.repaired) || strings.Contains(string(content), tc.broken) {
					t.Fatalf("malformed field was not repaired:\n%s", content)
				}
				if _, _, err := runIn(t, root, "task", "info", testutil.TaskID("alpha"), "--json"); err != nil {
					t.Fatalf("repair did not produce a readable task: %v", err)
				}
				after := testutil.SnapshotTree(t, root)
				delete(before, filepath.Join("tasks", filepath.Base(path)))
				delete(after, filepath.Join("tasks", filepath.Base(path)))
				if !maps.Equal(before, after) {
					t.Fatal("repair changed another file")
				}
			})
		}
	}
}

func TestTaskSetPreservesPreexistingCorruptionAttribution(t *testing.T) {
	for _, corruption := range []string{"tier: three", "tier: 2\ntier: 3"} {
		for _, dry := range []bool{false, true} {
			t.Run(corruption+"/"+map[bool]string{false: "write", true: "preview"}[dry], func(t *testing.T) {
				root := setupRepo(t)
				path := filepath.Join(root, "tasks", testutil.TaskID("alpha")+"-alpha.md")
				testutil.Write(t, path, "---\nstatus: ready-to-start\ndescription: alpha\ntags: [seed]\n"+corruption+"\n---\n# Alpha\n")
				before := testutil.SnapshotTree(t, root)
				args := []string{"task", "set", "alpha", "--priority", "high"}
				if dry {
					args = append(args, "--dry-run")
				}
				_, _, err := runIn(t, root, args...)
				if err == nil || ExitCode(err) != 11 || !strings.Contains(err.Error(), "already has malformed frontmatter (not caused by this update)") {
					t.Fatalf("preexisting corruption was not attributed to the file: %v", err)
				}
				if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatal("rejected repair changed files")
				}
			})
		}
	}
}

func TestTaskBodyMutationsStillRefuseMalformedFrontmatter(t *testing.T) {
	for _, verb := range []string{"criterion", "append", "replace"} {
		for _, dry := range []bool{false, true} {
			t.Run(verb+"/"+map[bool]string{false: "write", true: "preview"}[dry], func(t *testing.T) {
				root := setupRepoWithAC(t)
				path := filepath.Join(root, "tasks", testutil.TaskID("alpha")+"-alpha.md")
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				testutil.Write(t, path, strings.Replace(string(content), "\n---\n", "\ntier: three\n---\n", 1))
				before := testutil.SnapshotTree(t, root)
				args := map[string][]string{
					"criterion": {"task", "ac", "alpha", "--check", "1"},
					"append":    {"task", "append", "alpha", "--body", "note"},
					"replace":   {"task", "set", "alpha", "--body", "# Replacement"},
				}[verb]
				if dry {
					args = append(args, "--dry-run")
				}
				_, _, err = runIn(t, root, args...)
				if err == nil || ExitCode(err) != 11 || !strings.Contains(err.Error(), "malformed frontmatter") {
					t.Fatalf("body write no longer requires a reloadable document: %v", err)
				}
				if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatal("malformed body write changed files")
				}
			})
		}
	}
}
