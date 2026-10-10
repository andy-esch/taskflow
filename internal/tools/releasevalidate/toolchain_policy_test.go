package releasevalidate

import (
	"encoding/json"
	"fmt"
	"go/version"
	"os"
	"os/exec"
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
		{"automatic workflow toolchain", ".github/workflows/ci.yml", "GOTOOLCHAIN: local", "GOTOOLCHAIN: auto", "GOTOOLCHAIN: local"},
		{"automatic container toolchain", "build/release-validation/Containerfile", "GOTOOLCHAIN=local", "GOTOOLCHAIN=auto", "GOTOOLCHAIN=local"},
		{"missing container toolchain policy", "build/release-validation/Containerfile", "ENV GOTOOLCHAIN=local", "# no toolchain policy", "GOTOOLCHAIN=local"},
		{"combined container override", "build/release-validation/Containerfile", "ENV GOTOOLCHAIN=local", "ENV GOTOOLCHAIN=local\nENV OTHER=value GOTOOLCHAIN=auto", "GOTOOLCHAIN=local"},
		{"late container toolchain policy", "build/release-validation/Containerfile", "ENV GOTOOLCHAIN=local", "RUN go version\nENV GOTOOLCHAIN=local", "before tool installation"},
		{"conditional minimum setup", ".github/workflows/ci.yml", "test:\n    steps:\n      - uses: actions/setup-go@v7", "test:\n    steps:\n      - uses: actions/setup-go@v7\n        if: false", "unconditional setup-go"},
		{"conditional minimum job", ".github/workflows/ci.yml", "test:\n    steps:", "test:\n    if: false\n    steps:", "unconditional protected job"},
		{"conditional publisher", ".github/workflows/release.yml", "uses: goreleaser/goreleaser-action@v7", "uses: goreleaser/goreleaser-action@v7\n        if: false", "unconditional setup-go before goreleaser"},
		{"publisher toolchain override", ".github/workflows/release.yml", "uses: goreleaser/goreleaser-action@v7", "uses: goreleaser/goreleaser-action@v7\n        env: {GOTOOLCHAIN: auto}", "GOTOOLCHAIN: local"},
		{"prerelease CI linter", ".github/workflows/ci.yml", "version: v2.13", "version: v2.13.1-rc1", "exact stable patch"},
		{"floating CI linter", ".github/workflows/ci.yml", "version: v2.13", "version: latest", "linter line"},
		{"step toolchain override", ".github/workflows/ci.yml", "uses: actions/setup-go@v7", "uses: actions/setup-go@v7\n        env: {GOTOOLCHAIN: auto}", "GOTOOLCHAIN: local"},
		{"job toolchain override", ".github/workflows/ci.yml", "test:\n    steps:", "test:\n    env: {GOTOOLCHAIN: auto}\n    steps:", "GOTOOLCHAIN: local"},
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

func TestToolchainPolicyAcceptsGeneratedLinterPins(t *testing.T) {
	type update struct {
		Name, Workflow, Container string
	}
	updates := []update{
		{"patch", "v2.13.1", "v2.13.1"},
		{"minor", "v2.14.0", "v2.14.0"},
	}
	// The optional actual-engine check supplies real generated replacements to
	// this same production-policy helper without editing repository configuration.
	if raw, ok := os.LookupEnv("TASKFLOW_RENOVATE_LINTER_CASES"); ok {
		if err := json.Unmarshal([]byte(raw), &updates); err != nil || len(updates) == 0 {
			t.Fatalf("invalid generated linter cases: %v", err)
		}
	}
	for _, update := range updates {
		t.Run(update.Name, func(t *testing.T) {
			root := toolchainPolicyFixture(t)
			for file, replacements := range map[string][]string{
				"build/release-validation/Containerfile": {"v2.13.0", update.Container},
				".github/workflows/ci.yml":               {"version: v2.13}", "version: " + update.Workflow + "}"},
			} {
				path := filepath.Join(root, file)
				contents, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				writeFixtureFile(t, path, strings.Replace(string(contents), replacements[0], replacements[1], 1))
			}
			if _, _, err := checkToolchainPolicy(root); err != nil {
				t.Fatalf("generated coordinated pins rejected: %v", err)
			}
		})
	}
}

func TestToolchainPolicyBindsLintToEarlierCompilerSetup(t *testing.T) {
	const linter = "      - uses: golangci/golangci-lint-action@v9\n        with: {version: v2.13}\n"
	const setup = "      - uses: actions/setup-go@v7\n        with: {go-version: \"1.26\", check-latest: true}\n"
	for _, test := range []struct {
		name, lintSteps, want string
	}{
		{"setup-free linter job", linter, "setup-go before golangci-lint"},
		{"linter before compiler setup", linter + setup, "setup-go before golangci-lint"},
		{"linter moved to different job", setup + linter, "lint needs one CI linter selection"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := toolchainPolicyFixture(t)
			// Keep the original lint job's setup; move the linter to a second job.
			contents, err := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
			if err != nil {
				t.Fatal(err)
			}
			text := strings.Replace(string(contents), linter, "", 1)
			writeFixtureFile(t, filepath.Join(root, ".github/workflows/ci.yml"), text+"  extra-lint:\n    steps:\n"+test.lintSteps)
			_, _, err = checkToolchainPolicy(root)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("check error=%v; want %q", err, test.want)
			}
		})
	}
}

func TestLocalGoSelectionIgnoresModuleToolchainSuggestion(t *testing.T) {
	root := t.TempDir()
	selected := func() string {
		t.Helper()
		command := exec.Command("go", "env", "GOVERSION")
		command.Dir = root
		command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off", "GOPROXY=off", "GO111MODULE=on")
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("real Go selection failed: %v\n%s", err, output)
		}
		return strings.TrimSpace(string(output))
	}
	compiler := selected()
	// A future suggestion would need an automatic switch/download without local.
	writeFixtureFile(t, filepath.Join(root, "go.mod"), "module example.test/toolchain\n\ngo 1.26.0\ntoolchain go1.99.0\n")
	if got := selected(); got != compiler {
		t.Fatalf("module replaced selected compiler: got %q; want %q", got, compiler)
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
	path := filepath.Join(root, ".github/workflows/ci.yml")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, path, string(contents)+`  minimum-race:
    steps:
      - uses: actions/setup-go@v7
        with: {go-version: "1.26", check-latest: true}
`)
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
	Env  map[string]string
	Jobs map[string]struct {
		If    yaml.Node
		Env   map[string]string
		Steps []struct {
			Uses string
			Run  string
			If   yaml.Node
			Env  map[string]string
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
	goDirectivePattern    = regexp.MustCompile(`(?m)^go[ \t]+(\S+)[ \t]*(?://[^\n]*)?$`)
	containerGoPattern    = regexp.MustCompile(`(?m)^FROM\s+(?:--platform=\S+\s+)?golang:([^\s]+)`)
	linterARGPattern      = regexp.MustCompile(`(?m)^ARG\s+GOLANGCI_LINT_VERSION=(\S+)\s*$`)
	containerLocalPattern = regexp.MustCompile(`(?m)^ENV[ \t]+GOTOOLCHAIN=(\S+)[ \t]*$`)
	stablePatchPattern    = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
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
	localMatches := containerLocalPattern.FindAllStringSubmatchIndex(string(containerfile), -1)
	toolchainLines := 0
	for line := range strings.SplitSeq(string(containerfile), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") && strings.Contains(line, "GOTOOLCHAIN") {
			toolchainLines++
		}
	}
	if len(localMatches) != 1 || toolchainLines != 1 || string(containerfile[localMatches[0][2]:localMatches[0][3]]) != "local" {
		return "", "", fmt.Errorf("Containerfile needs one ENV GOTOOLCHAIN=local before tool installation")
	}
	if run := strings.Index(string(containerfile), "\nRUN "); run >= 0 && localMatches[0][0] > run {
		return "", "", fmt.Errorf("Containerfile needs ENV GOTOOLCHAIN=local before tool installation")
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
	lintJobCount := 0
	jobs := make([]string, 0, len(workflow.Jobs))
	for name := range workflow.Jobs {
		jobs = append(jobs, name)
	}
	sort.Strings(jobs)
	for _, name := range jobs {
		job := workflow.Jobs[name]
		protected := (file == "ci.yml" && (name == "lint" || name == "test")) || (file == "release.yml" && name == "goreleaser")
		if protected && job.If.Kind != 0 {
			return fmt.Errorf("%s/%s: need an unconditional protected job", file, name)
		}
		line := releaseLine
		if file == "ci.yml" && name == "test" {
			line = minimumLine
		}
		toolchain := workflow.Env["GOTOOLCHAIN"]
		if value, ok := job.Env["GOTOOLCHAIN"]; ok {
			toolchain = value
		}
		if protected && toolchain != "local" {
			return fmt.Errorf("%s/%s: need GOTOOLCHAIN: local to keep the selected compiler", file, name)
		}
		for _, step := range job.Steps {
			effective := toolchain
			if value, ok := step.Env["GOTOOLCHAIN"]; ok {
				effective = value
			}
			if (protected || strings.HasPrefix(step.Uses, "actions/setup-go@")) && effective != "local" {
				return fmt.Errorf("%s/%s: need GOTOOLCHAIN: local without step overrides", file, name)
			}
			if protected && step.Run != "" && goCounts[name] == 0 {
				return fmt.Errorf("%s/%s: need setup-go before protected run steps", file, name)
			}
			if strings.HasPrefix(step.Uses, "actions/setup-go@") {
				if step.If.Kind != 0 {
					return fmt.Errorf("%s/%s: need unconditional setup-go", file, name)
				}
				goCounts[name]++
				additionalMinimum := file == "ci.yml" && !protected && step.With.GoVersion == minimumLine
				if step.With.GoVersion != line && !additionalMinimum {
					return fmt.Errorf("%s/%s: need latest-patch selector for expected Go line %s, got %q", file, name, line, step.With.GoVersion)
				}
				if !step.With.CheckLatest || step.With.GoVersionFile != "" {
					return fmt.Errorf("%s/%s: need check-latest: true and no competing go-version-file", file, name)
				}
			}
			if strings.HasPrefix(step.Uses, "golangci/golangci-lint-action@") {
				linterCount++
				if name == "lint" {
					lintJobCount++
				}
				if goCounts[name] != 1 || effective != "local" || job.If.Kind != 0 || step.If.Kind != 0 {
					return fmt.Errorf("%s/%s: need unconditional setup-go before golangci-lint in the same job", file, name)
				}
				selected := strings.TrimPrefix(step.With.Version, "v")
				if goLine(selected) != linterLine || (selected != linterLine && !stablePatchPattern.MatchString(selected)) {
					return fmt.Errorf("%s/%s: select linter line v%s or an exact stable patch on it to match the container, got %q", file, name, linterLine, step.With.Version)
				}
			}
			if strings.HasPrefix(step.Uses, "goreleaser/goreleaser-action@") && (goCounts[name] != 1 || effective != "local" || job.If.Kind != 0 || step.If.Kind != 0) {
				return fmt.Errorf("%s/%s: need unconditional setup-go before goreleaser in the same job", file, name)
			}
		}
	}
	if file == "ci.yml" {
		for _, job := range []string{"lint", "test"} {
			if goCounts[job] != 1 {
				return fmt.Errorf("%s/%s needs exactly one setup-go", file, job)
			}
		}
		if linterCount != 1 || lintJobCount != 1 {
			return fmt.Errorf("ci.yml/lint needs one CI linter selection")
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
	writeFixtureFile(t, filepath.Join(root, "build/release-validation/Containerfile"), "FROM golang:1.26.9-bookworm\nENV GOTOOLCHAIN=local\nARG GOLANGCI_LINT_VERSION=v2.13.0\n")
	writeFixtureFile(t, filepath.Join(root, ".github/workflows/ci.yml"), `env: {GOTOOLCHAIN: local}
jobs:
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
	writeFixtureFile(t, filepath.Join(root, ".github/workflows/release.yml"), `env: {GOTOOLCHAIN: local}
jobs:
  goreleaser:
    steps:
      - uses: actions/setup-go@v7
        with: {go-version: "1.26", check-latest: true}
      - uses: goreleaser/goreleaser-action@v7
`)
	return root
}
