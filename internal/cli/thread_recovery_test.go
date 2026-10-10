package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/userconfig"
)

// Parse the emitted POSIX command with a real shell, then execute its argv using
// the production CLI harness. No installed binary or fixture cwd can make an
// incorrect hint pass. All commands and paths originate in these fixtures.
func runRecoveryHint(t *testing.T, diagnostic, repo, tree string, marker bool) {
	t.Helper()
	commands := regexp.MustCompile("`(tskflwctl [^`]+)`").FindAllStringSubmatch(diagnostic, -1)
	for _, match := range commands {
		command := match[1]
		if marker != strings.Contains(command, "config migrate") {
			continue
		}
		quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }
		command = strings.ReplaceAll(command, "<planning-repo>", quote(repo))
		command = strings.ReplaceAll(command, "<existing-tree>", quote(tree))
		shell, err := exec.LookPath("sh")
		if err != nil {
			t.Skip("POSIX shell unavailable for recovery command replay")
		}
		parsed, err := exec.Command(shell, "-c", "set -- "+command+"; printf '%s\\0' \"$@\"").Output()
		if err != nil {
			t.Fatalf("parse emitted hint %q: %v", command, err)
		}
		args := strings.Split(strings.TrimSuffix(string(parsed), "\x00"), "\x00")
		if len(args) < 2 || args[0] != "tskflwctl" {
			t.Fatalf("invalid command argv: %q", args)
		}
		runRoot(t, args[1:]...)
		return
	}
	t.Fatalf("no executable recovery hint for marker=%v in %s", marker, diagnostic)
}

func recoveryFixture(t *testing.T, tree string, marker bool) (repo, data, manifest, oldPlan string) {
	t.Helper()
	repo = filepath.Join(t.TempDir(), "planning repo's space")
	if _, err := config.Init(repo, tree, false); err != nil {
		t.Fatal(err)
	}
	data = filepath.Join(repo, tree)
	data, err := filepath.EvalSymlinks(data)
	if err != nil {
		t.Fatal(err)
	}
	id := writeThreadApplyCLITask(t, data, "recovery-member", domain.StatusNextUp)
	manifest = filepath.Join(repo, "manifest.yml")
	oldPlan = filepath.Join(repo, "old-plan.yml")
	mustWrite(t, manifest, "thread:\n  title: Recovery\n  description: Preserve existing documents\n  goal: Recover explicitly\nnodes:\n  - key: member\n    task_id: "+id+"\n")
	runRoot(t, "-C", repo, "thread", "compose", "--from", manifest, "--out", oldPlan)
	path := filepath.Join(repo, config.ConfigFile)
	if marker {
		// Unknown keys and comments must survive identity migration byte-for-byte.
		mustWrite(t, path, fmt.Sprintf("# legacy marker\ntaskflow_root = %q\nunknown = \"preserve me\"\n", tree))
	} else if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	return repo, data, manifest, oldPlan
}

func recoveryDocuments(t *testing.T, data string) map[string]map[string]testutil.TreeEntry {
	t.Helper()
	result := make(map[string]map[string]testutil.TreeEntry)
	for _, dir := range []string{domain.TasksDir, domain.EpicsDir, domain.AuditsDir, domain.ResearchDir, domain.ThreadsDir} {
		result[dir] = testutil.SnapshotTree(t, filepath.Join(data, dir))
	}
	return result
}

func TestThreadRecoveryHintsExecuteWithoutForkingExistingDocuments(t *testing.T) {
	for _, tree := range []string{".", "planning"} {
		for _, marker := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/marker=%v", tree, marker), func(t *testing.T) {
				home := t.TempDir()
				t.Setenv(userconfig.DirEnv, home)
				repo, data, manifest, oldPlan := recoveryFixture(t, tree, marker)
				t.Chdir(t.TempDir())
				t.Setenv("TSKFLW_SPACE", "unrelated-missing-space")
				before, beforeHome := recoveryDocuments(t, data), testutil.SnapshotTree(t, home)
				for _, args := range [][]string{
					{"-C", repo, "thread", "compose", "--from", manifest, "--out", filepath.Join(repo, "new-plan.yml")},
					{"-C", repo, "thread", "apply", oldPlan},
				} {
					_, err := runRootRC(t, args...)
					if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "no durable id") ||
						!strings.Contains(err.Error(), "recompose") || !strings.Contains(err.Error(), "--taskflow-root") {
						t.Fatalf("refusal did not explain recovery: %v", err)
					}
					if !reflect.DeepEqual(before, recoveryDocuments(t, data)) {
						t.Fatal("refusal wrote planning documents")
					}
				}
				_, err := runRootRC(t, "-C", repo, "thread", "apply", oldPlan)
				runRecoveryHint(t, err.Error(), repo, tree, marker)
				if !reflect.DeepEqual(before, recoveryDocuments(t, data)) || !reflect.DeepEqual(beforeHome, testutil.SnapshotTree(t, home)) {
					t.Fatal("recovery changed existing documents or registered a local space")
				}
				if tree != "." {
					if _, err := os.Stat(filepath.Join(repo, domain.TasksDir)); !os.IsNotExist(err) {
						t.Fatalf("nested recovery forked a flat tree: %v", err)
					}
				}
				cfg, err := config.Discover(repo)
				if err != nil || cfg.ID == "" || cfg.Root != data {
					t.Fatalf("recovered identity/root = %+v, %v", cfg, err)
				}
				if marker {
					b, err := os.ReadFile(filepath.Join(repo, config.ConfigFile))
					if err != nil || !strings.Contains(string(b), "# legacy marker") || !strings.Contains(string(b), "unknown = \"preserve me\"") {
						t.Fatalf("migration lost unknown config: %s, %v", b, err)
					}
				}
				if _, err := runRootRC(t, "-C", repo, "thread", "apply", oldPlan); !errors.Is(err, domain.ErrConflict) {
					t.Fatalf("old plan was adopted after regenerating identity: %v", err)
				}
				newPlan := filepath.Join(repo, "recomposed.yml")
				runRoot(t, "-C", repo, "thread", "compose", "--from", manifest, "--out", newPlan)
				runRoot(t, "-C", repo, "thread", "apply", newPlan)
			})
		}
	}
}

func TestPointerRecoveryHintsDoNotSilentlyRebind(t *testing.T) {
	for _, tree := range []string{".", "planning"} {
		for _, marker := range []bool{false, true} {
			for _, bound := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/marker=%v/bound=%v", tree, marker, bound), func(t *testing.T) {
					home := t.TempDir()
					t.Setenv(userconfig.DirEnv, home)
					repo, data, manifest, _ := recoveryFixture(t, tree, marker)
					impl := t.TempDir()
					pointer := fmt.Sprintf("# preserve pointer\nplanning_repo = %q\n", filepath.ToSlash(repo))
					if bound {
						pointer += "planning_repo_id = \"original-identity\"\n"
					}
					pointerPath := filepath.Join(impl, config.ConfigFile)
					mustWrite(t, pointerPath, pointer)
					t.Chdir(t.TempDir())
					t.Setenv("TSKFLW_SPACE", "unrelated-missing-space")
					before, beforeHome := recoveryDocuments(t, data), testutil.SnapshotTree(t, home)
					args := []string{"-C", impl, "config", "migrate"}
					want := domain.ErrValidation
					if bound {
						args = []string{"-C", impl, "thread", "compose", "--from", manifest, "--out", filepath.Join(repo, "pointer-plan.yml")}
						want = domain.ErrConflict
					}
					_, err := runRootRC(t, args...)
					if !errors.Is(err, want) || !strings.Contains(err.Error(), "identity") {
						t.Fatalf("pointer refusal = %v; want %v and recovery guidance", err, want)
					}
					runRecoveryHint(t, err.Error(), repo, tree, marker)
					cfg, discoverErr := config.Discover(repo)
					if discoverErr != nil || cfg.Root != data || cfg.ID == "" {
						t.Fatalf("pointer-target recovery selected a different tree: %+v, %v", cfg, discoverErr)
					}
					b, err := os.ReadFile(pointerPath)
					if err != nil || string(b) != pointer || !reflect.DeepEqual(before, recoveryDocuments(t, data)) ||
						!reflect.DeepEqual(beforeHome, testutil.SnapshotTree(t, home)) {
						t.Fatal("target recovery changed documents, pointer binding, or local registration")
					}
					if bound {
						if _, err := runRootRC(t, args...); !errors.Is(err, domain.ErrConflict) {
							t.Fatalf("opted-in pointer silently adopted the new target identity: %v", err)
						}
					} else {
						runRoot(t, args...)
						b, err := os.ReadFile(pointerPath)
						if err != nil || !strings.Contains(string(b), "planning_repo_id") || !strings.Contains(string(b), "# preserve pointer") {
							t.Fatalf("deliberate pointer migration = %s, %v", b, err)
						}
					}
				})
			}
		}
	}
}

func TestPointerRecoveryDoesNotInitializeInsideGoverningMarker(t *testing.T) {
	repo, impl := t.TempDir(), t.TempDir()
	result, err := config.Init(repo, "planning", false)
	if err != nil {
		t.Fatal(err)
	}
	pointer := fmt.Sprintf("planning_repo = %q\nplanning_repo_id = %q\n", filepath.ToSlash(filepath.Join(repo, "planning")), result.PlanningID)
	mustWrite(t, filepath.Join(impl, config.ConfigFile), pointer)
	before := testutil.SnapshotTree(t, repo)
	_, err = runRootRC(t, "-C", impl, "thread", "list")
	if !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "governing configuration") || strings.Contains(err.Error(), "tskflwctl init") {
		t.Fatalf("data-directory alias should direct recovery to the existing marker, not mint inside it: %v", err)
	}
	if !reflect.DeepEqual(before, testutil.SnapshotTree(t, repo)) {
		t.Fatal("refusal changed the planning target")
	}
}
