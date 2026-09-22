package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/userconfig"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestCommandSafetySurfaceCoversEveryRunnableCommand(t *testing.T) {
	var output bytes.Buffer
	root := NewRootCmd(strings.NewReader(""), &output, &output)
	commands, err := commandSafetySurface(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(commands) == 0 {
		t.Fatal("command safety surface is empty")
	}
	paths := make([]string, len(commands))
	for i, command := range commands {
		paths[i] = command.Path
		if !recognizedCommandSafety(command.Safety) {
			t.Errorf("%s safety = %q", command.Path, command.Safety)
		}
	}
	if !slices.IsSorted(paths) {
		t.Errorf("command paths are not byte-sorted: %v", paths)
	}
	for _, want := range []string{
		"tskflwctl help",            // framework-owned runnable command
		"tskflwctl completion bash", // framework-owned generated subtree
		"tskflwctl list",            // hidden bare-verb redirect
		"tskflwctl task promote",    // hidden deprecated compatibility leaf
		"tskflwctl task list",       // ordinary read
		"tskflwctl task set",        // ordinary mutation
	} {
		if !slices.Contains(paths, want) {
			t.Errorf("command safety surface is missing %q", want)
		}
	}

	// Cobra registers its hidden completion transport only while servicing one.
	// Execute it once, then prove that the runtime binding annotated the newly
	// registered runnable command and the same exhaustive walk still succeeds.
	root.SetArgs([]string{cobra.ShellCompRequestCmd, "version", ""})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute dynamic completion transport: %v", err)
	}
	commands, err = commandSafetySurface(root)
	if err != nil {
		t.Fatalf("command surface after dynamic completion registration: %v", err)
	}
	if !slices.ContainsFunc(commands, func(command wire.SchemaCommand) bool {
		return command.Path == "tskflwctl "+cobra.ShellCompRequestCmd &&
			command.Safety == commandSafetyReadOnly && command.Hidden
	}) {
		t.Error("dynamic completion transport was not classified read-only and hidden")
	}
}

func TestReadOnlyCommandCannotReachAnyMutatingApplicationBoundary(t *testing.T) {
	mutations := []struct {
		name string
		run  func(*App, string, bool) error
	}{
		{"planning-store", func(app *App, _ string, dryRun bool) error {
			_, err := app.Svc.SetFields("alpha", map[string]any{"priority": "high"}, false, dryRun)
			return err
		}},
		{"task-lifecycle", func(app *App, _ string, dryRun bool) error {
			_, err := app.Svc.Move("alpha", domain.StatusInProgress, dryRun, core.TaskLifecycleOverrideNone)
			return err
		}},
		{"configuration", func(app *App, repo string, dryRun bool) error {
			_, err := app.ConfigSvc.SetPreference(repo, core.PreferenceChange{
				Scope: core.ConfigScopeUser, Field: core.PreferencePagerCommand, Value: "less",
			}, dryRun)
			return err
		}},
		{"space-registry", func(app *App, repo string, dryRun bool) error {
			_, err := app.SpaceSvc.Add(repo, "probe", dryRun)
			return err
		}},
		{"opened-workspace", func(app *App, repo string, dryRun bool) error {
			workspace, err := app.WorkspaceSvc.Open(core.WorkspaceRequest{Start: repo})
			if err != nil {
				return err
			}
			_, err = workspace.Planning.SetFields("alpha", map[string]any{"priority": "high"}, false, dryRun)
			return err
		}},
	}

	for _, mutation := range mutations {
		for _, dryRun := range []bool{true, false} {
			mode := map[bool]string{true: "dry-run", false: "write"}[dryRun]
			t.Run(mutation.name+"/"+mode, func(t *testing.T) {
				repo := setupRepo(t)
				home := t.TempDir()
				t.Setenv(userconfig.DirEnv, home)
				before, err := os.ReadFile(alphaPath(repo))
				if err != nil {
					t.Fatal(err)
				}
				var out bytes.Buffer
				root, app := newRootCmd(strings.NewReader(""), &out, &out)
				root.AddCommand(&cobra.Command{
					Use:         "safety-probe",
					Hidden:      true,
					Args:        cobra.NoArgs,
					Annotations: map[string]string{commandSafetyAnnotation: commandSafetyReadOnly},
					RunE: func(_ *cobra.Command, _ []string) error {
						return mutation.run(app, repo, dryRun)
					},
				})
				root.SetArgs([]string{"-C", repo, "safety-probe"})
				err = root.Execute()
				if err == nil || !strings.Contains(err.Error(), `read-only command "tskflwctl safety-probe"`) {
					t.Fatalf("read-only mutation error = %v, want bound read-only command violation", err)
				}
				after, readErr := os.ReadFile(alphaPath(repo))
				if readErr != nil {
					t.Fatal(readErr)
				}
				if !bytes.Equal(after, before) {
					t.Fatal("read-only command changed the task before the safety guard failed")
				}
				entries, readErr := os.ReadDir(home)
				if readErr != nil {
					t.Fatal(readErr)
				}
				if len(entries) != 0 {
					t.Fatalf("read-only command changed home configuration: %v", entries)
				}
			})
		}
	}
}

func TestCommandSafetyFailsClosedBeforeBinding(t *testing.T) {
	var state commandSafetyState
	err := state.authorizeMutation()
	if err == nil || !strings.Contains(err.Error(), "before a command safety classification was bound") {
		t.Fatalf("unbound authorization error = %v", err)
	}
}

func TestDirectWriteCommandsRejectReadOnlyClassification(t *testing.T) {
	t.Run("init", func(t *testing.T) {
		t.Setenv(userconfig.DirEnv, t.TempDir())
		target := filepath.Join(t.TempDir(), "new-repo")
		var out bytes.Buffer
		root := NewRootCmd(strings.NewReader(""), &out, &out)
		cmd, _, err := root.Find([]string{"init"})
		if err != nil {
			t.Fatal(err)
		}
		cmd.Annotations[commandSafetyAnnotation] = commandSafetyReadOnly
		root.SetArgs([]string{"init", "--path", target, "--no-register"})
		err = root.Execute()
		if err == nil || !strings.Contains(err.Error(), `read-only command "tskflwctl init"`) {
			t.Fatalf("init error = %v, want bound read-only command violation", err)
		}
		if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
			t.Fatalf("denied init touched target: %v", statErr)
		}
	})

	for _, dryRun := range []bool{true, false} {
		mode := map[bool]string{true: "dry-run", false: "write"}[dryRun]
		t.Run("thread-compose/"+mode, func(t *testing.T) {
			repo := freshRepo(t)
			taskID := writeThreadApplyCLITask(t, repo, "safety-member", domain.StatusNextUp)
			manifestPath := filepath.Join(repo, "thread.yml")
			planPath := filepath.Join(repo, "thread-plan.yml")
			mustWrite(t, manifestPath, "thread:\n  title: Safety probe\n  description: Verify command authorization\n  goal: Reject read-only composition\nnodes:\n  - key: member\n    task_id: "+taskID+"\n")

			var out bytes.Buffer
			root := NewRootCmd(strings.NewReader(""), &out, &out)
			cmd, _, err := root.Find([]string{"thread", "compose"})
			if err != nil {
				t.Fatal(err)
			}
			cmd.Annotations[commandSafetyAnnotation] = commandSafetyReadOnly
			args := []string{"-C", repo, "thread", "compose", "--from", manifestPath, "--out", planPath}
			if dryRun {
				args = append(args, "--dry-run")
			}
			root.SetArgs(args)
			err = root.Execute()
			if err == nil || !strings.Contains(err.Error(), `read-only command "tskflwctl thread compose"`) {
				t.Fatalf("compose error = %v, want bound read-only command violation", err)
			}
			if _, statErr := os.Stat(planPath); !os.IsNotExist(statErr) {
				t.Fatalf("denied compose wrote plan: %v", statErr)
			}
		})
	}
}

func TestCustomPreRunsBindSelectedCommandSafety(t *testing.T) {
	repo := setupRepo(t)
	t.Setenv(userconfig.DirEnv, t.TempDir())
	for _, tc := range []struct {
		name string
		args []string
		path string
	}{
		{"doctor", []string{"-C", repo, "doctor"}, "tskflwctl doctor"},
		{"status-all", []string{"-C", repo, "status", "--all"}, "tskflwctl status"},
		{"template", []string{"-C", repo, "template", "list"}, "tskflwctl template list"},
		{"theme", []string{"-C", repo, "theme", "list"}, "tskflwctl theme list"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			root, app := newRootCmd(strings.NewReader(""), &out, &out)
			root.SetArgs(tc.args)
			_ = root.Execute() // command outcome is unrelated; its pre-run must have bound.
			if app.commandSafety.path != tc.path || !recognizedCommandSafety(app.commandSafety.safety) {
				t.Fatalf("bound safety = %+v, want recognized capability for %q", app.commandSafety, tc.path)
			}
		})
	}
}
