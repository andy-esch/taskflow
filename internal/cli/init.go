package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/prompt"
	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/domain"
)

func newInitCmd(app *App) *cobra.Command {
	var (
		path         string
		taskflowRoot string
		planningRepo string
		tracks       []string
		noLinkBack   bool
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold a planning tree here, or point at an external planning repo",
		Long: "Bootstrap a new planning topology: either scaffold a local planning tree or\n" +
			"point this repository at an existing external planning repo. Bare init against\n" +
			"an existing configuration reports its topology without changing it; use\n" +
			"`tskflwctl config migrate` for safe configuration upgrades.",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "mutating"},
		Example: "  tskflwctl init\n" +
			"  tskflwctl init --taskflow-root planning\n" +
			"  tskflwctl init --planning-repo ../desirelines-planning",
		// init may scaffold a NEW planning repo, so it must NOT require an existing
		// one — its own PersistentPreRunE overrides the root's resolve() (skips
		// discovery) and just sets up styling + the Gate/Prompter. The here-vs-
		// elsewhere prompt is TTY-gated (Gate.On()), so a headless agent/pipe never
		// hangs: with --planning-repo it goes straight to pointer mode, otherwise it
		// falls back to the full scaffold (today's non-interactive behavior).
		PersistentPreRunE: app.styleOnlyPreRun,
		RunE: func(cmd *cobra.Command, _ []string) error {
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			// A bare init against an existing config is an identity read, never a hidden
			// migration or a scaffold attempt. Explicit topology flags retain their existing
			// validation paths; only the natural re-run takes this lifecycle handoff.
			if existing, ok := config.Describe(abs); ok && !initTopologyFlagsChanged(cmd, tracks) {
				return runInitExisting(app, abs, existing)
			}
			pointer, repo, chosenRoot, err := app.resolveInitTarget(
				abs, planningRepo, cmd.Flags().Changed("planning-repo"), taskflowRoot, cmd.Flags().Changed("taskflow-root"))
			if err != nil {
				return err
			}
			if pointer {
				// tracked_repos lives in a PLANNING repo; a pointer repo only points.
				if len(tracks) > 0 {
					return fmt.Errorf("%w: --track records repos a PLANNING repo tracks; it can't combine with --planning-repo (pointer mode)", domain.ErrValidation)
				}
				if cmd.Flags().Changed("taskflow-root") {
					return fmt.Errorf("%w: --taskflow-root places a tree scaffolded HERE; it can't combine with --planning-repo (pointer mode, no tree)", domain.ErrValidation)
				}
				return runInitPointer(app, abs, repo, !noLinkBack)
			}
			// --no-link-back is pointer-only; reject it in scaffold mode for symmetry
			// with the --track guard above (don't silently ignore a misused flag).
			if cmd.Flags().Changed("no-link-back") {
				return fmt.Errorf("%w: --no-link-back only applies with --planning-repo (pointer mode)", domain.ErrValidation)
			}
			return runInitScaffold(app, abs, chosenRoot, tracks)
		},
	}
	cmd.Flags().StringVar(&path, "path", ".", "directory to initialize")
	cmd.Flags().StringVar(&taskflowRoot, "taskflow-root", "",
		"scaffold the planning tree in this subdirectory instead of the repo root (sets taskflow_root; e.g. planning)")
	cmd.Flags().StringVar(&planningRepo, "planning-repo", "",
		"point this repo at an external planning repo (relative to --path, or absolute): writes a pointer config, no tree")
	cmd.Flags().StringSliceVar(&tracks, "track", nil,
		"record an impl repo this planning repo tracks (repeatable; scaffold mode only)")
	cmd.Flags().BoolVar(&noLinkBack, "no-link-back", false,
		"pointer mode: don't add this repo to the planning repo's tracked_repos")
	return cmd
}

func initTopologyFlagsChanged(cmd *cobra.Command, tracks []string) bool {
	return cmd.Flags().Changed("planning-repo") || cmd.Flags().Changed("taskflow-root") ||
		cmd.Flags().Changed("no-link-back") || len(tracks) > 0
}

func runInitExisting(app *App, abs string, description config.Description) error {
	pending, err := config.PendingMigrations(abs)
	if err != nil {
		return err
	}
	mode := "scaffold"
	if description.PlanningRepo != "" {
		mode = "pointer"
	}
	pendingNames := make([]string, len(pending))
	for i, migration := range pending {
		pendingNames[i] = string(migration)
	}
	if app.JSON {
		return render.InitJSON(app.Out, render.InitEnvelope{
			DryRun: app.DryRun, Mode: mode, Root: abs,
			PlanningRepo: description.PlanningRepo, AlreadyInitialized: true,
			PendingMigrations: pendingNames,
		})
	}
	fmt.Fprintf(app.Out, "%s already initialized: %s\n", app.Style.Dim("·"), abs)
	printLayout(app, abs)
	if len(pendingNames) > 0 {
		fmt.Fprintf(app.Out, "\n%s configuration update available: %s\n", app.Style.Warn("⚠"), strings.Join(pendingNames, ", "))
		fmt.Fprintln(app.Out, app.Style.Dim("→ tskflwctl config migrate"))
	}
	return nil
}

// resolveInitTarget decides init's mode (the flag-twin pattern) AND, for scaffold mode,
// where in this repo the tree goes. --planning-repo (flagSet) means pointer mode outright.
// Otherwise, on a TTY (gate open) ask here-vs-elsewhere; off a TTY default to scaffold —
// the headless contract, so an agent/pipe never blocks.
//
// The second prompt exists because "here" used to mean "the repo root", always: there was
// no way to reach the config-at-root/tree-in-a-subdir layout this repo itself uses, and a
// user who typed `planning/` at the elsewhere prompt got told to go init it there first.
// Interactively we now recommend `planning/`; the NON-interactive default stays the repo
// root, so scripts and CI are unchanged.
//
// Extracted so the interactive branches are unit-testable with a Fake prompter
// (PersistentPreRunE's setStyle would otherwise reset Gate).
func (a *App) resolveInitTarget(dir, planningRepo string, repoFlagSet bool, taskflowRoot string, rootFlagSet bool) (pointer bool, repo, root string, err error) {
	if repoFlagSet {
		return true, planningRepo, "", nil
	}
	if rootFlagSet {
		return false, "", taskflowRoot, nil // explicit placement: nothing to ask
	}
	if !a.Gate.On() {
		return false, "", taskflowRoot, nil
	}
	// An already-initialized repo has settled this. Asking again would offer two answers
	// that are refused (they would fork the data) and one that is a no-op — a menu whose
	// options mostly error is worse than no menu. Fall through to the repair path, which
	// reports what is configured instead.
	if _, exists := config.Describe(dir); exists {
		return false, "", "", nil
	}
	where, err := a.Prompt.SelectOne("Where does this repo's planning live?", []prompt.Option{
		{Label: "Here — scaffold a planning tree in this repo", Value: "here"},
		{Label: "Another repo — point at an EXISTING planning repo", Value: "elsewhere"},
	})
	if err != nil {
		return false, "", "", err
	}
	if where == "elsewhere" {
		// Named as existing on purpose: this path validates the target and errors if it
		// isn't a planning repo yet, so promising less would be a lie.
		repo, err = a.Prompt.Text("Path to that existing planning repo (relative or absolute)", "../planning")
		if err != nil {
			return false, "", "", err
		}
		return true, repo, "", nil
	}
	root, err = a.promptTaskflowRoot()
	return false, "", root, err
}

// promptTaskflowRoot asks where in this repo the tree goes. `planning/` leads because it
// keeps the repo root clean and is the layout tskflwctl itself uses; the root remains one
// keystroke away for anyone who wants a planning-only repo.
func (a *App) promptTaskflowRoot() (string, error) {
	choice, err := a.Prompt.SelectOne("Where in this repo?", []prompt.Option{
		{Label: "planning/ — keep the repo root clean (recommended)", Value: "planning"},
		{Label: "The repo root — for a planning-only repo", Value: "."},
		{Label: "Somewhere else…", Value: "custom"},
	})
	if err != nil || choice != "custom" {
		return choice, err
	}
	return a.Prompt.Text("Subdirectory for the planning tree (relative to this repo)", "docs/planning")
}

// runInitScaffold writes a full planning tree + config under abs, then records
// any --track impl repos in its tracked_repos (deduped, surgical).
func runInitScaffold(app *App, abs, taskflowRoot string, tracks []string) error {
	created, err := config.Init(abs, taskflowRoot, app.DryRun)
	if err != nil {
		return err
	}
	var tracked []string
	for _, tr := range tracks {
		added, err := config.AddTrackedRepo(abs, tr, app.DryRun)
		if err != nil {
			return err
		}
		if added {
			tracked = append(tracked, tr)
		}
	}
	if app.JSON {
		return render.InitJSON(app.Out, render.InitEnvelope{
			DryRun: app.DryRun, Mode: "scaffold", Root: abs, Tracked: tracked, Created: created,
		})
	}
	if len(created) == 0 && len(tracked) == 0 {
		fmt.Fprintf(app.Out, "%s already initialized: %s\n", app.Style.Dim("·"), abs)
		printLayout(app, abs)
		return nil
	}
	verb := "initialized"
	switch {
	case app.DryRun && len(created) == 0:
		verb = "would update"
	case app.DryRun:
		verb = "would initialize"
	case len(created) == 0:
		verb = "updated"
	}
	fmt.Fprintf(app.Out, "%s %s %s\n", app.Style.Green("✔"), verb, app.Style.Bold(abs))
	for _, c := range created {
		fmt.Fprintf(app.Out, "  %s %s\n", app.Style.Dim("+"), c)
	}
	for _, tr := range tracked {
		fmt.Fprintf(app.Out, "  %s tracks %s\n", app.Style.Dim("+"), app.Style.Bold(tr))
	}
	printLayout(app, abs)
	if len(created) > 0 {
		fmt.Fprintf(app.Out, "\n%s\n", app.Style.Dim(`→ next: tskflwctl epic new "Title" --description "..."`))
	}
	return nil
}

// printLayout shows what a repo is actually configured for. On a re-run this is the whole
// point: init is a REPAIR there, not a scaffold, so the useful thing to print is the
// settled layout rather than a question about it.
func printLayout(app *App, dir string) {
	d, ok := config.Describe(dir)
	if !ok {
		return
	}
	row := func(k, v string) {
		if v != "" {
			fmt.Fprintf(app.Out, "    %s %s\n", app.Style.Dim(fmt.Sprintf("%-14s", k)), v)
		}
	}
	if d.PlanningRepo != "" {
		row("planning repo", d.PlanningRepo)
		if d.PlanningRepoID == "" {
			row("verifies", app.Style.Dim("no — run `tskflwctl config migrate`"))
		} else {
			row("verifies", d.PlanningRepoID)
		}
		return
	}
	row("planning tree", d.TaskflowRoot)
	if d.ID == "" {
		row("durable id", app.Style.Dim("none yet — run `tskflwctl config migrate`"))
	} else {
		row("durable id", d.ID)
	}
}

// runInitPointer writes a pointer config under abs (no tree), validating the
// external planning repo first, then (unless opted out) links back by recording
// this repo in the planning repo's tracked_repos.
func runInitPointer(app *App, abs, planningRepo string, linkBack bool) error {
	created, err := config.InitPointer(abs, planningRepo, app.DryRun)
	if err != nil {
		return err
	}
	var back string
	var linkErr error
	if linkBack {
		back, linkErr = config.LinkBack(abs, planningRepo, app.DryRun)
	}
	// Link-back is best-effort: the pointer config is already written, so a hiccup
	// (e.g. the planning repo isn't writable) warns rather than fails the init. The
	// warning goes to stderr, after the success line, so it never corrupts --json
	// stdout and reads in order on a combined terminal.
	warn := func() {
		if linkErr != nil {
			fmt.Fprintf(app.ErrOut, "%s link-back skipped: %v\n", app.Style.Warn("⚠"), linkErr)
		}
	}
	if app.JSON {
		warn()
		return render.InitJSON(app.Out, render.InitEnvelope{
			DryRun: app.DryRun, Mode: "pointer", Root: abs, PlanningRepo: planningRepo, LinkedBack: back, Created: created,
		})
	}
	if len(created) > 0 {
		// A pre-existing pointer that only gained planning_repo_id is a BACKFILL, not a
		// fresh init — saying "pointed X at Y" would misreport what changed.
		if backfilled := len(created) == 1 && strings.Contains(created[0], "planning_repo_id"); backfilled {
			verb := "recorded"
			if app.DryRun {
				verb = "would record"
			}
			fmt.Fprintf(app.Out, "%s %s planning_repo_id for %s — this pointer now verifies its target\n",
				app.Style.Green("✔"), verb, app.Style.Bold(planningRepo))
		} else {
			verb := "pointed"
			if app.DryRun {
				verb = "would point"
			}
			fmt.Fprintf(app.Out, "%s %s %s at planning repo %s\n",
				app.Style.Green("✔"), verb, app.Style.Bold(abs), app.Style.Bold(planningRepo))
		}
	} else {
		fmt.Fprintf(app.Out, "%s already initialized: %s\n", app.Style.Dim("·"), abs)
	}
	printLayout(app, abs)
	if back != "" {
		verb := "linked back"
		if app.DryRun {
			verb = "would link back"
		}
		fmt.Fprintf(app.Out, "  %s %s — %s now tracks this repo as %s\n",
			app.Style.Dim("+"), verb, app.Style.Bold(planningRepo), app.Style.Bold(back))
	}
	warn()
	if len(created) > 0 {
		fmt.Fprintf(app.Out, "\n%s\n", app.Style.Dim("→ next: run tskflwctl from here — planning resolves to the pointed repo"))
	}
	return nil
}
