package cli

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/configui"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

func newConfigCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect, migrate, diagnose, and edit configuration",
		Long: "Inspect and maintain repository and user configuration. Bare `config` is\n" +
			"the deterministic alias for `config show`; it never changes behavior based on\n" +
			"whether stdin is a terminal.",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runConfigShow(app)
		},
	}
	cmd.AddCommand(newConfigShowCmd(app))
	cmd.AddCommand(newConfigMigrateCmd(app))
	cmd.AddCommand(newConfigDoctorCmd(app))
	cmd.AddCommand(newConfigEditCmd(app))
	return cmd
}

func newConfigEditCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit safe user or repository preferences interactively",
		Long: "Open a typed configuration editor for theme and pager preferences. User\n" +
			"scope is the default; repository overrides must be selected explicitly. Planning\n" +
			"identity, topology, and the space registry are shown read-only.",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "mutating"},
		RunE: func(_ *cobra.Command, _ []string) error {
			if app.DryRun {
				return fmt.Errorf("%w: `config edit` has no --dry-run preview (it's interactive) — use `config show` to inspect values", domain.ErrValidation)
			}
			if !app.Gate.On() || !isTerminal(app.Out) {
				return fmt.Errorf("%w: `config edit` needs an interactive terminal — use `config show`, `config migrate`, or edit the documented typed TOML fields", domain.ErrValidation)
			}
			start, err := app.startDir()
			if err != nil {
				return err
			}
			// The gate guarantees a terminal for the real command. Keep injected test
			// streams usable by falling back to the dark variant when they are not
			// concrete terminal files.
			dark := true
			if in, inOK := app.In.(*os.File); inOK {
				if out, outOK := app.Out.(*os.File); outOK {
					dark = lipgloss.HasDarkBackground(in, out)
				}
			}
			return configui.Run(app.ConfigSvc, start, app.configurationOverrides(), dark, app.In, app.Out)
		},
	}
}

func newConfigShowCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show repository, user, and effective configuration with provenance",
		Long: "Show raw repository and user scopes separately, followed by each effective\n" +
			"theme and pager value and the source that won precedence. Also reports planning\n" +
			"topology, durable identity, config paths, tracked repos, and pending migrations.",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runConfigShow(app)
		},
	}
}

func runConfigShow(app *App) error {
	start, err := app.startDir()
	if err != nil {
		return err
	}
	snapshot, err := app.ConfigSvc.Snapshot(start, app.configurationOverrides())
	if err != nil {
		return err
	}
	if app.JSON {
		return render.ConfigJSON(app.Out, snapshot)
	}
	render.ConfigHuman(app.Out, app.Style, snapshot)
	return nil
}

func newConfigMigrateCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Apply safe, idempotent configuration upgrades",
		Long: "Apply safe configuration upgrades such as minting a missing planning-repo id\n" +
			"or recording planning_repo_id on a legacy external pointer. The edit is atomic\n" +
			"and preserves comments, ordering, relative path spelling, and unknown keys.",
		Example:     "  tskflwctl config migrate --dry-run\n  tskflwctl config migrate --json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "mutating"},
		RunE: func(_ *cobra.Command, _ []string) error {
			start, err := app.startDir()
			if err != nil {
				return err
			}
			migration, err := app.ConfigSvc.Migrate(start, app.DryRun)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.ConfigMigrationJSON(app.Out, migration, workspaceAfterMigration(app, migration))
			}
			render.ConfigMigrationHuman(app.Out, app.Style, migration)
			return nil
		},
	}
}

func workspaceAfterMigration(app *App, migration core.ConfigurationMigration) wire.WorkspaceJSON {
	workspace := app.workspace()
	for _, step := range migration.Steps {
		if step.Kind == core.ConfigurationMigrationRepoID {
			workspace.RepoID = step.Value
		}
	}
	return workspace
}

func (a *App) configurationOverrides() core.ConfigurationOverrides {
	var pagerEnabled *bool
	if a.NoPager {
		value := false
		pagerEnabled = &value
	} else if a.Paginate {
		value := true
		pagerEnabled = &value
	}
	return core.ConfigurationOverrides{
		ThemeFlag:          a.Theme,
		ThemeEnvironment:   os.Getenv("TSKFLW_THEME"),
		PagerEnabledFlag:   pagerEnabled,
		PagerEnvironment:   os.Getenv("TSKFLW_PAGER"),
		GenericPagerEnv:    os.Getenv("PAGER"),
		DefaultTheme:       design.Default().Name,
		KnownThemes:        design.Names(),
		DefaultPagerEnable: true,
		DefaultPager:       "less -FRX",
	}
}
