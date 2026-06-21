package render

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/invopop/jsonschema"
)

// TestSchemaComments_NotStale fails if the committed schema_comments.json drifts
// from the Go doc comments it's generated from (e.g. an envelope comment was
// edited but `go run ./internal/tools/schemacomments` wasn't re-run). It
// regenerates the map the same way the generator does — from the repo root with
// the same base+paths, so the gopath.Join(base, dir) keys match exactly — and
// compares byte-for-byte. CI runs this with source present; locally it's a normal
// test.
func TestSchemaComments_NotStale(t *testing.T) {
	root := repoRoot(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(cwd) }()

	r := new(jsonschema.Reflector)
	for _, dir := range []string{"./internal/cli/render", "./internal/domain"} {
		if err := r.AddGoComments("github.com/andy-esch/taskflow", dir); err != nil {
			t.Fatalf("AddGoComments(%s): %v", dir, err)
		}
	}
	// Floor guard: a silently-empty CommentMap (e.g. AddGoComments broke / a path
	// changed) must fail loudly here rather than false-pass a byte-equal-empty check.
	if len(r.CommentMap) < 20 {
		t.Fatalf("AddGoComments produced only %d comments — generator likely broken, not a real drift", len(r.CommentMap))
	}
	regenerated, err := json.MarshalIndent(r.CommentMap, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(append(regenerated, '\n'), schemaComments) {
		t.Errorf("internal/cli/render/schema_comments.json is stale — run `go run ./internal/tools/schemacomments`")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from the test dir")
		}
		dir = parent
	}
}
