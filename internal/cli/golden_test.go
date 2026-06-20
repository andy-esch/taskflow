package cli

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// updateGolden regenerates the committed snapshots: `go test ./internal/cli -update`.
var updateGolden = flag.Bool("update", false, "regenerate golden files under testdata/golden/")

// assertGolden compares got against testdata/golden/<name>.golden, rewriting it
// under -update. A tiny dep-free helper (the repo's "no library for a 15-line job"
// ethos): the byte-stable machine contract is exactly what these snapshots lock,
// so any unintended drift in a --json envelope / csv / schema trips a diff.
func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name+".golden")
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v\n(regenerate with: go test ./internal/cli -update)", path, err)
	}
	if got != string(want) {
		t.Errorf("output drift vs %s — regenerate with -update if intended.\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}
