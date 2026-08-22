package cli

import (
	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
)

// newSpaceCmd is the `space` group: the home-scoped registry of planning repos on this
// machine. Like `version` and `theme`, it works ANYWHERE — the registry lives in the home
// config, which exists independently of any planning repo, so its PreRun sets up styling
// and skips discovery entirely.
func newSpaceCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Manage planning spaces and their registered entry points",
		Long: "The spaces registry records which repo checkouts enter each planning space, so\n" +
			"they can be addressed by name instead of by path. Direct planning checkouts and\n" +
			"implementation pointers with the same durable planning id form one logical space.\n\n" +
			"It is ADVISORY: ordinary cwd discovery never consults it. Select an entry point\n" +
			"explicitly with --space (or TSKFLW_SPACE); without that opt-in, a machine with no\n" +
			"registry behaves exactly as before. Deleting it costs convenience — never data.",
		Annotations:       map[string]string{"safety": "read-only"},
		PersistentPreRunE: app.styleOnlyPreRun,
	}
	cmd.AddCommand(newSpaceListCmd(app), newSpaceAddCmd(app), newSpaceForgetCmd(app))
	return cmd
}

func newSpaceListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List planning spaces, entry points, and current health",
		Long: "List every registered planning entry point and diagnose it without changing the registry.\n\n" +
			"Registered paths that resolve to the same durable planning identity are grouped as\n" +
			"entry points. The first direct checkout anchors a group when one is registered;\n" +
			"indentation means shared planning data, not filesystem ownership.\n\n" +
			"Healthy repos are `ok` or `empty`. Missing paths, non-repos, unreadable configs,\n" +
			"and durable-id mismatches stay listed with a remedy; none is auto-forgotten.",
		Example:     "  tskflwctl space list\n  tskflwctl space list --json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			catalog, err := app.SpaceSvc.Catalog()
			if err != nil {
				return err
			}
			if app.JSON {
				return render.SpacesJSON(app.Out, catalog.Entries)
			}
			render.SpacesHuman(app.Out, app.Style, catalog.Groups)
			return nil
		},
	}
}

func newSpaceAddCmd(app *App) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "add [path]",
		Short: "Register a planning entry point (defaults to the current directory)",
		Long: "Register a direct planning checkout or implementation pointer as an entry point.\n\n" +
			"The path is VALIDATED as a planning repo before anything is written, so a typo\n" +
			"fails with nothing left behind. Registering the same path twice is a no-op —\n" +
			"identity is the physical directory, so relative, absolute and symlinked\n" +
			"spellings of one repo collapse to a single entry.",
		Example:     "  tskflwctl space add\n  tskflwctl space add ~/git/andy-esch/desirelines\n  tskflwctl space add ../other --id other",
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{"safety": "mutating"},
		RunE: func(_ *cobra.Command, args []string) error {
			target, err := app.startDir()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				target = args[0]
			}
			return runSpaceAdd(app, target, id)
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "label to address this entry point by (default: the directory name)")
	return cmd
}

func newSpaceForgetCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "forget <id>",
		Short: "Drop an entry point from the registry (the repo itself is untouched)",
		Long: "Remove an entry from the registry.\n\n" +
			"This never touches the repo on disk — forgetting is a registry edit, not a\n" +
			"deletion, so a space can always be re-added with `space add`.",
		Example:           "  tskflwctl space forget old-thing",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: completeSpaceIDs(app.SpaceSvc),
		RunE: func(_ *cobra.Command, args []string) error {
			return runSpaceForget(app, args[0])
		},
	}
}

func runSpaceAdd(app *App, target, id string) error {
	mutation, err := app.SpaceSvc.Add(target, id, app.DryRun)
	if err != nil {
		return err
	}
	if !mutation.Changed {
		return reportSpaceChange(app, mutation, "already registered as")
	}
	verb := "registered"
	if app.DryRun {
		verb = "would register"
	}
	return reportSpaceChange(app, mutation, verb)
}

func runSpaceForget(app *App, id string) error {
	mutation, err := app.SpaceSvc.Forget(id, app.DryRun)
	if err != nil {
		return err
	}
	verb := "forgot"
	if app.DryRun {
		verb = "would forget"
	}
	return reportSpaceChange(app, mutation, verb)
}

// reportSpaceChange emits the mutation receipt in whichever face is active.
func reportSpaceChange(app *App, mutation core.SpaceMutation, verb string) error {
	if app.JSON {
		return render.SpaceMutationJSON(app.Out, mutation)
	}
	render.SpaceMutationHuman(app.Out, app.Style, mutation, verb)
	return nil
}
