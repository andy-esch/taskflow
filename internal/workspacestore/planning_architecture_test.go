package workspacestore

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

// Reader counters cannot observe a direct config.Discover call. Keep the shared
// constructor's config dependency data-only; actual discovery belongs to its
// supplied reader and the opening adapters, never to a second construction scan.
func TestSharedPlanningConstructorUsesConfigOnlyAsData(t *testing.T) {
	positions := token.NewFileSet()
	file, err := parser.ParseFile(positions, "planning.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Fatal(err)
		}
		if path != "github.com/andy-esch/taskflow/internal/config" {
			continue
		}
		name := "config"
		if spec.Name != nil {
			name = spec.Name.Name
		}
		if name == "." {
			t.Fatal("shared construction must not hide config API access behind a dot import")
		}
		ast.Inspect(file, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := selector.X.(*ast.Ident)
			if ok && pkg.Name == name && selector.Sel.Name != "Config" {
				t.Errorf("%s: shared construction may use config.Config only; use the supplied discovery reader, not %s.%s",
					positions.Position(selector.Pos()), name, selector.Sel.Name)
			}
			return true
		})
	}
}
