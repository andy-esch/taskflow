package wire

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
	"testing"
)

// TestJSONEnvelopes_RegistryIsComplete is the OTHER direction of the coverage guard in
// envelopes_test.go, and the reason it exists is a bug it could not catch.
//
// That guard walks the jsonEnvelopes registry and demands every REGISTERED envelope have
// a validation case. It is blind to the inverse — an envelope type that is declared,
// constructed, and emitted by a real command but never added to the registry. Such a type
// gets no `$defs` entry, so `schema --json-schema` silently fails to describe output the
// tool actually produces, and an agent validating against the published schema cannot.
// That is exactly what happened to the research envelopes (epic 28): both shipped, both
// emitted `schema_version`, neither appeared in the schema, and every test passed.
//
// Reflection cannot close this gap — a type nothing references is invisible at runtime —
// so the package's own source is parsed instead. Any `type <Name>Envelope struct` must be
// a field of jsonEnvelopes.
func TestJSONEnvelopes_RegistryIsComplete(t *testing.T) {
	registered := map[string]bool{}
	rt := reflect.TypeOf(Envelopes())
	for i := range rt.NumField() {
		registered[rt.Field(i).Type.Name()] = true
	}

	// Parsed file-by-file rather than with parser.ParseDir (deprecated, and it drops files
	// whose build tags exclude them). Every declared envelope is part of the contract
	// regardless of build tags, so no-filter is the behaviour we want anyway.
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	fset := token.NewFileSet()
	declared := map[string]token.Pos{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue // a test fixture envelope is not part of the contract
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			if _, isStruct := ts.Type.(*ast.StructType); !isStruct {
				return true
			}
			if strings.HasSuffix(ts.Name.Name, "Envelope") {
				declared[ts.Name.Name] = ts.Pos()
			}
			return true
		})
	}
	if len(declared) == 0 {
		t.Fatal("parsed no *Envelope declarations — the guard would vacuously pass")
	}
	for name, pos := range declared {
		if !registered[name] {
			t.Errorf("%s: %s is declared but NOT in the jsonEnvelopes registry, so it gets no $defs entry "+
				"and `schema --json-schema` cannot describe it — add it to jsonEnvelopes",
				fset.Position(pos), name)
		}
	}
	t.Logf("%d declared envelope types, all registered", len(declared))
}
