package architecturecheck

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func markedImportMap(rows string) string {
	return importMapStart + "\n```text\n" + rows + "\n```\n" + importMapEnd
}

func TestParseImportMapIgnoresPresentationChanges(t *testing.T) {
	want := importGraph{"core": {"domain", "id"}, "cli/ports": {"core"}, "leaf": nil}
	for name, content := range map[string]string{
		"plain":                      markedImportMap("core -> domain, id\ncli/ports -> core\nleaf -> -"),
		"wrapped reordered comments": "# A different guide heading\n" + markedImportMap("# explanatory comment\n leaf\t->\t-\ncli/ports->core\ncore -> id, # continue after comma\n  domain\n") + "\nExtra prose and other fenced examples.",
		"CRLF":                       strings.ReplaceAll(markedImportMap("core -> domain, id\ncli/ports -> core\nleaf -> -"), "\n", "\r\n"),
	} {
		t.Run(name, func(t *testing.T) {
			got, err := parseImportMap(content)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("graph = %#v, error = %v; want %#v", got, err, want)
			}
		})
	}
}

func TestParseImportMapRejectsAmbiguousOrMissingFacts(t *testing.T) {
	for name, content := range map[string]string{
		"no markers":               "```text\ncore -> domain\n```",
		"missing close":            importMapStart + "\n```text\ncore -> domain\n```",
		"reversed markers":         importMapEnd + "\n" + importMapStart,
		"duplicate block":          markedImportMap("core -> domain") + markedImportMap("core -> domain"),
		"missing fence":            importMapStart + "\ncore -> domain\n" + importMapEnd,
		"empty":                    markedImportMap("# only commentary"),
		"missing targets":          markedImportMap("core ->"),
		"trailing comma":           markedImportMap("core -> domain,"),
		"duplicate source":         markedImportMap("core -> domain\ncore -> id"),
		"duplicate target":         markedImportMap("core -> domain, domain"),
		"shorthand":                markedImportMap("core -> domain + utilities"),
		"mixed leaf declaration":   markedImportMap("core -> -, domain"),
		"leaf token as source":     markedImportMap("- -> core"),
		"leaf token on both sides": markedImportMap("- -> -"),
		"stray continuation":       markedImportMap("domain\ncore -> id"),
		"extra arrow":              markedImportMap("core -> domain -> id"),
		"invalid source":           markedImportMap("../core -> domain"),
		"invalid target":           markedImportMap("core -> /domain"),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseImportMap(content); err == nil {
				t.Fatal("malformed import map passed")
			}
		})
	}
}

func TestProductionImportProjectionKeepsScopeAndLeaves(t *testing.T) {
	const module = "example.test/planning"
	data := `
{"ImportPath":"example.test/planning/internal/core","GoFiles":["service.go"],"Imports":["fmt","example.test/planning/internal/id","example.test/planning/internal/domain","example.test/planning/internal/domain","external.test/internal/other","example.test/planning/internality/other"],"TestImports":["example.test/planning/internal/store"],"XTestImports":["example.test/planning/internal/testutil"],"Deps":["example.test/planning/internal/config"]}
{"ImportPath":"example.test/planning/internal/domain","GoFiles":["domain.go"]}
{"ImportPath":"example.test/planning/internal/id","GoFiles":["id.go"]}
{"ImportPath":"example.test/planning/internal/leaf","GoFiles":["leaf.go"]}
{"ImportPath":"example.test/planning/internal/cgo","CgoFiles":["native.go"],"Imports":["C","example.test/planning/internal/leaf"]}
{"ImportPath":"example.test/planning/internal/runtime","GoFiles":["runtime.go"],"Imports":["example.test/planning/internal/testutil","example.test/planning/internal/tools/helper"]}
{"ImportPath":"example.test/planning/internal/testutil","GoFiles":["fixture.go"],"Imports":["example.test/planning/internal/store"]}
{"ImportPath":"example.test/planning/internal/tools/helper","GoFiles":["main.go"],"Imports":["example.test/planning/internal/store"]}
{"ImportPath":"example.test/planning/internal/onlytests","TestGoFiles":["contract_test.go"],"TestImports":["example.test/planning/internal/core"]}
{"ImportPath":"example.test/planning/internal/testutils","GoFiles":["runtime.go"]}
{"ImportPath":"example.test/planning/internal/toolsmith","GoFiles":["runtime.go"]}
{"ImportPath":"example.test/planning/internal","GoFiles":["shared.go"],"Imports":["example.test/planning/internal/leaf"]}
`
	want := importGraph{
		"core": {"domain", "id"}, "domain": nil, "id": nil, "leaf": nil, "cgo": {"leaf"},
		"runtime": {"testutil", "tools/helper"}, "testutils": nil, "toolsmith": nil, ".": {"leaf"},
	}
	got, err := decodeProductionImports(module, strings.NewReader(data))
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("graph = %#v, error = %v; want %#v", got, err, want)
	}
}

func TestProductionImportProjectionFailsClosed(t *testing.T) {
	const prefix = "example.test/planning/internal/"
	good := listedPackage{ImportPath: prefix + "leaf", GoFiles: []string{"leaf.go"}}
	encoded, err := json.Marshal(good)
	if err != nil {
		t.Fatal(err)
	}
	for name, pkg := range map[string]listedPackage{
		"incomplete":         {ImportPath: prefix + "core", GoFiles: good.GoFiles, Incomplete: true},
		"load error":         {ImportPath: prefix + "core", Error: &struct{ Err string }{Err: "cannot load"}},
		"dep error":          {ImportPath: prefix + "core", GoFiles: good.GoFiles, DepsErrors: []struct{ Err string }{{Err: "missing dependency"}}},
		"test variant":       {ImportPath: prefix + "core", GoFiles: good.GoFiles, ForTest: prefix + "core"},
		"foreign source":     {ImportPath: "example.test/another/internal/core", GoFiles: good.GoFiles},
		"invalid import":     {ImportPath: prefix + "core", GoFiles: good.GoFiles, Imports: []string{prefix + "../bad"}},
		"broken support":     {ImportPath: prefix + "tools/helper", GoFiles: good.GoFiles, Incomplete: true},
		"broken test-only":   {ImportPath: prefix + "onlytests", Error: &struct{ Err string }{Err: "cannot load"}},
		"missing target row": {ImportPath: prefix + "core", GoFiles: good.GoFiles, Imports: []string{prefix + "_hidden"}},
	} {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(pkg)
			if err != nil {
				t.Fatal(err)
			}
			// A valid row first prevents the empty-graph refusal from masking
			// validate-after-filter mutants for excluded support/test records.
			if _, err := decodeProductionImports("example.test/planning", strings.NewReader(string(encoded)+string(data))); err == nil {
				t.Fatal("invalid go metadata passed")
			}
		})
	}
	for name, data := range map[string]string{
		"empty":             "",
		"truncated":         string(encoded) + `{"ImportPath":`,
		"trailing garbage":  string(encoded) + "not json",
		"duplicate package": string(encoded) + string(encoded),
		"test-only corpus":  `{"ImportPath":"example.test/planning/internal/onlytests","TestGoFiles":["test.go"]}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeProductionImports("example.test/planning", strings.NewReader(data)); err == nil {
				t.Fatal("invalid/incomplete metadata stream passed")
			}
		})
	}
}

func TestProductionImportGraphUsesDefaultModuleBuildWithoutEdits(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod":                        "module example.test/mapfixture\n\ngo 1.20\n",
		"internal/a/a.go":               "package a\nimport (_ \"example.test/mapfixture/internal/b\"; _ \"example.test/mapfixture/internal/testutil\"; _ \"example.test/mapfixture/internal/tools\"; _ \"example.test/mapfixture/internal/tools/sub\")\n",
		"internal/a/a_test.go":          "package a\nimport _ \"example.test/mapfixture/internal/testutil\"\n",
		"internal/b/b.go":               "package b\n",
		"internal/testutil/f.go":        "package testutil\nimport _ \"example.test/mapfixture/internal/b\"\n",
		"internal/tools/helper.go":      "package tools\nimport _ \"example.test/mapfixture/internal/b\"\n",
		"internal/tools/sub/helper.go":  "package sub\nimport _ \"example.test/mapfixture/internal/b\"\n",
		"internal/shared.go":            "package shared\nimport _ \"example.test/mapfixture/internal/b\"\n",
		"internal/testutils/helpers.go": "package testutils\n",
		"internal/toolsmith/tools.go":   "package toolsmith\nimport _ \"example.test/mapfixture/internal/b\"\n",
		"internal/onlytests/t_test.go":  "package onlytests\nimport _ \"example.test/mapfixture/internal/a\"\n",
		"internal/tagged/default.go":    "//go:build !reviewtag\n\npackage tagged\nimport _ \"example.test/mapfixture/internal/b\"\n",
		"internal/tagged/custom.go":     "//go:build reviewtag\n\npackage tagged\nimport _ \"example.test/mapfixture/internal/testutil\"\n",
	}
	writeImportFixture(t, root, files)
	goenv := filepath.Join(root, "goenv")
	writeImportFixture(t, root, map[string]string{"goenv": "GOFLAGS=-tags=reviewtag\n"})
	// Neither an unrelated/broken workspace nor -find may erase production facts.
	t.Setenv("GOWORK", filepath.Join(root, "missing.go.work"))
	t.Setenv("GOENV", goenv)
	// Compile the root, support targets and near-collision names, not only
	// their synthetic metadata representations.
	if output, err := fixtureGo(root, "build", "./internal/..."); err != nil {
		t.Fatalf("real-module fixture is not compiler-valid: %v\n%s", err, output)
	}
	for _, flags := range []string{"-find", ""} {
		t.Run("GOFLAGS="+flags, func(t *testing.T) {
			t.Setenv("GOFLAGS", flags)
			got, err := productionImportGraph(root)
			want := importGraph{".": {"b"}, "a": {"b", "testutil", "tools", "tools/sub"}, "b": nil, "tagged": {"b"}, "testutils": nil, "toolsmith": {"b"}}
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("graph = %#v, error = %v; want %#v", got, err, want)
			}
		})
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		content, err := os.ReadFile(filepath.Join(root, name))
		if name == "go.sum" {
			if !os.IsNotExist(err) {
				t.Fatalf("unexpected go.sum creation: %q, %v", content, err)
			}
		} else if err != nil || string(content) != files[name] {
			t.Fatalf("module file changed: %q, %v", content, err)
		}
	}
}

func TestProductionImportGraphRefusesUncoveredTargets(t *testing.T) {
	for _, target := range []string{"_hidden", "p/testdata/td", "nested"} {
		t.Run(target, func(t *testing.T) {
			root := t.TempDir()
			files := map[string]string{
				"go.mod":                            "module example.test/mapfixture\n\ngo 1.20\n",
				"internal/p/p.go":                   "package p\nimport _ \"example.test/mapfixture/internal/" + target + "\"\n",
				"internal/" + target + "/target.go": "package target\n",
			}
			if target == "nested" {
				files["go.mod"] += "\nrequire example.test/mapfixture/internal/nested v0.0.0\nreplace example.test/mapfixture/internal/nested => ./internal/nested\n"
				files["internal/nested/go.mod"] = "module example.test/mapfixture/internal/nested\n\ngo 1.20\n"
			}
			writeImportFixture(t, root, files)
			if output, err := fixtureGo(root, "build", "./internal/p"); err != nil {
				t.Fatalf("fixture is not compiler-valid: %v\n%s", err, output)
			}
			if _, err := productionImportGraph(root); err == nil || !strings.Contains(err.Error(), "uncovered internal target "+target) {
				t.Fatalf("uncovered production target passed or wrong diagnostic: %v", err)
			}
		})
	}
}

func TestImportGraphDiffNamesPackageAndEdgeDriftDeterministically(t *testing.T) {
	before := importGraph{"core": {"domain", "id"}, "oldLeaf": nil}
	after := importGraph{"core": {"domain", "store"}, "newLeaf": nil}
	want := "- edge core -> id\n+ edge core -> store\n+ package newLeaf\n- package oldLeaf"
	for range 20 {
		if got := importGraphDiff(before, after); got != want {
			t.Fatalf("diff = %q, want %q", got, want)
		}
	}
	if diff := importGraphDiff(before, before); diff != "" {
		t.Fatalf("equal graphs drifted: %s", diff)
	}
	if diff := importGraphDiff(importGraph{"core": {"domain"}}, importGraph{"core": {"domain", "id"}}); diff != "+ edge core -> id" {
		t.Fatalf("new edge was hidden: %q", diff)
	}
	if diff := importGraphDiff(importGraph{"leaf": nil}, importGraph{}); diff != "- package leaf" {
		t.Fatalf("removed leaf was hidden: %q", diff)
	}
}
