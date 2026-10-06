package cli

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/appwiring"
	"github.com/andy-esch/taskflow/internal/cli/ports"
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// Integration tests select the same local composition as the binary. These
// test-only helpers are not a production default or a fallback inside the CLI.
func newTestRootCmd(in io.Reader, out, errOut io.Writer) *cobra.Command {
	return NewRootCmd(in, out, errOut, appwiring.LocalBindings())
}

func newTestRootCmdWithApp(in io.Reader, out, errOut io.Writer) (*cobra.Command, *App) {
	return newRootCmd(in, out, errOut, appwiring.LocalBindings())
}

func newTestChromeTheme(args []string) design.Theme {
	return ChromeTheme(args, appwiring.LocalBindings())
}

func TestCommandTreeConstructionDoesNotReadInvocationData(t *testing.T) {
	reads, compositions := 0, 0
	bindings := ports.Bindings{
		Compose: func(authorize func() error) (ports.Services, error) {
			compositions++
			if err := authorize(); err == nil {
				t.Fatal("new command tree authorized a mutation before binding safety")
			}
			return ports.Services{OpenPlanning: func(string) (ports.Planning, error) {
				reads++
				return ports.Planning{}, errors.New("must not open during construction")
			}}, nil
		},
		ReadUser: func() (core.UserConfiguration, error) {
			reads++
			return core.UserConfiguration{}, nil
		},
		ReadRepository: func(string) (core.RepositoryConfiguration, error) {
			reads++
			return core.RepositoryConfiguration{}, nil
		},
	}
	root := NewRootCmd(strings.NewReader(""), io.Discard, io.Discard, bindings)
	if _, err := commandSafetySurface(root); err != nil || reads != 0 || compositions != 1 {
		t.Fatalf("tree construction reads=%d compositions=%d err=%v", reads, compositions, err)
	}
}

func TestMetadataOnlyCommandTreeAndMissingBindingsHaveNoLocalFallback(t *testing.T) {
	t.Chdir(setupRepo(t)) // a real cwd fallback would have records to find
	var stdout, stderr bytes.Buffer
	root := NewRootCmd(strings.NewReader(""), &stdout, &stderr, ports.Bindings{})
	root.SetArgs([]string{"schema", "--json"})
	if err := root.Execute(); err != nil || stdout.Len() == 0 {
		t.Fatalf("metadata-only schema failed: %v stdout=%q", err, stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	root = NewRootCmd(strings.NewReader(""), &stdout, &stderr, ports.Bindings{})
	root.SetArgs([]string{"task", "show", "alpha"})
	if err := root.Execute(); !errors.Is(err, domain.ErrValidation) || stdout.Len() != 0 {
		t.Fatalf("missing composition used cwd: err=%v stdout=%q", err, stdout.String())
	}
	stdout.Reset()
	root = NewRootCmd(strings.NewReader(""), &stdout, &stderr, ports.Bindings{})
	root.SetArgs([]string{"__complete", "task", "show", ""})
	if err := root.Execute(); err != nil || stdout.String() != ":4\n" {
		t.Fatalf("missing completion composition used cwd: err=%v stdout=%q", err, stdout.String())
	}
}

func TestInjectedPlanningOpenerRunsAfterCompletionFlagsAndSafety(t *testing.T) {
	fake := &cliPlanningStub{witness: core.NewSourceSetID()}
	svc := core.MustNewService(nil, core.WithCompletionSource(fake))
	var stdout, stderr bytes.Buffer
	var starts []string
	bindings := ports.Bindings{Compose: func(authorize func() error) (ports.Services, error) {
		return ports.Services{OpenPlanning: func(start string) (ports.Planning, error) {
			starts = append(starts, start)
			if err := authorize(); err == nil || !strings.Contains(err.Error(), "read-only command") {
				t.Fatalf("completion opened before read-only safety binding: %v", err)
			}
			return ports.Planning{
				Repository: core.RepositoryConfiguration{PlanningRoot: "opaque:corpus", ThemeName: "neon"},
				Service:    svc,
			}, nil
		}}, nil
	}}
	root, app := newRootCmd(strings.NewReader(""), &stdout, &stderr, bindings)
	root.SetArgs([]string{"__complete", "-C", "opaque:entry", "task", "show", "por"})
	if err := root.Execute(); err != nil || stdout.String() != "portable\n:4\n" || len(starts) != 1 || starts[0] != "opaque:entry" {
		t.Fatalf("completion err=%v stdout=%q starts=%v", err, stdout.String(), starts)
	}
	if app.Svc != svc || app.Cfg.PlanningRoot != "opaque:corpus" || fake.completionCalls != 1 {
		t.Fatal("CLI substituted another planning source for injected composition")
	}
}

func TestMissingNamedServicesFailWithoutPersistenceFallback(t *testing.T) {
	t.Chdir(setupRepo(t))
	bindings := ports.Bindings{Compose: func(func() error) (ports.Services, error) {
		return ports.Services{OpenPlanning: func(string) (ports.Planning, error) {
			return ports.Planning{
				Repository: core.RepositoryConfiguration{PlanningRoot: "opaque:corpus"},
				Service:    core.MustNewService(nil),
			}, nil
		}}, nil
	}}
	for _, args := range [][]string{
		{"config"}, {"config", "show"}, {"config", "migrate"}, {"config", "doctor"},
		{"doctor"}, {"space", "list"}, {"space", "add"}, {"space", "forget", "probe"},
		{"status", "--all"},
	} {
		t.Run(strings.Join(args, "/"), func(t *testing.T) {
			root := NewRootCmd(strings.NewReader(""), io.Discard, io.Discard, bindings)
			root.SetArgs(args)
			if err := root.Execute(); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "unavailable from this invocation") {
				t.Fatalf("missing service did not fail explicitly: %v", err)
			}
		})
	}
	_, app := newRootCmd(strings.NewReader(""), io.Discard, io.Discard, ports.Bindings{})
	if _, _, _, err := app.uiStartup(); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("missing atlas services did not fail explicitly: %v", err)
	}
}

func TestFailedPlanningOpenerDoesNotPublishPartialInvocation(t *testing.T) {
	sentinel := errors.New("explicit target failed")
	for _, failed := range []bool{false, true} {
		t.Run(map[bool]string{false: "incomplete", true: "failed"}[failed], func(t *testing.T) {
			bindings := ports.Bindings{Compose: func(func() error) (ports.Services, error) {
				return ports.Services{OpenPlanning: func(string) (ports.Planning, error) {
					partial := ports.Planning{Repository: core.RepositoryConfiguration{PlanningRoot: "partial"}}
					if failed {
						return partial, sentinel
					}
					return partial, nil
				}}, nil
			}}
			_, app := newRootCmd(strings.NewReader(""), io.Discard, io.Discard, bindings)
			err := app.resolveFrom("opaque:entry")
			if err == nil || (failed && !errors.Is(err, sentinel)) || app.Cfg != nil || app.Svc != nil || app.Layout != nil {
				t.Fatalf("partial invocation published: err=%v cfg=%v svc=%v layout=%v", err, app.Cfg, app.Svc, app.Layout)
			}
		})
	}
}

func TestFailedPlanningOpenerCommandsDoNotFallbackToPopulatedCWD(t *testing.T) {
	t.Setenv("TSKFLW_SPACE", "")
	t.Chdir(setupRepo(t))
	var control bytes.Buffer
	root := newTestRootCmd(strings.NewReader(""), &control, io.Discard)
	root.SetArgs([]string{"__complete", "task", "show", ""})
	if err := root.Execute(); err != nil || !strings.Contains(control.String(), "alpha\n") || !strings.Contains(control.String(), "beta\n") {
		t.Fatalf("fallback fixture has no real candidates: stdout=%q err=%v", control.String(), err)
	}
	sentinel := errors.New("explicit opener failed")
	for _, complete := range []bool{false, true} {
		t.Run(map[bool]string{false: "show", true: "completion"}[complete], func(t *testing.T) {
			calls := 0
			bindings := ports.Bindings{Compose: func(func() error) (ports.Services, error) {
				return ports.Services{OpenPlanning: func(string) (ports.Planning, error) {
					calls++
					return ports.Planning{}, sentinel
				}}, nil
			}}
			var stdout, stderr bytes.Buffer
			root, app := newRootCmd(strings.NewReader(""), &stdout, &stderr, bindings)
			args := []string{"-C", "opaque:broken", "task", "show", "alpha"}
			if complete {
				args = []string{"__complete", "-C", "opaque:broken", "task", "show", ""}
			}
			root.SetArgs(args)
			err := root.Execute()
			if complete {
				if err != nil || stdout.String() != ":4\n" || strings.Contains(stderr.String(), sentinel.Error()) {
					t.Fatalf("failed completion fell back or leaked error: stdout=%q stderr=%q err=%v", stdout.String(), stderr.String(), err)
				}
			} else if !errors.Is(err, sentinel) || stdout.Len() != 0 {
				t.Fatalf("failed command fell back: stdout=%q err=%v", stdout.String(), err)
			}
			if calls != 1 || app.Cfg != nil || app.Svc != nil || app.Layout != nil {
				t.Fatalf("failed opening published/substituted capabilities: calls=%d cfg=%v svc=%v layout=%v", calls, app.Cfg, app.Svc, app.Layout)
			}
		})
	}
}

func TestLocalCompositionKeepsAuthorizationInvocationScoped(t *testing.T) {
	repo := setupRepo(t)
	bindings := appwiring.LocalBindings()
	read, readApp := newRootCmd(strings.NewReader(""), io.Discard, io.Discard, bindings)
	write, writeApp := newRootCmd(strings.NewReader(""), io.Discard, io.Discard, bindings)
	read.SetArgs([]string{"-C", repo, "task", "show", "alpha"})
	if err := read.Execute(); err != nil {
		t.Fatal(err)
	}
	write.SetArgs([]string{"-C", repo, "task", "set", "alpha", "--priority", "low"})
	if err := write.Execute(); err != nil {
		t.Fatal(err)
	}
	if readApp.Svc == writeApp.Svc || readApp.SpaceSvc == writeApp.SpaceSvc || readApp.ConfigSvc == writeApp.ConfigSvc {
		t.Fatal("command trees shared an application bundle")
	}
	if _, err := readApp.Svc.SetFields("alpha", map[string]any{"priority": "high"}, false, true); err == nil || !strings.Contains(err.Error(), "read-only command") {
		t.Fatalf("another invocation changed the reader's authorization: %v", err)
	}
}

func TestChromeThemeUsesOnlyExplicitPresentationReaders(t *testing.T) {
	t.Setenv("TSKFLW_THEME", "")
	repo := setupRepo(t)
	testutil.Write(t, filepath.Join(repo, config.ConfigFile), "taskflow_root = \".\"\n[theme]\nname = \"catppuccin\"\n")
	t.Chdir(repo)
	homeConfig(t, "[theme]\nname = \"catppuccin\"\n")
	if got := newTestChromeTheme(nil).Name; got != "catppuccin" {
		t.Fatalf("local theme fixture is not discriminating: %q", got)
	}
	repoReads, userReads := 0, 0
	bindings := ports.Bindings{
		ReadRepository: func(string) (core.RepositoryConfiguration, error) {
			repoReads++
			return core.RepositoryConfiguration{ThemeName: "neon"}, nil
		},
		ReadUser: func() (core.UserConfiguration, error) {
			userReads++
			return core.UserConfiguration{ThemeName: "catppuccin"}, nil
		},
	}
	if got := ChromeTheme(nil, bindings).Name; got != "neon" || repoReads != 1 || userReads != 1 {
		t.Fatalf("chrome=%q repository reads=%d user reads=%d", got, repoReads, userReads)
	}
	for _, absent := range []ports.Bindings{
		{},
		{ReadRepository: func(string) (core.RepositoryConfiguration, error) { return core.RepositoryConfiguration{}, nil }},
		{ReadUser: func() (core.UserConfiguration, error) { return core.UserConfiguration{}, nil }},
	} {
		if got := ChromeTheme(nil, absent).Name; got != design.Default().Name {
			t.Fatalf("missing presentation readers discovered a local theme: %q", got)
		}
	}
}
