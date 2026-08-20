package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/spacehealth"
	"github.com/andy-esch/taskflow/internal/userconfig"
	"github.com/andy-esch/taskflow/internal/wire"
)

// newSpaceCmd is the `space` group: the home-scoped registry of planning repos on this
// machine. Like `version` and `theme`, it works ANYWHERE — the registry lives in the home
// config, which exists independently of any planning repo, so its PreRun sets up styling
// and skips discovery entirely.
func newSpaceCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Manage the registry of planning repos on this machine",
		Long: "The spaces registry records which planning repos exist on this machine, so\n" +
			"they can be addressed by name instead of by path.\n\n" +
			"It is ADVISORY: nothing in it changes what a command run in a directory\n" +
			"resolves to. With no registry, everything behaves exactly as it did before one\n" +
			"existed, and deleting it costs convenience — never data, never addressability.",
		Annotations:       map[string]string{"safety": "read-only"},
		PersistentPreRunE: app.styleOnlyPreRun,
	}
	cmd.AddCommand(newSpaceListCmd(app), newSpaceAddCmd(app), newSpaceForgetCmd(app))
	return cmd
}

func newSpaceListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registered planning repos with their current health",
		Long: "List every registered planning repo and diagnose it without changing the registry.\n\n" +
			"Healthy repos are `ok` or `empty`. Missing paths, non-repos, unreadable configs,\n" +
			"and durable-id mismatches stay listed with a remedy; none is auto-forgotten.",
		Example:     "  tskflwctl space list\n  tskflwctl space list --json",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "read-only"},
		RunE: func(_ *cobra.Command, _ []string) error {
			entries, err := loadSpaceEntries()
			if err != nil {
				return err
			}
			if app.JSON {
				return render.SpacesJSON(app.Out, entries)
			}
			render.SpacesHuman(app.Out, app.Style, entries)
			return nil
		},
	}
}

func newSpaceAddCmd(app *App) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "add [path]",
		Short: "Register a planning repo (defaults to the current directory)",
		Long: "Register a planning repo so it can be addressed by name.\n\n" +
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
	cmd.Flags().StringVar(&id, "id", "", "label to address this space by (default: the directory name)")
	return cmd
}

func newSpaceForgetCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "forget <id>",
		Short: "Drop a space from the registry (the repo itself is untouched)",
		Long: "Remove an entry from the registry.\n\n" +
			"This never touches the repo on disk — forgetting is a registry edit, not a\n" +
			"deletion, so a space can always be re-added with `space add`.",
		Example:           "  tskflwctl space forget old-thing",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: completeSpaceIDs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runSpaceForget(app, args[0])
		},
	}
}

// loadSpaceEntries reads the registry and resolves each entry's current state. A broken
// entry is REPORTED, never dropped and never fatal: the registry describes what you told
// it about, and a missing or mismatched repo is information, not an error.
func loadSpaceEntries() ([]wire.SpaceEntry, error) {
	diagnoses, err := spacehealth.DiagnoseRegistry()
	if err != nil {
		return nil, classifySpaceRegistryError(err)
	}
	out := make([]wire.SpaceEntry, 0, len(diagnoses))
	for _, diagnosis := range diagnoses {
		out = append(out, spaceEntry(diagnosis))
	}
	return out, nil
}

func spaceEntry(diagnosis spacehealth.SpaceProblem) wire.SpaceEntry {
	s := diagnosis.Space
	e := wire.SpaceEntry{
		ID: s.ID, Path: s.Path, VerifyID: s.VerifyID, Label: s.Label, Added: s.Added,
		State: string(diagnosis.Kind), Root: diagnosis.Root,
		Detail: diagnosis.Message, Remedy: diagnosis.Remedy,
	}
	return e
}

func runSpaceAdd(app *App, target, id string) error {
	abs, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	// Validate BEFORE writing: the "require + error, leave nothing behind" contract that
	// `init --planning-repo` already follows.
	cfg, err := config.Discover(abs)
	if err != nil {
		return err
	}
	// Store the repo directory — the place carrying .tskflwctl.toml — rather than the
	// exact subdirectory the user happened to name or the resolved planning root. A pointer
	// repo therefore stays registered as the pointer checkout, and ordinary discovery does
	// the routing when it is used later.
	repoDir := cfg.Dir
	if repoDir == "" {
		repoDir = cfg.Root // config-less bare tasks/ tree: root is the only repo anchor
	}
	if id == "" {
		id = defaultSpaceID(repoDir)
	}
	if err := validateSpaceID(id); err != nil {
		return err
	}
	space := userconfig.Space{
		ID:       id,
		Path:     userconfig.TildePath(repoDir),
		VerifyID: cfg.ID,
	}
	added, existing, err := userconfig.AddSpace(space, app.DryRun)
	if err != nil {
		return classifySpaceRegistryError(err)
	}
	if !added {
		return reportSpaceChange(app, spaceEntry(spacehealth.DiagnoseSpace(existing)), false, "already registered as")
	}
	verb := "registered"
	if app.DryRun {
		verb = "would register"
	}
	return reportSpaceChange(app, spaceEntry(spacehealth.DiagnoseSpace(existing)), true, verb)
}

func runSpaceForget(app *App, id string) error {
	removed, existing, err := userconfig.ForgetSpace(id, app.DryRun)
	if err != nil {
		return classifySpaceRegistryError(err)
	}
	if !removed {
		return fmt.Errorf("%w: no space named %q — `space list` shows the registered ones", domain.ErrNotFound, id)
	}
	verb := "forgot"
	if app.DryRun {
		verb = "would forget"
	}
	return reportSpaceChange(app, spaceEntry(spacehealth.DiagnoseSpace(existing)), true, verb)
}

func classifySpaceRegistryError(err error) error {
	switch {
	case errors.Is(err, userconfig.ErrSpaceIDConflict):
		return fmt.Errorf("%w: %s", domain.ErrConflict, err.Error())
	case errors.Is(err, userconfig.ErrInvalidRegistry):
		return fmt.Errorf("%w: %s", domain.ErrValidation, err.Error())
	default:
		return err // permission/I/O/lock failures are operational, not bad user input
	}
}

// reportSpaceChange emits the mutation receipt in whichever face is active.
func reportSpaceChange(app *App, e wire.SpaceEntry, changed bool, verb string) error {
	if app.JSON {
		return render.SpaceMutationJSON(app.Out, e, changed, app.DryRun)
	}
	render.SpaceMutationHuman(app.Out, app.Style, e, changed, verb)
	return nil
}

// defaultSpaceID derives a label from the directory name — the thing a person would call
// the repo anyway. Collisions are refused at registration (see AddSpace), never silently
// suffixed, so an accidental clash is visible rather than producing `taskflow-2`.
func defaultSpaceID(dir string) string {
	return strings.ToLower(filepath.Base(filepath.Clean(dir)))
}

// validateSpaceID keeps a label usable as a command-line word and a completion candidate.
func validateSpaceID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: a space needs a label; pass --id", domain.ErrValidation)
	}
	if strings.ContainsAny(id, " \t/\\\"'") {
		return fmt.Errorf("%w: space label %q may not contain spaces, quotes or path separators", domain.ErrValidation, id)
	}
	return nil
}
