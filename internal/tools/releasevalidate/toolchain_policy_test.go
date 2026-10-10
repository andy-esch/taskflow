package releasevalidate

import (
	"fmt"
	"go/version"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// This checks configuration agreement, not patch freshness or compiler support.
// govulncheck and the real linter's build-version preflight remain release gates.
// Policy: docs/RELEASING.md#go-and-linter-version-policy.
func TestRepositoryToolchainPolicy(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	minimum, container, err := checkToolchainPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("minimum/test Go %s; release/lint line %s; container Go %s", minimum, goLine(container), container)
}

func TestToolchainPolicyRejectsDrift(t *testing.T) {
	tests := []struct {
		name, file, old, replacement, want string
	}{
		{"module line", "go.mod", "go 1.26.0", "go 1.27.0", "release Go line is older than module minimum"},
		{"raised minimum patch", "go.mod", "go 1.26.0", "go 1.26.10", "older than module minimum"},
		{"missing minimum", "go.mod", "go 1.26.0", "// no minimum", "one valid go directive"},
		{"container line", "build/release-validation/Containerfile", "1.26.9-bookworm", "1.27.2-bookworm", "release Go line"},
		{"floating container", "build/release-validation/Containerfile", "1.26.9-bookworm", "1.26-bookworm", "exact stable Go patch"},
		{"missing container pin", "build/release-validation/Containerfile", "FROM golang:1.26.9-bookworm", "FROM debian:bookworm", "one pinned golang base"},
		{"workflow line", ".github/workflows/ci.yml", `go-version: "1.26"`, `go-version: "1.27"`, "expected Go line"},
		{"workflow exact patch", ".github/workflows/release.yml", `go-version: "1.26"`, `go-version: "1.26.9"`, "latest-patch line selector"},
		{"workflow latest false", ".github/workflows/ci.yml", "check-latest: true", "check-latest: false", "check-latest: true"},
		{"workflow competing selector", ".github/workflows/ci.yml", `go-version: "1.26"`, `go-version: "1.26", go-version-file: go.mod`, "no competing go-version-file"},
		{"missing test setup", ".github/workflows/ci.yml", "test:\n    steps:\n      - uses: actions/setup-go@v7", "test:\n    steps:\n      - uses: actions/checkout@v7", "test needs exactly one setup-go"},
		{"linter line", "build/release-validation/Containerfile", "v2.13.0", "v2.12.2", "linter line"},
		{"floating linter", "build/release-validation/Containerfile", "v2.13.0", "v2.13", "exact linter patch"},
		{"missing CI linter", ".github/workflows/ci.yml", "golangci/golangci-lint-action@v9", "actions/checkout@v7", "one CI linter selection"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := toolchainPolicyFixture(t)
			path := filepath.Join(root, test.file)
			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(contents), test.old) {
				t.Fatalf("fixture lacks %q", test.old)
			}
			writeFixtureFile(t, path, strings.ReplaceAll(string(contents), test.old, test.replacement))
			_, _, err = checkToolchainPolicy(root)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("check error = %v; want %q", err, test.want)
			}
		})
	}
}

func TestToolchainPolicyAllowsNewerReleaseLineWithoutRaisingMinimum(t *testing.T) {
	root := toolchainPolicyFixture(t)
	for file, replacements := range map[string][]string{
		"build/release-validation/Containerfile": {"1.26.9-bookworm", "1.27.2-bookworm"},
		".github/workflows/release.yml":          {`go-version: "1.26"`, `go-version: "1.27"`},
		".github/workflows/ci.yml":               {`go-version: "1.26"`, `go-version: "1.27"`},
	} {
		path := filepath.Join(root, file)
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		// Only the lint job moves forward; the test job still covers the minimum line.
		writeFixtureFile(t, path, strings.Replace(string(contents), replacements[0], replacements[1], 1))
	}
	minimum, container, err := checkToolchainPolicy(root)
	if err != nil || minimum != "1.26.0" || container != "1.27.2" {
		t.Fatalf("minimum=%q container=%q err=%v", minimum, container, err)
	}
}

func TestToolchainPolicyAllowsIndependentPatchRefresh(t *testing.T) {
	root := toolchainPolicyFixture(t)
	path := filepath.Join(root, "build", "release-validation", "Containerfile")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	contents = []byte(strings.ReplaceAll(string(contents), "1.26.9-bookworm", "1.26.10-bookworm@sha256:"+strings.Repeat("a", 64)))
	contents = []byte(strings.ReplaceAll(string(contents), "v2.13.0", "v2.13.1"))
	writeFixtureFile(t, path, string(contents))
	minimum, container, err := checkToolchainPolicy(root)
	if err != nil || minimum != "1.26.0" || container != "1.26.10" {
		t.Fatalf("minimum=%q container=%q err=%v", minimum, container, err)
	}
}

type toolchainWorkflow struct {
	Jobs map[string]struct {
		Steps []struct {
			Uses string
			With struct {
				GoVersion     string `yaml:"go-version"`
				GoVersionFile string `yaml:"go-version-file"`
				CheckLatest   bool   `yaml:"check-latest"`
				Version       string
			}
		}
	}
}

var (
	goDirectivePattern = regexp.MustCompile(`(?m)^go[ \t]+(\S+)[ \t]*(?://[^\n]*)?$`)
	containerGoPattern = regexp.MustCompile(`(?m)^FROM\s+(?:--platform=\S+\s+)?golang:([^\s]+)`)
	linterARGPattern   = regexp.MustCompile(`(?m)^ARG\s+GOLANGCI_LINT_VERSION=(\S+)\s*$`)
	stablePatchPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
)

func checkToolchainPolicy(root string) (minimum, container string, err error) {
	module, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", "", err
	}
	matches := goDirectivePattern.FindAllStringSubmatch(string(module), -1)
	if len(matches) != 1 || !version.IsValid("go"+matches[0][1]) {
		return "", "", fmt.Errorf("go.mod needs one valid go directive")
	}
	minimum = matches[0][1]
	release, err := readToolchainWorkflow(root, "release.yml")
	if err != nil {
		return "", "", err
	}
	releaseLine := ""
	for _, step := range release.Jobs["goreleaser"].Steps {
		if strings.HasPrefix(step.Uses, "actions/setup-go@") {
			if releaseLine != "" {
				return "", "", fmt.Errorf("release.yml/goreleaser needs exactly one setup-go")
			}
			releaseLine = step.With.GoVersion
		}
	}
	if releaseLine == "" || !version.IsValid("go"+releaseLine) || goLine(releaseLine) != releaseLine {
		return "", "", fmt.Errorf("release.yml/goreleaser needs one latest-patch line selector")
	}
	if version.Compare("go"+releaseLine, "go"+goLine(minimum)) < 0 {
		return "", "", fmt.Errorf("release Go line is older than module minimum %s", minimum)
	}
	containerfile, err := os.ReadFile(filepath.Join(root, "build", "release-validation", "Containerfile"))
	if err != nil {
		return "", "", err
	}
	matches = containerGoPattern.FindAllStringSubmatch(string(containerfile), -1)
	if len(matches) != 1 {
		return "", "", fmt.Errorf("Containerfile needs one pinned golang base")
	}
	tag, _, _ := strings.Cut(matches[0][1], "@")
	container, _, _ = strings.Cut(tag, "-")
	if !stablePatchPattern.MatchString(container) || !version.IsValid("go"+container) {
		return "", "", fmt.Errorf("Containerfile needs an exact stable Go patch, got %q", container)
	}
	if goLine(container) != releaseLine {
		return "", "", fmt.Errorf("container Go %s must match release Go line %s", container, releaseLine)
	}
	if version.Compare("go"+container, "go"+minimum) < 0 {
		return "", "", fmt.Errorf("container Go %s is older than module minimum %s", container, minimum)
	}
	matches = linterARGPattern.FindAllStringSubmatch(string(containerfile), -1)
	if len(matches) != 1 || !stablePatchPattern.MatchString(strings.TrimPrefix(matches[0][1], "v")) {
		return "", "", fmt.Errorf("Containerfile needs one exact linter patch")
	}
	linterLine := goLine(strings.TrimPrefix(matches[0][1], "v"))
	for _, file := range []string{"ci.yml", "release.yml"} {
		if err := checkWorkflowToolchains(root, file, goLine(minimum), releaseLine, linterLine); err != nil {
			return "", "", err
		}
	}
	return minimum, container, nil
}

func readToolchainWorkflow(root, file string) (toolchainWorkflow, error) {
	var workflow toolchainWorkflow
	contents, err := os.ReadFile(filepath.Join(root, ".github", "workflows", file))
	if err != nil {
		return workflow, err
	}
	if err := yaml.Unmarshal(contents, &workflow); err != nil {
		return workflow, fmt.Errorf("%s: %w", file, err)
	}
	return workflow, nil
}

func checkWorkflowToolchains(root, file, minimumLine, releaseLine, linterLine string) error {
	workflow, err := readToolchainWorkflow(root, file)
	if err != nil {
		return err
	}
	goCounts := map[string]int{}
	linterCount := 0
	jobs := make([]string, 0, len(workflow.Jobs))
	for name := range workflow.Jobs {
		jobs = append(jobs, name)
	}
	sort.Strings(jobs)
	for _, name := range jobs {
		line := releaseLine
		if file == "ci.yml" && name == "test" {
			line = minimumLine
		}
		for _, step := range workflow.Jobs[name].Steps {
			if strings.HasPrefix(step.Uses, "actions/setup-go@") {
				goCounts[name]++
				if step.With.GoVersion != line {
					return fmt.Errorf("%s/%s: need latest-patch selector for expected Go line %s, got %q", file, name, line, step.With.GoVersion)
				}
				if !step.With.CheckLatest || step.With.GoVersionFile != "" {
					return fmt.Errorf("%s/%s: need check-latest: true and no competing go-version-file", file, name)
				}
			}
			if strings.HasPrefix(step.Uses, "golangci/golangci-lint-action@") {
				linterCount++
				if strings.TrimPrefix(step.With.Version, "v") != linterLine {
					return fmt.Errorf("%s/%s: select linter line v%s to match the container, got %q", file, name, linterLine, step.With.Version)
				}
			}
		}
	}
	if file == "ci.yml" {
		for _, job := range []string{"lint", "test"} {
			if goCounts[job] != 1 {
				return fmt.Errorf("%s/%s needs exactly one setup-go", file, job)
			}
		}
		if linterCount != 1 {
			return fmt.Errorf("ci.yml needs one CI linter selection")
		}
	} else if goCounts["goreleaser"] != 1 {
		return fmt.Errorf("release.yml/goreleaser needs exactly one setup-go")
	}
	return nil
}

func goLine(value string) string {
	parts := strings.Split(value, ".")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[:2], ".")
}

func toolchainPolicyFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{".github/workflows", "build/release-validation"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeFixtureFile(t, filepath.Join(root, "go.mod"), "module example.test/toolchains\n\ngo 1.26.0\n")
	writeFixtureFile(t, filepath.Join(root, "build/release-validation/Containerfile"), "FROM golang:1.26.9-bookworm\nARG GOLANGCI_LINT_VERSION=v2.13.0\n")
	writeFixtureFile(t, filepath.Join(root, ".github/workflows/ci.yml"), `jobs:
  lint:
    steps:
      - uses: actions/setup-go@v7
        with: {go-version: "1.26", check-latest: true}
      - uses: golangci/golangci-lint-action@v9
        with: {version: v2.13}
  test:
    steps:
      - uses: actions/setup-go@v7
        with: {go-version: "1.26", check-latest: true}
`)
	writeFixtureFile(t, filepath.Join(root, ".github/workflows/release.yml"), `jobs:
  goreleaser:
    steps:
      - uses: actions/setup-go@v7
        with: {go-version: "1.26", check-latest: true}
`)
	return root
}
