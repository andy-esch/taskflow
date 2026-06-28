// Command schemacomments regenerates the embedded Go-doc comment map that
// populates field descriptions in `tskflwctl schema --json-schema`.
//
// invopop's AddGoComments reads doc comments from the *source tree*, which a
// shipped binary doesn't have — so it runs here at build time and the resulting
// map is committed + //go:embed-ed by wire.JSONSchema. A drift test
// (wire.TestSchemaComments_NotStale) fails if this output goes stale.
//
// Run from the repo root: `go run ./internal/tools/schemacomments`.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
)

const (
	module  = "github.com/andy-esch/taskflow"
	outPath = "internal/wire/schema_comments.json"
)

func main() {
	r := new(jsonschema.Reflector)
	// Only the packages the schema references — the envelope/DTO types (wire) and the
	// domain types they embed (FileProblem, Issue, FieldDoc). Scoping it here means
	// an unrelated comment edit elsewhere can't make this file "stale".
	for _, dir := range []string{"./internal/wire", "./internal/domain"} {
		if err := r.AddGoComments(module, dir); err != nil {
			fmt.Fprintln(os.Stderr, "AddGoComments:", err)
			os.Exit(1)
		}
	}
	// json.Marshal sorts map keys, so the committed file is stable across runs.
	b, err := json.MarshalIndent(r.CommentMap, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "marshal:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(outPath, append(b, '\n'), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %d comments to %s\n", len(r.CommentMap), outPath)
}
