package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

// newWorkspaceCmd is the cheap read that answers "which planning tree would a
// mutation hit from here?" — the pre-flight companion to --expect-root
// (audit 2026-07-24-ai-agent-cli-ergonomics, H1).
func newWorkspaceCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "workspace",
		Short: "Print the planning tree this directory resolves to",
		Long: "Print the planning tree a command run from here would read and write.\n\n" +
			"External-planning routing is deliberately transparent: a repo whose\n" +
			".tskflwctl.toml carries a planning_repo resolves into ANOTHER repo, so the\n" +
			"directory you are standing in is not necessarily the tree you would change.\n" +
			"This reports the resolved root, the config that selected it, and which\n" +
			"mechanism won — cheaply, before a mutation rather than after.",
		Example:     "  tskflwctl workspace\n  tskflwctl workspace --json\n  tskflwctl task complete foo --expect-root \"$(tskflwctl workspace --json | jq -r .workspace.planning_root)\"",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			ws := app.workspace()
			if app.JSON {
				return render.WorkspaceJSON(app.Out, ws)
			}
			render.WorkspaceHuman(app.Out, app.Style, ws)
			return nil
		},
	}
}

// workspace describes the resolved planning tree. Paths are absolute and
// symlink-resolved so two spellings of the same tree compare equal — the same
// physical-path discipline CheckLinks uses.
func (a *App) workspace() wire.WorkspaceJSON {
	if a.Cfg == nil {
		return wire.WorkspaceJSON{}
	}
	ws := wire.WorkspaceJSON{
		PlanningRoot: physicalPath(a.Cfg.Root),
		Source:       wire.WorkspaceSourceDiscovered,
	}
	if a.Cfg.Dir != "" {
		ws.ConfigPath = physicalPath(filepath.Join(a.Cfg.Dir, config.ConfigFile))
		ws.Source = wire.WorkspaceSourceConfig
		if a.Cfg.PlanningRepo != "" {
			// The case worth noticing: the cwd is NOT the planning tree.
			ws.Source = wire.WorkspaceSourcePointer
		}
	}
	return ws
}

// checkExpectedRoot enforces --expect-root BEFORE anything is written: if the
// resolved planning tree is not the one the caller asserted, fail with ErrConflict
// (exit 14, the same code a CAS clash uses — "the world is not what you assumed").
//
// Comparison is on PHYSICAL paths, so a relative spelling, an absolute one, and a
// symlinked checkout of the same tree all match; only a genuinely different tree
// fails. Empty means the caller made no assertion and nothing is checked.
func (a *App) checkExpectedRoot() error {
	if a.ExpectRoot == "" {
		return nil
	}
	if a.Cfg == nil {
		return fmt.Errorf("%w: --expect-root given but no planning repo resolved", domain.ErrConflict)
	}
	want, got := physicalPath(a.ExpectRoot), physicalPath(a.Cfg.Root)
	if want != got {
		return fmt.Errorf(
			"%w: --expect-root %s but this directory resolves to %s — refusing to touch the wrong planning tree",
			domain.ErrConflict, want, got)
	}
	return nil
}

// physicalPath resolves p to an absolute, symlink-free path so two spellings of the
// same directory compare equal. A path that does not exist yet degrades to its
// lexical absolute form rather than erroring — the comparison is still meaningful,
// and a bogus --expect-root should fail as a MISMATCH (with both paths shown), not as
// an opaque stat error.
func physicalPath(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	return abs
}
