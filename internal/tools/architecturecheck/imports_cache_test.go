package architecturecheck

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Exercise the ordinary cacheable go test route, not -count=1: the checker
// must register the production files that its child go list process reads.
func TestImportMapCacheInvalidatesProductionDrift(t *testing.T) {
	checker := readFile(t, "imports_test.go")
	for _, mutation := range []struct {
		name, path, content, want string
	}{
		{"edge", "internal/a/a.go", "package a\nimport _ \"example.test/mapfixture/internal/c\"\n", "+ edge a -> c"},
		{"new leaf", "internal/newleaf/leaf.go", "package newleaf\n", "+ package newleaf"},
		{"removed leaf", "internal/c/c.go", "", "- package c"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			root := t.TempDir()
			writeImportFixture(t, root, map[string]string{
				"go.mod":                          "module example.test/mapfixture\n\ngo 1.26\n",
				"docs/architecture/boundaries.md": markedImportMap("a -> b\nb -> -\nc -> -"),
				"internal/a/a.go":                 "package a\nimport _ \"example.test/mapfixture/internal/b\"\n",
				"internal/b/b.go":                 "package b\n",
				"internal/c/c.go":                 "package c\n",
				"internal/tools/architecturecheck/imports_test.go": checker,
				"internal/tools/architecturecheck/read_test.go": `package architecturecheck
import ("os"; "testing")
func readFile(t *testing.T, path string) string {
 t.Helper()
 data, err := os.ReadFile(path)
 if err != nil { t.Fatal(err) }
 return string(data)
}
`,
			})
			// Go refuses caching recently opened regular files. Age the fixture
			// explicitly rather than adding wall-clock sleeps to the regression.
			old := time.Unix(1_000_000_000, 0)
			if err := filepath.WalkDir(root, func(path string, _ os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				return os.Chtimes(path, old, old)
			}); err != nil {
				t.Fatal(err)
			}
			args := []string{"test", "-race", "-run", "^TestArchitectureImportMap$", "./internal/tools/architecturecheck"}
			if output, err := fixtureGo(root, args...); err != nil {
				t.Fatalf("warm cache: %v\n%s", err, output)
			}
			if output, err := fixtureGo(root, args...); err != nil || !strings.Contains(output, "(cached)") {
				t.Fatalf("cache control did not hit: %v\n%s", err, output)
			}
			if mutation.content == "" {
				if err := os.Remove(filepath.Join(root, mutation.path)); err != nil {
					t.Fatal(err)
				}
			} else {
				writeImportFixture(t, root, map[string]string{mutation.path: mutation.content})
			}
			if output, err := fixtureGo(root, args...); err == nil || !strings.Contains(output, mutation.want) {
				t.Fatalf("production drift reused a pass or failed for the wrong reason: %v\n%s; want %q", err, output, mutation.want)
			}
		})
	}
}

func writeImportFixture(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for path, content := range files {
		path = filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func fixtureGo(root string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=readonly")
	output, err := cmd.CombinedOutput()
	return string(output), err
}
