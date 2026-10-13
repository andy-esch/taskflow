package architecturecheck

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

const (
	importMapStart = "<!-- taskflow:internal-import-map:v1 -->"
	importMapEnd   = "<!-- /taskflow:internal-import-map:v1 -->"
)

var internalPackageName = regexp.MustCompile(`^[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*$`)

// Each source has a row, including leaves. The map describes observed imports,
// not permission to introduce them; depguard and the policy ADR remain separate.
type importGraph map[string][]string

func TestArchitectureImportMap(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	guide := filepath.Join(root, "docs", "architecture", "boundaries.md")
	documented, err := parseImportMap(readFile(t, guide))
	if err != nil {
		t.Fatalf("%s: %v", guide, err)
	}
	if err := recordImportMapInputs(root); err != nil {
		t.Fatal(err)
	}
	actual, err := productionImportGraph(root)
	if err != nil {
		t.Fatal(err)
	}
	if diff := importGraphDiff(documented, actual); diff != "" {
		t.Fatalf("%s: production internal import map drift (- documented, + go list):\n%s\nReview the code change against .golangci.yml and the guide's named exceptions before updating the marked map. Recheck with go test ./internal/tools/architecturecheck -run '^TestArchitectureImportMap$' -count=1. This inventory is not an import allowlist; new policy requires a separate decision.", guide, diff)
	}
}

// Go's test-result cache only sees filesystem access by this test process, not
// its go list child. Opening every internal directory registers entry metadata
// (including newly added/removed packages); statting Go files also tracks symlink
// targets. Module manifests affect discovery too. Keep ordinary cached test runs
// sensitive to source drift without disabling caching for the whole suite.
func recordImportMapInputs(root string) error {
	if err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			_, err = os.Stat(path)
		}
		return err
	}); err != nil {
		return fmt.Errorf("record import-map source inputs: %w", err)
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil && (name != "go.sum" || !os.IsNotExist(err)) {
			return fmt.Errorf("record import-map module input %s: %w", name, err)
		}
	}
	return nil
}

func parseImportMap(content string) (importGraph, error) {
	if strings.Count(content, importMapStart) != 1 || strings.Count(content, importMapEnd) != 1 {
		return nil, fmt.Errorf("expected exactly one internal-import-map:v1 marker pair")
	}
	_, rest, _ := strings.Cut(content, importMapStart)
	block, _, found := strings.Cut(rest, importMapEnd)
	if !found {
		return nil, fmt.Errorf("internal import map closing marker precedes its opening marker")
	}
	lines := strings.Split(strings.TrimSpace(block), "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "```text" || strings.TrimSpace(lines[len(lines)-1]) != "```" {
		return nil, fmt.Errorf("internal import map must contain one fenced text block")
	}
	graph := make(importGraph)
	var row string
	flush := func() error {
		if row == "" {
			return nil
		}
		if strings.Count(row, "->") != 1 {
			return fmt.Errorf("invalid import row %q: expected source -> targets", row)
		}
		source, targetList, _ := strings.Cut(row, "->")
		source, targetList = strings.TrimSpace(source), strings.TrimSpace(targetList)
		if !validInternalPackageName(source) {
			return fmt.Errorf("invalid import-map source %q", source)
		}
		if _, exists := graph[source]; exists {
			return fmt.Errorf("duplicate import-map source %q", source)
		}
		var targets []string
		if targetList != "-" {
			for _, target := range strings.Split(targetList, ",") {
				target = strings.TrimSpace(target)
				if target == "-" || !validInternalPackageName(target) {
					return fmt.Errorf("invalid import-map target %q for %s; use literal package paths or '-' alone", target, source)
				}
				if slices.Contains(targets, target) {
					return fmt.Errorf("duplicate import-map edge %s -> %s", source, target)
				}
				targets = append(targets, target)
			}
		}
		slices.Sort(targets)
		graph[source] = targets
		return nil
	}
	for _, line := range lines[1 : len(lines)-1] {
		line, _, _ = strings.Cut(line, "#")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "->") {
			if err := flush(); err != nil {
				return nil, err
			}
			row = line
		} else {
			if row == "" {
				return nil, fmt.Errorf("import-map continuation without a source: %q", line)
			}
			row += " " + line
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	if len(graph) == 0 {
		return nil, fmt.Errorf("internal import map has no source rows")
	}
	return graph, nil
}

func validInternalPackageName(name string) bool {
	return fs.ValidPath(name) && internalPackageName.MatchString(name)
}

func productionImportGraph(root string) (importGraph, error) {
	moduleJSON, err := goList(root, "-m", "-mod=readonly", "-json")
	if err != nil {
		return nil, err
	}
	var module struct {
		Path string
		Main bool
	}
	if err := json.Unmarshal(moduleJSON, &module); err != nil {
		return nil, fmt.Errorf("decode current go module: %w", err)
	}
	if !module.Main || module.Path == "" {
		return nil, fmt.Errorf("go list did not identify the current main module")
	}
	packages, err := goList(root, "-mod=readonly", "-json", "./internal/...")
	if err != nil {
		return nil, err
	}
	return decodeProductionImports(module.Path, bytes.NewReader(packages))
}

func goList(root string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", append([]string{"list"}, args...)...)
	cmd.Dir = root
	// Check this module's default build, not an ambient workspace or -test/-find/
	// custom-tag flags. Retain the normal toolchain, cache and GOOS/GOARCH context.
	// An empty GOFLAGS would fall back to a persisted go env -w value. A
	// non-empty, already-required flag overrides it without dropping GOPRIVATE
	// or other useful user settings via GOENV=off.
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=readonly")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list %s: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return output, nil
}

type listedPackage struct {
	ImportPath string
	GoFiles    []string
	CgoFiles   []string
	Imports    []string
	ForTest    string
	Incomplete bool
	Error      *struct{ Err string }
	DepsErrors []struct{ Err string }
}

func decodeProductionImports(module string, input io.Reader) (importGraph, error) {
	graph := make(importGraph)
	decoder := json.NewDecoder(input)
	for {
		var pkg listedPackage
		if err := decoder.Decode(&pkg); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("decode go list package metadata: %w", err)
		}
		if pkg.Incomplete || pkg.Error != nil || len(pkg.DepsErrors) != 0 || pkg.ForTest != "" {
			return nil, fmt.Errorf("go list returned incomplete, erroneous or test-variant metadata for %q", pkg.ImportPath)
		}
		source, ok := internalImportName(module, pkg.ImportPath)
		if !ok || !validInternalPackageName(source) {
			return nil, fmt.Errorf("go list returned unexpected source package %q", pkg.ImportPath)
		}
		if importMapSupportPackage(source) || len(pkg.GoFiles)+len(pkg.CgoFiles) == 0 {
			continue
		}
		if _, exists := graph[source]; exists {
			return nil, fmt.Errorf("go list returned duplicate source package %q", pkg.ImportPath)
		}
		var targets []string
		for _, path := range pkg.Imports {
			if target, ok := internalImportName(module, path); ok {
				if !validInternalPackageName(target) {
					return nil, fmt.Errorf("go list returned invalid internal import %q", path)
				}
				targets = append(targets, target)
			}
		}
		slices.Sort(targets)
		graph[source] = slices.Compact(targets)
	}
	if len(graph) == 0 {
		return nil, fmt.Errorf("go list returned no covered production internal packages")
	}
	// ./internal/... skips _/. directories, testdata and nested modules even
	// when an ordinary production package imports them. Do not bless a partial
	// graph: excluded support targets are the only intentional missing rows.
	var uncovered []string
	for source, targets := range graph {
		for _, target := range targets {
			if _, covered := graph[target]; !covered && !importMapSupportPackage(target) {
				uncovered = append(uncovered, fmt.Sprintf("uncovered internal target %s (imported by %s)", target, source))
			}
		}
	}
	if len(uncovered) != 0 {
		slices.Sort(uncovered)
		return nil, fmt.Errorf("%s; go list ./internal/... skips _/. directories, testdata and nested modules; move the package into the covered tree or explicitly revise discovery before updating the map", strings.Join(uncovered, "; "))
	}
	return graph, nil
}

func importMapSupportPackage(name string) bool {
	return name == "testutil" || name == "tools" || strings.HasPrefix(name, "tools/")
}

func internalImportName(module, path string) (string, bool) {
	root := module + "/internal"
	if path == root {
		return ".", true
	}
	return strings.CutPrefix(path, root+"/")
}

func importGraphDiff(documented, actual importGraph) string {
	var packages []string
	for name := range documented {
		packages = append(packages, name)
	}
	for name := range actual {
		packages = append(packages, name)
	}
	slices.Sort(packages)
	var lines []string
	for _, source := range slices.Compact(packages) {
		before, inDoc := documented[source]
		after, inCode := actual[source]
		if !inDoc {
			lines = append(lines, "+ package "+source)
		} else if !inCode {
			lines = append(lines, "- package "+source)
		}
		var targets []string
		targets = append(targets, before...)
		targets = append(targets, after...)
		slices.Sort(targets)
		for _, target := range slices.Compact(targets) {
			switch {
			case !slices.Contains(before, target):
				lines = append(lines, "+ edge "+source+" -> "+target)
			case !slices.Contains(after, target):
				lines = append(lines, "- edge "+source+" -> "+target)
			}
		}
	}
	return strings.Join(lines, "\n")
}
