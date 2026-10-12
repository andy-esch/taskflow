// Package reviewworkspace drives scripts/isolated-review-workspace.sh as a black
// box against disposable Git repositories. The helper stays a standalone shell
// tool; these tests only pin its observable safety contract.
package reviewworkspace

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const (
	deliverableRel = "planning/audits/review.md"
	pushGuardURL   = "isolated-review-push-disabled://source-is-read-only"
)

type fixture struct {
	root         string
	source       string
	parent       string
	globalConfig string
	env          []string
}

// newFixture builds a source repository with staged, unstaged, deleted, and
// untracked state plus a second branch, so an unguarded push to a non-checked-out
// ref would succeed. Git's global and system configuration are redirected into
// the fixture and inherited GIT_* variables are dropped, so no command can reach
// the developer's configuration or checkout.
func newFixture(t *testing.T) fixture {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := fixture{
		root:         root,
		source:       filepath.Join(root, "source"),
		parent:       filepath.Join(root, "sandboxes"),
		globalConfig: filepath.Join(root, "home", "gitconfig"),
	}
	for _, dir := range []string{f.source, f.parent, filepath.Join(root, "home"), filepath.Join(root, "tmp")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeFile(t, f.globalConfig, "[user]\n\tname = Review Fixture\n\temail = review@example.test\n", 0o644)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "GIT_") && !strings.HasPrefix(entry, "HOME=") && !strings.HasPrefix(entry, "TMPDIR=") {
			f.env = append(f.env, entry)
		}
	}
	f.env = append(f.env,
		"HOME="+filepath.Join(root, "home"),
		"TMPDIR="+filepath.Join(root, "tmp"),
		"GIT_CONFIG_GLOBAL="+f.globalConfig,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_TERMINAL_PROMPT=0",
	)

	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	script, err := os.ReadFile(filepath.Join(repositoryRoot, "scripts", "isolated-review-workspace.sh"))
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(f.source, "scripts", "isolated-review-workspace.sh"), string(script), 0o755)
	writeFile(t, filepath.Join(f.source, deliverableRel), "# Review\n\n## Reviewer report\n\nNot yet performed.\n", 0o644)
	writeFile(t, filepath.Join(f.source, "src", "kept.txt"), "kept\n", 0o644)
	writeFile(t, filepath.Join(f.source, "src", "staged.txt"), "staged\n", 0o644)
	writeFile(t, filepath.Join(f.source, "src", "deleted.txt"), "deleted\n", 0o644)
	f.git(t, f.source, "init", "-q", "-b", "main")
	f.git(t, f.source, "add", ".")
	f.git(t, f.source, "commit", "-qm", "fixture")
	f.git(t, f.source, "branch", "feature")

	writeFile(t, filepath.Join(f.source, "src", "kept.txt"), "kept, edited\n", 0o644)
	writeFile(t, filepath.Join(f.source, "src", "staged.txt"), "staged, edited\n", 0o644)
	f.git(t, f.source, "add", "src/staged.txt")
	if err := os.Remove(filepath.Join(f.source, "src", "deleted.txt")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(f.source, "src", "untracked.txt"), "untracked\n", 0o644)
	return f
}

func (f fixture) gitResult(dir string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	command.Dir = dir
	command.Env = f.env
	output, err := command.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func (f fixture) git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	output, err := f.gitResult(dir, args...)
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return output
}

// helper runs the copy of the script inside root, as the protocol does: create
// uses the source checkout's copy, verify and transfer use the sandbox's copy.
func (f fixture) helper(root string, args ...string) (string, error) {
	command := exec.Command("/usr/bin/env", append([]string{"bash", filepath.Join(root, "scripts", "isolated-review-workspace.sh")}, args...)...)
	command.Dir = f.root
	command.Env = f.env
	output, err := command.CombinedOutput()
	return string(output), err
}

func (f fixture) create(t *testing.T) string {
	t.Helper()
	output, err := f.helper(f.source, "create", "--source", f.source, "--deliverable", deliverableRel,
		"--sandbox-parent", f.parent, "--print-path")
	if err != nil {
		t.Fatalf("create failed: %v\n%s", err, output)
	}
	sandbox := strings.TrimSpace(output)
	if filepath.Dir(sandbox) != f.parent {
		t.Fatalf("sandbox %q is not under %q", sandbox, f.parent)
	}
	return sandbox
}

func (f fixture) workspace(sandbox, command string) (string, error) {
	return f.helper(sandbox, command, "--sandbox", sandbox)
}

// sourceState captures everything a push or a stray write could change in the
// source checkout: refs, Git administration bytes, objects, and working files.
type sourceState struct {
	refs, config, head, index, objects, files string
}

func (f fixture) sourceState(t *testing.T) sourceState {
	t.Helper()
	gitDir := filepath.Join(f.source, ".git")
	return sourceState{
		refs:    f.git(t, f.source, "for-each-ref", "--format=%(objectname) %(refname)"),
		config:  readFile(t, filepath.Join(gitDir, "config")),
		head:    readFile(t, filepath.Join(gitDir, "HEAD")),
		index:   readFile(t, filepath.Join(gitDir, "index")),
		objects: treeDigest(t, filepath.Join(gitDir, "objects"), ""),
		files:   treeDigest(t, f.source, ".git"),
	}
}

func requireSameSource(t *testing.T, before, after sourceState) {
	t.Helper()
	for _, field := range []struct{ name, before, after string }{
		{"refs", before.refs, after.refs},
		{"config", before.config, after.config},
		{"HEAD", before.head, after.head},
		{"index", before.index, after.index},
		{"objects", before.objects, after.objects},
		{"working files", before.files, after.files},
	} {
		if field.before != field.after {
			t.Errorf("source %s changed:\nbefore: %s\nafter:  %s", field.name, field.before, field.after)
		}
	}
}

func TestPushGuardRefusesOrdinaryPushesWithoutChangingSource(t *testing.T) {
	f := newFixture(t)
	sandbox := f.create(t)
	if got := f.git(t, sandbox, "remote", "get-url", "--push", "--all", "origin"); got != pushGuardURL {
		t.Fatalf("origin push URL = %q, want only %q", got, pushGuardURL)
	}
	if got := f.git(t, sandbox, "remote", "get-url", "origin"); got != f.source {
		t.Fatalf("origin fetch URL = %q, want the source %q", got, f.source)
	}
	f.git(t, sandbox, "tag", "review-tag")
	globalBefore := readFile(t, f.globalConfig)
	before := f.sourceState(t)

	for _, push := range [][]string{
		{"push"},
		{"push", "origin"},
		{"push", "origin", "HEAD:refs/heads/review-leak"},
		{"push", "origin", "HEAD:refs/heads/feature"},
		{"push", "--all", "origin"},
		{"push", "origin", "--tags"},
	} {
		output, err := f.gitResult(sandbox, push...)
		if err == nil || !strings.Contains(output, "isolated-review-push-disabled") {
			t.Errorf("git %s should fail at the push guard; err=%v output:\n%s", strings.Join(push, " "), err, output)
		}
	}
	requireSameSource(t, before, f.sourceState(t))
	if readFile(t, f.globalConfig) != globalBefore {
		t.Error("user-global Git configuration changed")
	}
	if strings.Contains(before.config, pushGuardURL) {
		t.Error("the guard was written to the source configuration")
	}

	// Control and documented limit: the same source accepts a push to another
	// branch when the sandbox names it explicitly. The guard prevents accidents;
	// it does not contain a deliberate explicit-URL push.
	f.git(t, sandbox, "push", f.source, "HEAD:refs/heads/explicit-url-limit")
	if !strings.Contains(f.git(t, f.source, "for-each-ref", "refs/heads/explicit-url-limit"), "explicit-url-limit") {
		t.Fatal("control push did not reach the source; the refusals above prove nothing")
	}
}

func TestVerifyAndTransferRejectAWeakenedPushGuard(t *testing.T) {
	for _, tc := range []struct {
		name   string
		weaken func(f fixture, sandbox string) []string
	}{
		{"push URL removed", func(f fixture, sandbox string) []string {
			return []string{"config", "--unset-all", "remote.origin.pushurl"}
		}},
		{"push URL repointed at source", func(f fixture, sandbox string) []string {
			return []string{"remote", "set-url", "--push", "origin", f.source}
		}},
		{"second push URL", func(f fixture, sandbox string) []string {
			return []string{"config", "--add", "remote.origin.pushurl", f.source}
		}},
		{"disabled URL rewritten", func(f fixture, sandbox string) []string {
			return []string{"config", "url." + f.source + ".insteadOf", pushGuardURL}
		}},
		{"unguarded remote added", func(f fixture, sandbox string) []string {
			return []string{"remote", "add", "leak", f.source}
		}},
		{"origin removed", func(f fixture, sandbox string) []string {
			return []string{"remote", "remove", "origin"}
		}},
		{"push default names a path", func(f fixture, sandbox string) []string {
			return []string{"config", "remote.pushDefault", f.source}
		}},
		{"branch push remote names a path", func(f fixture, sandbox string) []string {
			return []string{"config", "branch.main.pushRemote", f.source}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			sandbox := f.create(t)
			appendFile(t, filepath.Join(sandbox, deliverableRel), "\nReviewed.\n")
			if output, err := f.workspace(sandbox, "verify"); err != nil {
				t.Fatalf("verify failed before weakening: %v\n%s", err, output)
			}
			globalBefore := readFile(t, f.globalConfig)
			before := f.sourceState(t)
			f.git(t, sandbox, tc.weaken(f, sandbox)...)

			for _, command := range []string{"verify", "transfer"} {
				output, err := f.workspace(sandbox, command)
				if err == nil || !strings.Contains(output, "push guard") || !strings.Contains(output, "git -C '"+sandbox+"'") {
					t.Errorf("%s should reject the weakened guard with a remediation command; err=%v output:\n%s", command, err, output)
				}
			}
			requireSameSource(t, before, f.sourceState(t))
			if readFile(t, f.globalConfig) != globalBefore {
				t.Error("user-global Git configuration changed")
			}
		})
	}
}

func TestPushGuardAllowsLocalUpstreams(t *testing.T) {
	f := newFixture(t)
	sandbox := f.create(t)
	// A local tracking branch records branch.<name>.remote=".", which pushes into
	// the sandbox itself rather than past the guard.
	f.git(t, sandbox, "branch", "--track", "local-copy", "main")
	if output, err := f.workspace(sandbox, "verify"); err != nil {
		t.Fatalf("verify rejected a local upstream: %v\n%s", err, output)
	}
}

func TestReviewWorkspaceTransfersOnlyTheDeliverable(t *testing.T) {
	f := newFixture(t)
	before := f.sourceState(t)
	otherFilesBefore := treeDigest(t, f.source, ".git", deliverableRel)
	sandbox := f.create(t)
	for path, want := range map[string]string{
		"src/kept.txt":      "kept, edited\n",
		"src/staged.txt":    "staged, edited\n",
		"src/untracked.txt": "untracked\n",
	} {
		if got := readFile(t, filepath.Join(sandbox, path)); got != want {
			t.Errorf("sandbox %s = %q, want captured source state %q", path, got, want)
		}
	}
	if _, err := os.Stat(filepath.Join(sandbox, "src", "deleted.txt")); !os.IsNotExist(err) {
		t.Errorf("sandbox kept a file deleted in the source: %v", err)
	}
	if got := f.git(t, sandbox, "rev-parse", "--absolute-git-dir"); got != filepath.Join(sandbox, ".git") {
		t.Errorf("sandbox Git directory = %q, want its own .git", got)
	}

	appendFile(t, filepath.Join(sandbox, deliverableRel), "\nNo findings.\n")
	output, err := f.workspace(sandbox, "verify")
	if err != nil || !strings.Contains(output, "push_guard="+pushGuardURL) || !strings.Contains(output, "transfer=pending") {
		t.Fatalf("verify result=%v output:\n%s", err, output)
	}
	output, err = f.workspace(sandbox, "transfer")
	if err != nil || !strings.Contains(output, "transfer=succeeded") {
		t.Fatalf("transfer result=%v output:\n%s", err, output)
	}
	if got, want := readFile(t, filepath.Join(f.source, deliverableRel)), readFile(t, filepath.Join(sandbox, deliverableRel)); got != want {
		t.Fatalf("source deliverable = %q, want the sandbox report %q", got, want)
	}

	after := f.sourceState(t)
	before.files, after.files = "", "" // the deliverable is the one permitted file change
	requireSameSource(t, before, after)
	if got := treeDigest(t, f.source, ".git", deliverableRel); got != otherFilesBefore {
		t.Errorf("transfer changed source files other than the deliverable:\nbefore: %s\nafter:  %s", otherFilesBefore, got)
	}
}

func writeFile(t *testing.T, path, contents string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), mode); err != nil {
		t.Fatal(err)
	}
}

func appendFile(t *testing.T, path, contents string) {
	t.Helper()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, writeErr := file.WriteString(contents)
	if closeErr := file.Close(); writeErr != nil || closeErr != nil {
		t.Fatalf("append %s: write=%v close=%v", path, writeErr, closeErr)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

// treeDigest lists every regular file under root with its content hash,
// skipping the named relative paths (directories or files).
func treeDigest(t *testing.T, root string, skip ...string) string {
	t.Helper()
	var lines []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		for _, skipped := range skip {
			if skipped != "" && relative == skipped {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lines = append(lines, fmt.Sprintf("%s %x", relative, sha256.Sum256(contents)))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}
