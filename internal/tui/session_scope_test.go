package tui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// A pointer type is unnamed, so reflect reports no package path for it. Without the
// dereference, a pointer-shaped runtime message would be swallowed into sessionMsg and
// the runtime would never see it.
func TestSessionScopingKeepsPointerShapedRuntimeMessagesVisible(t *testing.T) {
	msg := scopeSession(1, func() tea.Msg { return &tea.QuitMsg{} })()
	if _, hidden := msg.(sessionMsg); hidden {
		t.Fatalf("pointer-typed runtime message was hidden by session scoping: %T", msg)
	}
	// Application results, pointer or not, must still be stamped.
	app := scopeSession(1, func() tea.Msg { return &fsEventMsg{} })()
	if _, scoped := app.(sessionMsg); !scoped {
		t.Fatalf("pointer-typed application message escaped session scoping: %T", app)
	}
}

// tea.Sequence's payload is an unexported []Cmd, so scopeSession cannot reach inside it
// the way it does for tea.Batch: a sequence's commands would run unstamped and a result
// from the previous space could act on the new one. Until that changes, this package
// must not use it — which is only a safe rule if something checks.
func TestSessionScopingHasNoSequenceCommandsToLeak(t *testing.T) {
	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list package sources: %v", err)
	}
	fset := token.NewFileSet()
	for _, name := range sources {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Sequence" {
				return true
			}
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "tea" {
				t.Errorf("%s uses tea.Sequence, which escapes session scoping (see scopeSession)",
					fset.Position(sel.Pos()))
			}
			return true
		})
	}
}
