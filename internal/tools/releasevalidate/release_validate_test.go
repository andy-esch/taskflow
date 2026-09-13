package releasevalidate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseValidateSuccessAndFailureBoundaries(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fixture := newReleaseValidationFixture(t)
		output, err := fixture.run()
		if err != nil {
			t.Fatalf("release validation failed: %v\n%s", err, output)
		}
		for _, want := range []string{
			"==> focused package tests",
			"==> full race suite",
			"==> generated CLI documentation",
			"==> isolated release snapshot",
			"release validation passed for",
		} {
			if !strings.Contains(output, want) {
				t.Errorf("output omitted %q:\n%s", want, output)
			}
		}
		fixture.requireClean(t)
	})

	t.Run("dirty candidate", func(t *testing.T) {
		fixture := newReleaseValidationFixture(t)
		if err := os.WriteFile(filepath.Join(fixture.repo, "dirty.txt"), []byte("dirty\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		output, err := fixture.run()
		if err == nil || !strings.Contains(output, "candidate is not clean") || !strings.Contains(output, "dirty.txt") {
			t.Fatalf("dirty candidate result=%v output:\n%s", err, output)
		}
	})

	t.Run("missing prerequisite", func(t *testing.T) {
		fixture := newReleaseValidationFixture(t)
		if err := os.Remove(filepath.Join(fixture.bin, "goreleaser")); err != nil {
			t.Fatal(err)
		}
		output, err := fixture.run()
		if err == nil || !strings.Contains(output, "missing required command 'goreleaser'") {
			t.Fatalf("missing tool result=%v output:\n%s", err, output)
		}
		fixture.requireClean(t)
	})

	t.Run("unsupported toolchain", func(t *testing.T) {
		fixture := newReleaseValidationFixture(t)
		output, err := fixture.run("FAKE_GO_VERSION=1.24.9")
		if err == nil || !strings.Contains(output, "Go 1.25.12 or newer is required") {
			t.Fatalf("old Go result=%v output:\n%s", err, output)
		}
		fixture.requireClean(t)
	})

	t.Run("race failure stops later phases", func(t *testing.T) {
		fixture := newReleaseValidationFixture(t)
		output, err := fixture.run("FAKE_GO_FAIL_MATCH=test -race")
		if err == nil || !strings.Contains(output, "phase failed: full race suite") {
			t.Fatalf("race failure result=%v output:\n%s", err, output)
		}
		log := fixture.readLog(t)
		if strings.Contains(log, "goreleaser release") {
			t.Fatalf("snapshot ran after race failure:\n%s", log)
		}
		fixture.requireClean(t)
	})

	t.Run("stale generated docs fail without rewriting source", func(t *testing.T) {
		fixture := newReleaseValidationFixture(t)
		output, err := fixture.run("FAKE_STALE_DOCS=1")
		if err == nil || !strings.Contains(output, "phase failed: generated CLI documentation") {
			t.Fatalf("stale docs result=%v output:\n%s", err, output)
		}
		fixture.requireClean(t)
	})
}

type releaseValidationFixture struct {
	repo   string
	bin    string
	log    string
	script string
}

func newReleaseValidationFixture(t *testing.T) releaseValidationFixture {
	t.Helper()
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(filepath.Join(repo, "docs", "cli"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "internal", "wire"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFixtureFile(t, filepath.Join(repo, "go.mod"), "module example.test/release\n\ngo 1.25.12\n")
	writeFixtureFile(t, filepath.Join(repo, "docs", "cli", "fixture.md"), "generated docs\n")
	writeFixtureFile(t, filepath.Join(repo, "internal", "wire", "schema_comments.json"), "{}\n")

	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	scriptBytes, err := os.ReadFile(filepath.Join(repositoryRoot, "scripts", "release-validate.sh"))
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(repo, "release-validate.sh")
	if err := os.WriteFile(script, scriptBytes, 0o755); err != nil {
		t.Fatal(err)
	}

	log := filepath.Join(root, "commands.log")
	writeFixtureFile(t, filepath.Join(bin, "go"), fakeGoCommand)
	writeFixtureFile(t, filepath.Join(bin, "gofmt"), "#!/usr/bin/env bash\nexit 0\n")
	writeFixtureFile(t, filepath.Join(bin, "golangci-lint"), fakeLintCommand)
	writeFixtureFile(t, filepath.Join(bin, "govulncheck"), fakeLoggedCommand("govulncheck"))
	writeFixtureFile(t, filepath.Join(bin, "goreleaser"), fakeGoReleaserCommand)
	for _, name := range []string{"go", "gofmt", "golangci-lint", "govulncheck", "goreleaser"} {
		if err := os.Chmod(filepath.Join(bin, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	runGit(t, repo, "init", "-q")
	runGit(t, repo, "add", ".")
	runGit(t, repo, "-c", "user.name=Release Test", "-c", "user.email=release@example.test", "commit", "-qm", "fixture")
	return releaseValidationFixture{repo: repo, bin: bin, log: log, script: script}
}

func (fixture releaseValidationFixture) run(extra ...string) (string, error) {
	command := exec.Command("/usr/bin/env", "bash", fixture.script)
	command.Dir = fixture.repo
	command.Env = append(os.Environ(),
		"PATH="+fixture.bin+":/usr/bin:/bin",
		"FAKE_LOG="+fixture.log,
		"FAKE_REPO="+fixture.repo,
	)
	command.Env = append(command.Env, extra...)
	output, err := command.CombinedOutput()
	return string(output), err
}

func (fixture releaseValidationFixture) requireClean(t *testing.T) {
	t.Helper()
	command := exec.Command("git", "status", "--porcelain", "--untracked-files=all")
	command.Dir = fixture.repo
	output, err := command.CombinedOutput()
	if err != nil || len(output) != 0 {
		t.Fatalf("fixture changed: err=%v status=%s", err, output)
	}
}

func (fixture releaseValidationFixture) readLog(t *testing.T) string {
	t.Helper()
	contents, err := os.ReadFile(fixture.log)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return string(contents)
}

func writeFixtureFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
}

func fakeLoggedCommand(name string) string {
	return fmt.Sprintf("#!/usr/bin/env bash\nprintf '%s %%s\\n' \"$*\" >>\"$FAKE_LOG\"\n", name)
}

const fakeGoCommand = `#!/usr/bin/env bash
set -euo pipefail
printf 'go %s\n' "$*" >>"$FAKE_LOG"
if [[ "$*" == "env GOVERSION" ]]; then
	printf 'go%s\n' "${FAKE_GO_VERSION:-1.25.12}"
	exit 0
fi
if [[ -n "${FAKE_GO_FAIL_MATCH:-}" && "$*" == *"$FAKE_GO_FAIL_MATCH"* ]]; then
	exit 42
fi
if [[ "$*" == run\ ./internal/tools/docgen* ]]; then
	output=${!#}
	mkdir -p "$output"
	cp "$FAKE_REPO/docs/cli/fixture.md" "$output/fixture.md"
	if [[ "${FAKE_STALE_DOCS:-}" == 1 ]]; then
		printf 'stale\n' >>"$output/fixture.md"
	fi
fi
if [[ "$*" == run\ ./internal/tools/schemacomments* ]]; then
	output=${!#}
	cp "$FAKE_REPO/internal/wire/schema_comments.json" "$output"
fi
`

const fakeLintCommand = `#!/usr/bin/env bash
if [[ "${1:-}" == version ]]; then
	printf 'golangci-lint has version 2.12.2 built with go1.25.12\n'
	exit 0
fi
printf 'golangci-lint %s\n' "$*" >>"$FAKE_LOG"
`

const fakeGoReleaserCommand = `#!/usr/bin/env bash
if [[ "${1:-}" == --version ]]; then
	printf 'GitVersion: v2.16.0\n'
	exit 0
fi
printf 'goreleaser %s\n' "$*" >>"$FAKE_LOG"
`
