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
	"time"
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
	writeFile(t, filepath.Join(f.source, ".gitignore"), "ignored/\n", 0o644)
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

// helper runs the copy of the script inside root. The protocol always runs the
// source checkout's copy, which reviewers cannot mutate from their sandbox.
func (f fixture) helper(root string, args ...string) (string, error) {
	command := exec.Command("/usr/bin/env", append([]string{"bash", filepath.Join(root, "scripts", "isolated-review-workspace.sh")}, args...)...)
	command.Dir = f.root
	command.Env = f.env
	output, err := command.CombinedOutput()
	return string(output), err
}

func (f fixture) create(t *testing.T) string {
	t.Helper()
	return f.createFrom(t, f.source)
}

func (f fixture) createFrom(t *testing.T, source string) string {
	t.Helper()
	output, err := f.helper(source, "create", "--source", source, "--deliverable", deliverableRel,
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
	return f.helper(f.source, command, "--sandbox", sandbox)
}

// sharedAdmin captures the shared repository's refs, worktree registry and
// per-worktree administration, index, and objects, using only Git reads that
// never refresh an index, so a snapshot cannot cause the change it detects.
func (f fixture) sharedAdmin(t *testing.T) string {
	t.Helper()
	gitDir := filepath.Join(f.source, ".git")
	parts := []string{
		f.git(t, f.source, "--no-optional-locks", "for-each-ref", "--format=%(objectname) %(refname)"),
		f.git(t, f.source, "--no-optional-locks", "worktree", "list", "--porcelain"),
		readFile(t, filepath.Join(gitDir, "index")),
		treeDigest(t, filepath.Join(gitDir, "objects")),
	}
	if _, err := os.Stat(filepath.Join(gitDir, "worktrees")); err == nil {
		parts = append(parts, treeDigest(t, filepath.Join(gitDir, "worktrees")))
	}
	return strings.Join(parts, "\n--\n")
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

// Every refusal must also be repairable: the printed restore lines are run
// literally and verify must then pass, so advice cannot silently rot.
func TestVerifyAndTransferRejectAWeakenedPushGuard(t *testing.T) {
	config := func(args ...string) func(t *testing.T, f fixture, sandbox string) {
		return func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, append([]string{"config"}, args...)...)
		}
	}
	include := func(key string) func(t *testing.T, f fixture, sandbox string) {
		return func(t *testing.T, f fixture, sandbox string) {
			file := filepath.Join(f.root, "included.gitconfig")
			writeFile(t, file, "[remote]\n\tpushDefault = "+f.source+"\n", 0o644)
			f.git(t, sandbox, "config", key, file)
		}
	}
	for _, tc := range []struct {
		name   string
		weaken func(t *testing.T, f fixture, sandbox string)
	}{
		{"push URL removed", config("--unset-all", "remote.origin.pushurl")},
		{"push URL repointed at source", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "remote", "set-url", "--push", "origin", f.source)
		}},
		{"second push URL", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "config", "--add", "remote.origin.pushurl", f.source)
		}},
		{"disabled URL rewritten", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "config", "url."+f.source+".insteadOf", pushGuardURL)
		}},
		{"unguarded remote added", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "remote", "add", "leak", f.source)
		}},
		{"unguarded remote with a space in its name", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "config", "remote.a b.url", f.source)
		}},
		{"origin removed", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "remote", "remove", "origin")
		}},
		{"push default names a path", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "config", "remote.pushDefault", f.source)
		}},
		{"two push defaults", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "config", "--add", "remote.pushDefault", "origin")
			f.git(t, sandbox, "config", "--add", "remote.pushDefault", f.source)
		}},
		{"branch push remote names a path", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "config", "branch.main.pushRemote", f.source)
		}},
		{"branch remote names a path", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "config", "branch.main.remote", f.source)
		}},
		{"branch name with an apostrophe", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "config", "branch.it's.pushRemote", f.source)
		}},
		{"included push default", include("include.path")},
		{"conditionally included push default", include("includeIf.onbranch:main.path")},
		{"worktree-scoped push default", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "config", "core.repositoryFormatVersion", "1")
			f.git(t, sandbox, "config", "extensions.worktreeConfig", "true")
			f.git(t, sandbox, "config", "--worktree", "remote.pushDefault", f.source)
		}},
		{"legacy remotes file", func(t *testing.T, f fixture, sandbox string) {
			writeFile(t, filepath.Join(sandbox, ".git", "remotes", "src"), "URL: "+f.source+"\n", 0o644)
		}},
		{"symlinked legacy remotes file", func(t *testing.T, f fixture, sandbox string) {
			file := filepath.Join(f.root, "legacy-remote")
			writeFile(t, file, "URL: "+f.source+"\n", 0o644)
			if err := os.MkdirAll(filepath.Join(sandbox, ".git", "remotes"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(file, filepath.Join(sandbox, ".git", "remotes", "src")); err != nil {
				t.Fatal(err)
			}
		}},
		{"legacy remotes file with a newline in its name", func(t *testing.T, f fixture, sandbox string) {
			writeFile(t, filepath.Join(sandbox, ".git", "remotes", "a\nb"), "URL: "+f.source+"\n", 0o644)
		}},
		{"dormant conditional include", func(t *testing.T, f fixture, sandbox string) {
			file := filepath.Join(f.root, "notes.gitconfig")
			writeFile(t, file, "[remote]\n\tpushDefault = "+f.source+"\n", 0o644)
			f.git(t, sandbox, "config", "includeIf.onbranch:notes/**.path", file)
		}},
		{"push URL from an included file", func(t *testing.T, f fixture, sandbox string) {
			file := filepath.Join(f.root, "pushurl.gitconfig")
			writeFile(t, file, "[remote \"origin\"]\n\tpushurl = "+f.source+"\n", 0o644)
			f.git(t, sandbox, "config", "include.path", file)
		}},
		{"push URL in worktree config", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "config", "core.repositoryFormatVersion", "1")
			f.git(t, sandbox, "config", "extensions.worktreeConfig", "true")
			f.git(t, sandbox, "config", "--worktree", "--add", "remote.origin.pushurl", f.source)
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
			tc.weaken(t, f, sandbox)

			var restore []string
			for _, command := range []string{"verify", "transfer"} {
				output, err := f.workspace(sandbox, command)
				if err == nil || !strings.Contains(output, "push guard") {
					t.Fatalf("%s should reject the weakened guard; err=%v output:\n%s", command, err, output)
				}
				restore = restore[:0]
				for _, line := range strings.Split(output, "\n") {
					if command, ok := strings.CutPrefix(line, "restore: "); ok {
						restore = append(restore, command)
					}
				}
				if len(restore) == 0 {
					t.Fatalf("%s printed no restore command:\n%s", command, output)
				}
			}
			requireSameSource(t, before, f.sourceState(t))
			if readFile(t, f.globalConfig) != globalBefore {
				t.Error("user-global Git configuration changed")
			}

			for _, command := range restore {
				shell := exec.Command("/usr/bin/env", "bash", "-c", command)
				shell.Dir = f.root
				shell.Env = f.env
				if output, err := shell.CombinedOutput(); err != nil {
					t.Fatalf("printed restore command failed: %s\n%v\n%s", command, err, output)
				}
			}
			if output, err := f.workspace(sandbox, "verify"); err != nil {
				t.Fatalf("verify still refuses after running the printed restore commands %q: %v\n%s", restore, err, output)
			}
		})
	}
}

// A linked worktree's .git is a file naming the shared repository's
// administration. If the overlay copied it over the clone's own .git, the
// baseline add and commit would land in the source repository.
func TestLinkedWorktreeSourceStaysIndependent(t *testing.T) {
	f := newFixture(t)
	linked := filepath.Join(f.root, "linked")
	f.git(t, f.source, "worktree", "add", "-q", "-b", "review-branch", linked)
	appendFile(t, filepath.Join(linked, "src", "kept.txt"), "linked branch work\n")
	f.git(t, linked, "commit", "-qam", "linked branch work")
	writeFile(t, filepath.Join(linked, "src", "linked-untracked.txt"), "untracked in the worktree\n", 0o644)
	linkedHead := f.git(t, linked, "rev-parse", "HEAD")
	mainReport := readFile(t, filepath.Join(f.source, deliverableRel))

	before := f.sharedAdmin(t)

	sandbox := f.createFrom(t, linked)
	if after := f.sharedAdmin(t); after != before {
		t.Fatalf("creating a workspace from a linked worktree changed the shared repository:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if info, err := os.Lstat(filepath.Join(sandbox, ".git")); err != nil || !info.IsDir() {
		t.Fatalf("sandbox .git must be its own directory: info=%v err=%v", info, err)
	}
	if got := f.git(t, sandbox, "rev-parse", "--absolute-git-dir"); got != filepath.Join(sandbox, ".git") {
		t.Fatalf("sandbox Git directory = %q", got)
	}
	if got := f.git(t, sandbox, "rev-parse", "HEAD^"); got != linkedHead {
		t.Fatalf("baseline parent = %s, want the linked worktree's branch head %s", got, linkedHead)
	}
	if got := readFile(t, filepath.Join(sandbox, "src", "kept.txt")); !strings.Contains(got, "linked branch work") {
		t.Errorf("sandbox did not capture the linked branch commit: %q", got)
	}
	if _, err := os.Stat(filepath.Join(sandbox, "src", "linked-untracked.txt")); err != nil {
		t.Errorf("sandbox did not capture worktree-untracked state: %v", err)
	}

	appendFile(t, filepath.Join(sandbox, deliverableRel), "\nReviewed from a linked worktree.\n")
	for _, command := range []string{"verify", "transfer", "receipt"} {
		if output, err := f.workspace(sandbox, command); err != nil {
			t.Fatalf("%s from a linked-worktree source: %v\n%s", command, err, output)
		}
	}
	if got, want := readFile(t, filepath.Join(linked, deliverableRel)), readFile(t, filepath.Join(sandbox, deliverableRel)); got != want {
		t.Errorf("linked worktree report = %q, want the delivered %q", got, want)
	}
	if got := readFile(t, filepath.Join(f.source, deliverableRel)); got != mainReport {
		t.Errorf("transfer wrote the main checkout's report instead of the linked worktree's: %q", got)
	}
}

// An ignored nested worktree (like this repository's .claude/worktrees/*) holds
// a .git file naming the shared repository. Its content may be copied, but never
// that metadata, and verify must refuse it if it reappears in the workspace.
func TestNestedWorktreeIsCopiedWithoutItsGitMetadata(t *testing.T) {
	f := newFixture(t)
	nested := filepath.Join(f.source, "ignored", "other")
	f.git(t, f.source, "worktree", "add", "-q", "-b", "other-work", nested)
	before := f.sharedAdmin(t)

	sandbox := f.create(t)
	copied := filepath.Join(sandbox, "ignored", "other")
	if _, err := os.Stat(filepath.Join(copied, "src", "kept.txt")); err != nil {
		t.Fatalf("ignored nested content was not copied: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(copied, ".git")); !os.IsNotExist(err) {
		t.Fatalf("nested .git was copied into the workspace: %v", err)
	}
	// Without its .git, a Git command there resolves to the workspace itself.
	if got := f.git(t, copied, "rev-parse", "--show-toplevel"); got != sandbox {
		t.Fatalf("nested copy resolves to %q, want the workspace %q", got, sandbox)
	}
	appendFile(t, filepath.Join(sandbox, deliverableRel), "\nReviewed.\n")
	if output, err := f.workspace(sandbox, "verify"); err != nil {
		t.Fatalf("verify: %v\n%s", err, output)
	}

	// A reviewer accidentally restores the pointer: verify must not attest it.
	writeFile(t, filepath.Join(copied, ".git"), readFile(t, filepath.Join(nested, ".git")), 0o644)
	for _, command := range []string{"verify", "transfer"} {
		output, err := f.workspace(sandbox, command)
		if err == nil || !strings.Contains(output, "nested Git metadata") {
			t.Errorf("%s should refuse nested Git metadata; err=%v output:\n%s", command, err, output)
		}
	}
	if after := f.sharedAdmin(t); after != before {
		t.Fatalf("the shared repository changed:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// A non-ignored nested repository cannot be fingerprinted or reproduced, so
// create must refuse it outright rather than skip it and attest a capture.
func TestNestedRepositoryInTheSourceIsRefused(t *testing.T) {
	f := newFixture(t)
	f.git(t, f.source, "worktree", "add", "-q", "-b", "nested-work", filepath.Join(f.source, "nested"))
	before := f.sharedAdmin(t)
	output, err := f.helper(f.source, "create", "--source", f.source, "--deliverable", deliverableRel,
		"--sandbox-parent", f.parent, "--print-path")
	if err == nil || !strings.Contains(output, "cannot fingerprint the source") {
		t.Fatalf("create should refuse a nested repository; err=%v output:\n%s", err, output)
	}
	if entries, err := os.ReadDir(f.parent); err != nil || len(entries) != 0 {
		t.Fatalf("a refused create left a workspace behind: %v %v", entries, err)
	}
	if after := f.sharedAdmin(t); after != before {
		t.Fatalf("the shared repository changed:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// rsync copies symlinks verbatim, so an absolute or escaping link lets a Git
// command or a write through it reach the source; so does a case-variant .GIT
// gitfile on a case-insensitive filesystem. create and verify must refuse all.
func TestSymlinksLeadingOutOfTheWorkspaceAreRefused(t *testing.T) {
	createCases := []struct{ name, link, target string }{
		{"absolute link in an ignored directory", "ignored/source-root", ""},
		{"relative link climbing out", "src/escape", "../../.."},
	}
	for _, tc := range createCases {
		t.Run("create: "+tc.name, func(t *testing.T) {
			f := newFixture(t)
			target := tc.target
			if target == "" {
				target = f.source
			}
			if err := os.MkdirAll(filepath.Dir(filepath.Join(f.source, tc.link)), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, filepath.Join(f.source, tc.link)); err != nil {
				t.Fatal(err)
			}
			before := f.sharedAdmin(t)
			output, err := f.helper(f.source, "create", "--source", f.source, "--deliverable", deliverableRel,
				"--sandbox-parent", f.parent, "--print-path")
			if err == nil || !strings.Contains(output, "points outside the workspace") {
				t.Fatalf("create should refuse an escaping symlink; err=%v output:\n%s", err, output)
			}
			if entries, _ := os.ReadDir(f.parent); len(entries) != 0 {
				t.Fatalf("a refused create left a workspace behind: %v", entries)
			}
			if f.sharedAdmin(t) != before {
				t.Fatal("the shared repository changed")
			}
		})
	}

	for _, tc := range []struct {
		name  string
		plant func(t *testing.T, f fixture, sandbox string)
		want  string
	}{
		{"absolute link planted after create", func(t *testing.T, f fixture, sandbox string) {
			if err := os.Symlink(f.source, filepath.Join(sandbox, "ignored-up")); err != nil {
				t.Fatal(err)
			}
		}, "points outside the workspace"},
		{"case-variant gitfile planted after create", func(t *testing.T, f fixture, sandbox string) {
			writeFile(t, filepath.Join(sandbox, "ignored", "up", ".GIT"), "gitdir: "+filepath.Join(f.source, ".git")+"\n", 0o644)
		}, "nested Git metadata"},
	} {
		t.Run("verify: "+tc.name, func(t *testing.T) {
			f := newFixture(t)
			sandbox := f.create(t)
			appendFile(t, filepath.Join(sandbox, deliverableRel), "\nReviewed.\n")
			tc.plant(t, f, sandbox)
			for _, command := range []string{"verify", "transfer"} {
				if output, err := f.workspace(sandbox, command); err == nil || !strings.Contains(output, tc.want) {
					t.Errorf("%s should refuse; err=%v output:\n%s", command, err, output)
				}
			}
		})
	}
}

// Links that stay inside the repository are ordinary content: a dangling link
// or one to a directory is captured by its link text, as git add stores it.
func TestInTreeSymlinksAreAccepted(t *testing.T) {
	f := newFixture(t)
	for link, target := range map[string]string{"src/latest": "does-not-exist", "src/plan-link": "../planning"} {
		if err := os.Symlink(target, filepath.Join(f.source, link)); err != nil {
			t.Fatal(err)
		}
	}
	sandbox := f.create(t)
	for link, target := range map[string]string{"src/latest": "does-not-exist", "src/plan-link": "../planning"} {
		if got, err := os.Readlink(filepath.Join(sandbox, link)); err != nil || got != target {
			t.Errorf("sandbox %s = %q (%v), want link to %q", link, got, err, target)
		}
	}
	appendFile(t, filepath.Join(sandbox, deliverableRel), "\nReviewed.\n")
	if output, err := f.workspace(sandbox, "verify"); err != nil {
		t.Fatalf("verify: %v\n%s", err, output)
	}
}

// Workspaces are disposable once processed, so by default they live in /tmp,
// where the operating system's temporary-file cleanup reclaims leftovers.
func TestWorkspacesDefaultToTmp(t *testing.T) {
	tmp, err := filepath.EvalSymlinks("/tmp")
	if err != nil {
		t.Fatal(err)
	}
	f := newFixture(t)
	output, err := f.helper(f.source, "create", "--source", f.source, "--deliverable", deliverableRel, "--print-path")
	if err != nil {
		t.Fatalf("create: %v\n%s", err, output)
	}
	sandbox := strings.TrimSpace(output)
	t.Cleanup(func() { _ = os.RemoveAll(sandbox) })
	if filepath.Dir(sandbox) != tmp {
		t.Fatalf("default workspace %q is not directly under %q", sandbox, tmp)
	}
}

// The nested-metadata scan once used the sandbox path as a find pattern, so a
// parent directory containing brackets made create refuse its own root .git.
func TestPatternCharactersInTheSandboxParentPath(t *testing.T) {
	f := newFixture(t)
	parent := filepath.Join(f.root, "par[1]*?")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	output, err := f.helper(f.source, "create", "--source", f.source, "--deliverable", deliverableRel,
		"--sandbox-parent", parent, "--print-path")
	if err != nil {
		t.Fatalf("create under %q: %v\n%s", parent, err, output)
	}
	sandbox := strings.TrimSpace(output)
	appendFile(t, filepath.Join(sandbox, deliverableRel), "\nReviewed.\n")
	if output, err := f.workspace(sandbox, "verify"); err != nil {
		t.Fatalf("verify under %q: %v\n%s", parent, err, output)
	}
}

// A content-identical but stat-dirty tracked file makes porcelain diff rewrite
// the index. Capturing the source must leave its index bytes untouched, both
// for a plain checkout and for a linked worktree's shared administration.
func TestCreateLeavesAStatDirtySourceIndexUntouched(t *testing.T) {
	f := newFixture(t)
	linked := filepath.Join(f.root, "linked")
	f.git(t, f.source, "worktree", "add", "-q", "-b", "review-branch", linked)
	later := time.Now().Add(2 * time.Hour)
	for _, root := range []string{f.source, linked} {
		if err := os.Chtimes(filepath.Join(root, "scripts", "isolated-review-workspace.sh"), later, later); err != nil {
			t.Fatal(err)
		}
	}
	for _, source := range []struct{ root, index string }{
		{f.source, filepath.Join(f.source, ".git", "index")},
		{linked, filepath.Join(f.source, ".git", "worktrees", "linked", "index")},
	} {
		before := readFile(t, source.index)
		f.createFrom(t, source.root)
		if readFile(t, source.index) != before {
			t.Errorf("create from %s rewrote the source index %s", source.root, source.index)
		}
	}
}

// Before the fix, an overlay could replace the clone's .git with a linked
// worktree's .git file; GNU rsync does this with the old exclude. A shim
// reproduces that on any platform: create must stop before its first Git write.
func TestReplacedGitDirectoryIsCaughtBeforeAnyWorkspaceWrite(t *testing.T) {
	realRsync, err := exec.LookPath("rsync")
	if err != nil {
		t.Fatal(err)
	}
	f := newFixture(t)
	linked := filepath.Join(f.root, "linked")
	f.git(t, f.source, "worktree", "add", "-q", "-b", "review-branch", linked)
	shimDir := filepath.Join(f.root, "rsync-shim")
	writeFile(t, filepath.Join(shimDir, "rsync"), "#!/usr/bin/env bash\nset -e\n\"$REAL_RSYNC\" \"$@\"\nfor destination; do :; done\nrm -rf \"$destination/.git\"\ncp \"$REPLACEMENT_GIT_FILE\" \"$destination/.git\"\n", 0o755)
	shimmed := f
	shimmed.env = append(append([]string{}, f.env...),
		"PATH="+shimDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"REAL_RSYNC="+realRsync,
		"REPLACEMENT_GIT_FILE="+filepath.Join(linked, ".git"),
	)
	before := f.sharedAdmin(t)

	output, err := shimmed.helper(linked, "create", "--source", linked, "--deliverable", deliverableRel,
		"--sandbox-parent", f.parent, "--print-path")
	if err == nil || !strings.Contains(output, "not an independent in-tree .git directory") {
		t.Fatalf("create should stop at the independence check; err=%v output:\n%s", err, output)
	}
	if after := f.sharedAdmin(t); after != before {
		t.Fatalf("the baseline reached the shared repository:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if entries, err := os.ReadDir(f.parent); err != nil || len(entries) != 0 {
		t.Fatalf("a refused create left a workspace behind: %v %v", entries, err)
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

func receiptPath(sandbox string) string {
	return filepath.Join(sandbox, ".git", "isolated-review-workspace", "transfer-receipt")
}

func requireNoReceipt(t *testing.T, f fixture, sandbox string) {
	t.Helper()
	for _, path := range []string{receiptPath(sandbox), receiptPath(sandbox) + ".pending"} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s exists without a successful transfer: %v", filepath.Base(path), err)
		}
	}
	output, err := f.workspace(sandbox, "receipt")
	if err == nil || !strings.Contains(output, "no transfer receipt") {
		t.Errorf("receipt should refuse before a successful transfer; err=%v output:\n%s", err, output)
	}
}

// The report is final before transfer, so it can carry only the verification
// attestation. Delivery evidence is the helper's receipt, which exists only
// after the transfer and fails once the source no longer holds those bytes.
func TestTransferReceiptIsTheOnlyDeliveryEvidence(t *testing.T) {
	f := newFixture(t)
	sandbox := f.create(t)
	appendFile(t, filepath.Join(sandbox, deliverableRel), "\nNo findings.\n")

	output, err := f.workspace(sandbox, "verify")
	if err != nil || !strings.Contains(output, "verify=passed") || !strings.Contains(output, "transfer=pending") ||
		strings.Contains(output, "transfer=succeeded") {
		t.Fatalf("verify must attest verification only; err=%v output:\n%s", err, output)
	}
	// A reviewer pastes that attestation into the report; it is still not delivery.
	appendFile(t, filepath.Join(sandbox, deliverableRel), "\n```text\n"+output+"```\n")
	requireNoReceipt(t, f, sandbox)

	output, err = f.workspace(sandbox, "transfer")
	if err != nil || !strings.Contains(output, "transfer=succeeded") {
		t.Fatalf("transfer result=%v output:\n%s", err, output)
	}
	delivered := f.git(t, f.source, "hash-object", "--", deliverableRel)
	if !strings.Contains(output, "transferred_blob="+delivered) || readFile(t, receiptPath(sandbox)) != output {
		t.Fatalf("transfer must print the receipt it recorded for blob %s:\n%s", delivered, output)
	}

	output, err = f.workspace(sandbox, "receipt")
	if err != nil {
		t.Fatalf("receipt after transfer: %v\n%s", err, output)
	}
	for _, want := range []string{"verify=passed", "transfer=succeeded", "transferred_blob=" + delivered, "delivery=confirmed"} {
		if !strings.Contains(output, want) {
			t.Errorf("receipt omitted %q:\n%s", want, output)
		}
	}
	for _, command := range []string{"verify", "transfer"} {
		output, err := f.workspace(sandbox, command)
		if err == nil || !strings.Contains(output, "already transferred") || !strings.Contains(output, "receipt --sandbox") {
			t.Errorf("%s after transfer should point to the receipt; err=%v output:\n%s", command, err, output)
		}
	}
	if got := f.git(t, f.source, "hash-object", "--", deliverableRel); got != delivered {
		t.Fatalf("a refused repeat changed the delivered report: %s -> %s", delivered, got)
	}

	// A workspace edit after transfer means the retained sandbox no longer shows
	// what was delivered, so the receipt must stop vouching for it.
	reportPath := filepath.Join(sandbox, deliverableRel)
	deliveredReport := readFile(t, reportPath)
	appendFile(t, reportPath, "\nLate edit.\n")
	output, err = f.workspace(sandbox, "receipt")
	if err == nil || !strings.Contains(output, "workspace deliverable changed after transfer") {
		t.Fatalf("receipt should refuse after a workspace edit; err=%v output:\n%s", err, output)
	}
	writeFile(t, reportPath, deliveredReport, 0o644)

	// The owner may triage after confirming receipt; the receipt then stops
	// vouching for the current source bytes rather than silently passing.
	appendFile(t, filepath.Join(f.source, deliverableRel), "\nTriaged.\n")
	output, err = f.workspace(sandbox, "receipt")
	if err == nil || !strings.Contains(output, "no longer matches the transferred report") {
		t.Fatalf("receipt should refuse after the source changed; err=%v output:\n%s", err, output)
	}
}

func TestFailedTransfersRecordNoReceipt(t *testing.T) {
	for _, tc := range []struct {
		name    string
		disturb func(t *testing.T, f fixture, sandbox string)
		want    string
	}{
		{"source deliverable drift", func(t *testing.T, f fixture, sandbox string) {
			appendFile(t, filepath.Join(f.source, deliverableRel), "\nOwner edit.\n")
		}, "source deliverable changed"},
		{"unrelated workspace change", func(t *testing.T, f fixture, sandbox string) {
			appendFile(t, filepath.Join(sandbox, "src", "kept.txt"), "probe\n")
		}, "changes outside the deliverable"},
		{"reviewer commit", func(t *testing.T, f fixture, sandbox string) {
			f.git(t, sandbox, "commit", "-qam", "reviewer commit")
		}, "reviewer commits are not permitted"},
		{"source write failure", func(t *testing.T, f fixture, sandbox string) {
			if os.Geteuid() == 0 {
				t.Skip("a read-only directory does not stop root")
			}
			dir := filepath.Dir(filepath.Join(f.source, deliverableRel))
			if err := os.Chmod(dir, 0o555); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
		}, "the source was not changed"},
		// Both reviewers reproduced this one after the rename; the pending record
		// now proves the state directory writable before the source is touched.
		{"unwritable state directory", func(t *testing.T, f fixture, sandbox string) {
			if os.Geteuid() == 0 {
				t.Skip("a read-only directory does not stop root")
			}
			dir := filepath.Dir(receiptPath(sandbox))
			if err := os.Chmod(dir, 0o500); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
		}, "the source was not changed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			sandbox := f.create(t)
			appendFile(t, filepath.Join(sandbox, deliverableRel), "\nNo findings.\n")
			tc.disturb(t, f, sandbox)
			sourceReport := readFile(t, filepath.Join(f.source, deliverableRel))

			output, err := f.workspace(sandbox, "transfer")
			if err == nil || !strings.Contains(output, tc.want) {
				t.Fatalf("transfer should fail closed with %q; err=%v output:\n%s", tc.want, err, output)
			}
			if strings.Contains(output, "transfer=succeeded") {
				t.Errorf("a failed transfer printed success:\n%s", output)
			}
			if got := readFile(t, filepath.Join(f.source, deliverableRel)); got != sourceReport {
				t.Errorf("a failed transfer changed the source report:\n%s", got)
			}
			requireNoReceipt(t, f, sandbox)
		})
	}
}

// withShims returns a fixture whose helper runs PATH cp, mv, and ln shims that
// inject the races and failures the reviewers reproduced at exact points:
//   - EDIT_BEFORE_STAGE: append to the workspace report just before staging;
//   - RUN_BEFORE_STAGE / RUN_AFTER_RENAME: run a concurrent helper command;
//   - KILL_AT_STAGE: SIGKILL the transfer mid-staging, so no trap runs;
//   - EDIT_AFTER_RENAME: an editor saves into the source after the rename;
//   - FAIL_RECEIPT_PUBLISH: the receipt cannot be linked into place.
func (f fixture) withShims(t *testing.T, env ...string) fixture {
	t.Helper()
	shimDir := filepath.Join(f.root, "shims")
	shimmed := f
	shimmed.env = append(append([]string{}, f.env...), "PATH="+shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, tool := range []string{"cp", "mv", "ln"} {
		real, err := exec.LookPath(tool)
		if err != nil {
			t.Fatal(err)
		}
		shimmed.env = append(shimmed.env, "REAL_"+strings.ToUpper(tool)+"="+real)
	}
	writeFile(t, filepath.Join(shimDir, "cp"), `#!/usr/bin/env bash
for destination; do :; done
if [[ "$destination" == *.review-transfer.* ]]; then
	source="${@: -2:1}"
	[[ -z "${EDIT_BEFORE_STAGE:-}" ]] || printf 'late edit\n' >>"$source"
	[[ -z "${RUN_BEFORE_STAGE:-}" ]] || bash -c "$RUN_BEFORE_STAGE" >>"$SHIM_LOG" 2>&1 || true
	[[ -z "${KILL_AT_STAGE:-}" ]] || { "$REAL_CP" "$@"; kill -9 "$PPID"; exit 1; }
fi
exec "$REAL_CP" "$@"
`, 0o755)
	writeFile(t, filepath.Join(shimDir, "mv"), `#!/usr/bin/env bash
for destination; do :; done
"$REAL_MV" "$@" || exit
if [[ "$destination" == "${SOURCE_REPORT:-}" ]]; then
	[[ -z "${EDIT_AFTER_RENAME:-}" ]] || printf 'owner save\n' >>"$destination"
	[[ -z "${RUN_AFTER_RENAME:-}" ]] || bash -c "$RUN_AFTER_RENAME" >>"$SHIM_LOG" 2>&1 || true
fi
`, 0o755)
	writeFile(t, filepath.Join(shimDir, "ln"), `#!/usr/bin/env bash
for destination; do :; done
if [[ -n "${FAIL_RECEIPT_PUBLISH:-}" && "$destination" == */transfer-receipt ]]; then exit 1; fi
exec "$REAL_LN" "$@"
`, 0o755)
	shimmed.env = append(shimmed.env,
		"SOURCE_REPORT="+filepath.Join(f.source, deliverableRel), "SHIM_LOG="+filepath.Join(f.root, "shim.log"))
	shimmed.env = append(shimmed.env, env...)
	return shimmed
}

// helperCommand is a shell command that runs the source copy of the helper.
func (f fixture) helperCommand(command, sandbox string) string {
	return fmt.Sprintf("bash %q %s --sandbox %q", filepath.Join(f.source, "scripts", "isolated-review-workspace.sh"), command, sandbox)
}

func exitCode(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return -1
}

// A failure after the rename used to leave a delivered report that every
// command described as undelivered. Exit 3 now says the source was replaced,
// and receipt records the delivery once both copies hold the transferred bytes.
func TestInterruptedTransferIsRecordedByReceipt(t *testing.T) {
	f := newFixture(t)
	sandbox := f.create(t)
	appendFile(t, filepath.Join(sandbox, deliverableRel), "\nNo findings.\n")
	output, err := f.withShims(t, "FAIL_RECEIPT_PUBLISH=1").workspace(sandbox, "transfer")
	if exitCode(err) != 3 || !strings.Contains(output, "receipt was not recorded") {
		t.Fatalf("transfer should exit 3 after delivering without a receipt; err=%v output:\n%s", err, output)
	}
	if readFile(t, filepath.Join(f.source, deliverableRel)) != readFile(t, filepath.Join(sandbox, deliverableRel)) {
		t.Fatal("the interrupted transfer did not deliver the report")
	}
	for _, command := range []string{"verify", "transfer"} {
		output, err := f.workspace(sandbox, command)
		if err == nil || !strings.Contains(output, "receipt was not recorded") || strings.Contains(output, "source deliverable changed") {
			t.Errorf("%s should point to receipt, not blame the owner; err=%v output:\n%s", command, err, output)
		}
	}
	output, err = f.workspace(sandbox, "receipt")
	delivered := f.git(t, f.source, "hash-object", "--", deliverableRel)
	if err != nil || !strings.Contains(output, "delivery=confirmed") || !strings.Contains(output, "transferred_blob="+delivered) ||
		!strings.Contains(output, "recorded_by=receipt") {
		t.Fatalf("receipt should record the interrupted delivery; err=%v output:\n%s", err, output)
	}
	if output, err := f.workspace(sandbox, "receipt"); err != nil || !strings.Contains(output, "delivery=confirmed") {
		t.Fatalf("the recorded receipt should keep confirming; err=%v output:\n%s", err, output)
	}
}

// If the source changes between the rename and its confirmation, nothing may
// later claim delivery: transfer exits 3 and receipt reports the uncertainty.
func TestUnconfirmedTransferIsNeverReportedAsDelivered(t *testing.T) {
	f := newFixture(t)
	sandbox := f.create(t)
	appendFile(t, filepath.Join(sandbox, deliverableRel), "\nNo findings.\n")
	shimmed := f.withShims(t, "EDIT_AFTER_RENAME=1")
	output, err := shimmed.workspace(sandbox, "transfer")
	if exitCode(err) != 3 || !strings.Contains(output, "no longer holds the transferred bytes") {
		t.Fatalf("transfer should exit 3 when the delivered bytes changed; err=%v output:\n%s", err, output)
	}
	output, err = f.workspace(sandbox, "receipt")
	if err == nil || !strings.Contains(output, "delivery uncertain") || strings.Contains(output, "delivery=confirmed") {
		t.Fatalf("receipt must not confirm an unconfirmed delivery; err=%v output:\n%s", err, output)
	}
	if output, err := f.workspace(sandbox, "transfer"); err == nil || !strings.Contains(output, "receipt was not recorded") {
		t.Fatalf("a retry should point to receipt; err=%v output:\n%s", err, output)
	}
}

// A concurrent receipt or verify used to delete the live pending record, so the
// transfer published an empty, unconfirmable receipt. Neither may now touch it,
// and publishing never overwrites a receipt another publisher recorded.
func TestConcurrentCommandsCannotCorruptATransfer(t *testing.T) {
	for _, tc := range []struct{ name, hook, command string }{
		{"receipt during staging", "RUN_BEFORE_STAGE", "receipt"},
		{"verify during staging", "RUN_BEFORE_STAGE", "verify"},
		{"receipt right after the rename", "RUN_AFTER_RENAME", "receipt"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			sandbox := f.create(t)
			appendFile(t, filepath.Join(sandbox, deliverableRel), "\nNo findings.\n")
			shimmed := f.withShims(t, tc.hook+"="+f.helperCommand(tc.command, sandbox))
			output, err := shimmed.workspace(sandbox, "transfer")
			delivered := f.git(t, f.source, "hash-object", "--", deliverableRel)
			if err != nil || !strings.Contains(output, "transfer=succeeded") || !strings.Contains(output, "transferred_blob="+delivered) {
				t.Fatalf("transfer should publish a complete receipt despite a concurrent %s; err=%v output:\n%s\nconcurrent:\n%s",
					tc.command, err, output, readFile(t, filepath.Join(f.root, "shim.log")))
			}
			if output, err := f.workspace(sandbox, "receipt"); err != nil || !strings.Contains(output, "delivery=confirmed") {
				t.Fatalf("receipt after the race: err=%v output:\n%s", err, output)
			}
		})
	}
}

// A transfer killed mid-staging runs no cleanup. Its leftovers must be reported
// with working restore lines rather than silently discarded or blamed on drift.
func TestKilledTransferLeavesRepairableState(t *testing.T) {
	f := newFixture(t)
	sandbox := f.create(t)
	appendFile(t, filepath.Join(sandbox, deliverableRel), "\nNo findings.\n")
	original := readFile(t, filepath.Join(f.source, deliverableRel))
	if _, err := f.withShims(t, "KILL_AT_STAGE=1").workspace(sandbox, "transfer"); err == nil {
		t.Fatal("the killed transfer reported success")
	}
	if readFile(t, filepath.Join(f.source, deliverableRel)) != original {
		t.Fatal("a transfer killed before its rename changed the source report")
	}
	output, err := f.workspace(sandbox, "verify")
	if err == nil || !strings.Contains(output, "interrupted before replacing the source report") || !strings.Contains(output, ".review-transfer.") {
		t.Fatalf("verify should report the leftovers; err=%v output:\n%s", err, output)
	}
	for _, line := range strings.Split(output, "\n") {
		if command, ok := strings.CutPrefix(line, "restore: "); ok {
			if out, err := f.runShell(command); err != nil {
				t.Fatalf("restore line %q failed: %v\n%s", command, err, out)
			}
		}
	}
	if output, err := f.workspace(sandbox, "transfer"); err != nil || !strings.Contains(output, "transfer=succeeded") {
		t.Fatalf("transfer after cleanup: err=%v output:\n%s", err, output)
	}
	if strays, _ := filepath.Glob(filepath.Join(f.source, deliverableRel) + ".review-transfer.*"); len(strays) != 0 {
		t.Fatalf("staged copies left in the source: %v", strays)
	}
}

// A workspace edit between hashing and staging must be refused before the
// source is touched (exit 2), not discovered after the rename (exit 3).
func TestWorkspaceEditDuringStagingLeavesTheSourceUnchanged(t *testing.T) {
	f := newFixture(t)
	sandbox := f.create(t)
	appendFile(t, filepath.Join(sandbox, deliverableRel), "\nNo findings.\n")
	original := readFile(t, filepath.Join(f.source, deliverableRel))
	output, err := f.withShims(t, "EDIT_BEFORE_STAGE=1").workspace(sandbox, "transfer")
	if exitCode(err) != 2 || !strings.Contains(output, "workspace report changed during transfer") {
		t.Fatalf("transfer should refuse with exit 2; err=%v output:\n%s", err, output)
	}
	if readFile(t, filepath.Join(f.source, deliverableRel)) != original {
		t.Fatal("a refused transfer changed the source report")
	}
	requireNoReceipt(t, f, sandbox)
	if strays, _ := filepath.Glob(filepath.Join(f.source, deliverableRel) + ".review-transfer.*"); len(strays) != 0 {
		t.Fatalf("staged copies left in the source: %v", strays)
	}
}

// A report copied by hand, or by a helper version without receipts, is not
// delivery evidence; both commands must say so rather than blame owner drift.
func TestHandCopiedReportIsNotVouchedFor(t *testing.T) {
	f := newFixture(t)
	sandbox := f.create(t)
	appendFile(t, filepath.Join(sandbox, deliverableRel), "\nNo findings.\n")
	writeFile(t, filepath.Join(f.source, deliverableRel), readFile(t, filepath.Join(sandbox, deliverableRel)), 0o644)
	if output, err := f.workspace(sandbox, "receipt"); err == nil || !strings.Contains(output, "cannot vouch for it") {
		t.Fatalf("receipt should refuse to vouch for a hand copy; err=%v output:\n%s", err, output)
	}
	if output, err := f.workspace(sandbox, "verify"); err == nil || !strings.Contains(output, "already holds this workspace's report") {
		t.Fatalf("verify should name the existing copy; err=%v output:\n%s", err, output)
	}
}

// runShell runs a script the way a reviewer's shell would, in the fixture root.
func (f fixture) runShell(script string) (string, error) {
	command := exec.Command("/usr/bin/env", "bash", "-c", script)
	command.Dir = f.root
	command.Env = f.env
	output, err := command.CombinedOutput()
	return string(output), err
}

// The generated brief is the reviewer's literal instructions. The generator runs
// for real with a fake CLI, and its handoff values and shell blocks are executed
// verbatim: once with a failing create, which must never fall through to the
// shared checkout, and once as a complete compliant walkthrough.
func TestGeneratedBriefFollowsTheReceiptContract(t *testing.T) {
	f := newFixture(t)
	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	capture := filepath.Join(f.root, "generated-brief.md")
	fakeCLI := filepath.Join(f.root, "fake-tskflwctl")
	writeFile(t, fakeCLI, `#!/usr/bin/env bash
case " $* " in
*" audit path "*) printf '%s\n' "$FAKE_AUDIT_PATH"; exit 0 ;;
*" audit lint "*) exit 0 ;;
esac
while (($# > 0)); do
	if [[ "$1" == --body-file ]]; then cp "$2" "$CAPTURE"; fi
	shift
done
`, 0o755)
	brief := filepath.Join(f.root, "brief.md")
	writeFile(t, brief, strings.Join([]string{
		"## Review brief", "Require a consumer inventory and one mutation probe.",
		"## Review target", "## Intended contract to challenge", "## Mandatory evidence floor",
		"## Required hostile angles", "## Validation and restoration", "## Deliverable", "## Reviewer report", "",
	}, "\n\n"), 0o644)

	command := exec.Command("/usr/bin/env", "bash", filepath.Join(repositoryRoot, "scripts", "prepare-adversarial-review-audits.sh"),
		"--area", "receipt-contract", "--title", "Receipt contract", "--brief-file", brief,
		"--reviewer", "claude", "--date", "2026-10-12", "-C", f.root)
	command.Dir = f.root
	command.Env = append(append([]string{}, f.env...), "TSKFLWCTL="+fakeCLI, "CAPTURE="+capture,
		"FAKE_AUDIT_PATH="+filepath.Join(f.source, deliverableRel))
	handoffBytes, err := command.CombinedOutput()
	handoff := string(handoffBytes)
	if err != nil {
		t.Fatalf("generator: %v\n%s", err, handoff)
	}
	for _, want := range []string{
		"SOURCE_ROOT=" + f.source + "\n", "AUDIT_REL=" + deliverableRel + "\n",
		"transfer receipt in your handoff message", "[delivery]", "source checkout's `scripts/isolated-review-workspace.sh receipt",
	} {
		if !strings.Contains(handoff, want) {
			t.Errorf("handoff omitted %q:\n%s", want, handoff)
		}
	}
	generated := readFile(t, capture)
	if strings.Contains(generated, `"$SANDBOX/scripts/isolated-review-workspace.sh"`) {
		t.Error("the brief runs the reviewer-mutable sandbox copy of the helper")
	}

	var blocks []string
	for _, part := range strings.Split(generated, "```sh\n")[1:] {
		blocks = append(blocks, part[:strings.Index(part, "\n```")])
	}
	if len(blocks) != 3 {
		t.Fatalf("generated brief has %d shell blocks, want create, verify, and transfer:\n%s", len(blocks), generated)
	}
	fill := func(block, auditRel string) string {
		block = strings.ReplaceAll(block, "<SOURCE_ROOT from your handoff prompt>", f.source)
		return strings.ReplaceAll(block, "<AUDIT_REL from your handoff prompt>", auditRel)
	}

	// The historical mistake: the absolute path the prompt also prints.
	failed := fill(blocks[0], filepath.Join(f.source, deliverableRel))
	if output, err := f.runShell(failed); err == nil {
		t.Fatalf("a failing create block exited 0:\n%s", output)
	}
	later := "SANDBOX=\nSOURCE_ROOT=" + f.source + "\nAUDIT_REL=" + deliverableRel + "\n" + blocks[1]
	if output, err := f.runShell(later); err == nil || !strings.Contains(output, "create failed") || strings.Contains(output, "diff --git") {
		t.Fatalf("a later block ran without a sandbox; err=%v output:\n%s", err, output)
	}

	walkthrough := strings.Join([]string{
		"set -e",
		fill(blocks[0], deliverableRel),
		`printf '\nWalkthrough report.\n' >>"$SANDBOX/$AUDIT_REL"`,
		blocks[1], blocks[2],
		`"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" receipt --sandbox "$SANDBOX"`,
	}, "\n")
	if output, err := f.runShell(walkthrough); err != nil || !strings.Contains(output, "transfer=succeeded") ||
		!strings.Contains(output, "delivery=confirmed") {
		t.Fatalf("the literal walkthrough failed: %v\n%s", err, output)
	}

	var calls []string
	for _, line := range strings.Split(generated, "\n") {
		for _, field := range strings.Fields(line) {
			if strings.HasSuffix(strings.Trim(field, "`\"$()"), "isolated-review-workspace.sh") {
				// Any other spelling, such as ./scripts/… after cd "$SANDBOX",
				// runs the reviewer-mutable sandbox copy.
				if !strings.Contains(field, "$SOURCE_ROOT/scripts/isolated-review-workspace.sh") {
					t.Errorf("generated brief runs a helper copy other than the source's: %q", line)
				}
				rest := strings.Fields(line[strings.Index(line, field)+len(field):])
				if len(rest) > 0 {
					calls = append(calls, rest[0])
				}
				break
			}
		}
	}
	if got, want := strings.Join(calls, " "), "create verify verify transfer receipt"; got != want {
		t.Errorf("generated helper calls = %q, want %q", got, want)
	}
	for _, want := range []string{"`verify=passed`", "`transfer=pending`", "receipt in your handoff message"} {
		if !strings.Contains(generated, want) {
			t.Errorf("generated brief omitted %q", want)
		}
	}
	if strings.Contains(generated, "transfer result—in the report") {
		t.Error("generated brief still requires the report to record its own transfer")
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
