package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/domain"
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
		Use:         "list",
		Short:       "List the registered planning repos and whether each still resolves",
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
			target := "."
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
		Example:     "  tskflwctl space forget old-thing",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"safety": "mutating"},
		RunE: func(_ *cobra.Command, args []string) error {
			return runSpaceForget(app, args[0])
		},
	}
}

// loadSpaceEntries reads the registry and resolves each entry's current state. A broken
// entry is REPORTED, never dropped and never fatal: the registry describes what you told
// it about, and a repo that has moved is information, not an error.
func loadSpaceEntries() ([]wire.SpaceEntry, error) {
	spaces, err := userconfig.Spaces()
	if err != nil {
		return nil, err
	}
	out := make([]wire.SpaceEntry, 0, len(spaces))
	for _, s := range spaces {
		out = append(out, resolveSpace(s))
	}
	return out, nil
}

// resolveSpace diagnoses one entry by attempting the same discovery a command run there
// would do. Deliberately tolerant — every failure becomes a State + Detail, never an error.
func resolveSpace(s userconfig.Space) wire.SpaceEntry {
	e := wire.SpaceEntry{
		ID: s.ID, Path: s.Path, VerifyID: s.VerifyID, Label: s.Label, Added: s.Added,
	}
	dir := userconfig.ExpandTilde(s.Path)
	if !dirExists(dir) {
		e.State, e.Detail = wire.SpaceStateMissing, "not found at "+s.Path
		return e
	}
	cfg, err := config.Discover(dir)
	if err != nil {
		e.State = wire.SpaceStateNotARepo
		e.Detail = "no planning repo here — run `tskflwctl init` there, or `space forget " + s.ID + "`"
		if errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrValidation) {
			// A config that exists but does not resolve is a DIFFERENT problem from an
			// un-inited directory, and the message already says which.
			e.State, e.Detail = wire.SpaceStateUnreadable, firstLine(err.Error())
		}
		return e
	}
	e.State, e.Root = wire.SpaceStateOK, cfg.Root
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
	if id == "" {
		id = defaultSpaceID(abs)
	}
	if err := validateSpaceID(id); err != nil {
		return err
	}
	space := userconfig.Space{
		ID:       id,
		Path:     userconfig.TildePath(abs),
		VerifyID: cfg.ID,
	}
	if app.DryRun {
		return reportSpaceChange(app, resolveSpace(space), true, "would register")
	}
	added, existing, err := userconfig.AddSpace(space)
	if err != nil {
		return fmt.Errorf("%w: %s", domain.ErrValidation, err.Error())
	}
	if !added {
		return reportSpaceChange(app, resolveSpace(existing), false, "already registered as")
	}
	return reportSpaceChange(app, resolveSpace(existing), true, "registered")
}

func runSpaceForget(app *App, id string) error {
	spaces, err := userconfig.Spaces()
	if err != nil {
		return err
	}
	var found *userconfig.Space
	for i := range spaces {
		if spaces[i].ID == id {
			found = &spaces[i]
			break
		}
	}
	if found == nil {
		return fmt.Errorf("%w: no space named %q — `space list` shows the registered ones", domain.ErrNotFound, id)
	}
	if app.DryRun {
		return reportSpaceChange(app, resolveSpace(*found), true, "would forget")
	}
	removed, err := userconfig.ForgetSpace(id)
	if err != nil {
		return err
	}
	return reportSpaceChange(app, resolveSpace(*found), removed, "forgot")
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

// dirExists reports whether p is an existing directory.
func dirExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
