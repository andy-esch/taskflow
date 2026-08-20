package cli

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/wire"
)

// newWorkspaceCmd is the cheap read that answers "which planning tree would a
// mutation hit from here?" — the pre-flight half of audit
// 2026-07-24-ai-agent-cli-ergonomics H1, whose after-the-fact half is the `workspace`
// object on every mutation receipt.
func newWorkspaceCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "workspace",
		Short: "Print the planning tree this directory resolves to",
		Long: "Print the planning tree a command run from here would read and write.\n\n" +
			"External-planning routing is deliberately transparent: a repo whose\n" +
			".tskflwctl.toml carries a planning_repo resolves into ANOTHER repo, so the\n" +
			"directory you are standing in is not necessarily the tree you would change.\n" +
			"This reports the resolved root, the config that selected it, and which\n" +
			"mechanism won — cheaply, before a mutation rather than after.\n\n" +
			"EXPERIMENTAL: this command and the `workspace` object on mutation receipts\n" +
			"may change shape while the multi-space work settles. Pin `schema_version`\n" +
			"if you script against it.",
		Example: "  tskflwctl workspace\n  tskflwctl workspace --json",
		Args:    cobra.NoArgs,
		// stability is declarative today (nothing reads it yet) but it is where a
		// `schema --type cli` surface would pick this up, so it is recorded next to
		// safety rather than living only in prose.
		Annotations: map[string]string{"safety": "read-only", "stability": "experimental"},
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
		RepoID:       a.Cfg.ID,
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

// physicalPath resolves p to an absolute, symlink-free path so two spellings of the
// same directory compare equal. A path that does not exist yet degrades to its
// lexical absolute form rather than erroring: a path we cannot stat is still worth
// reporting as the resolved root, and an opaque stat error here would be less useful
// than the path itself.
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
