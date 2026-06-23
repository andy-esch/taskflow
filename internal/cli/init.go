package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/prompt"
	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/config"
)

func newInitCmd(app *App) *cobra.Command {
	var (
		path         string
		planningRepo string
	)
	cmd := &cobra.Command{
		Use:         "init",
		Short:       "Scaffold a planning tree here, or point at an external planning repo",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "mutating"},
		Example: "  tskflwctl init\n" +
			"  tskflwctl init --path ./planning\n" +
			"  tskflwctl init --planning-repo ../desirelines-planning",
		// init may scaffold a NEW planning repo, so it must NOT require an existing
		// one — its own PersistentPreRunE overrides the root's resolve() (skips
		// discovery) and just sets up styling + the Gate/Prompter. The here-vs-
		// elsewhere prompt is TTY-gated (Gate.On()), so a headless agent/pipe never
		// hangs: with --planning-repo it goes straight to pointer mode, otherwise it
		// falls back to the full scaffold (today's non-interactive behavior).
		PersistentPreRunE: func(*cobra.Command, []string) error { app.setStyle(); return nil },
		RunE: func(cmd *cobra.Command, _ []string) error {
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			pointer, repo, err := app.resolveInitTarget(planningRepo, cmd.Flags().Changed("planning-repo"))
			if err != nil {
				return err
			}
			if pointer {
				return runInitPointer(app, abs, repo)
			}
			return runInitScaffold(app, abs)
		},
	}
	cmd.Flags().StringVar(&path, "path", ".", "directory to initialize")
	cmd.Flags().StringVar(&planningRepo, "planning-repo", "",
		"point this repo at an external planning repo (relative to --path, or absolute): writes a pointer config, no tree")
	return cmd
}

// resolveInitTarget decides init's mode (the flag-twin pattern). --planning-repo
// (flagSet) means pointer mode outright. Otherwise, on a TTY (gate open) ask
// here-vs-elsewhere; off a TTY default to scaffold — the headless contract, so an
// agent/pipe never blocks. Extracted so the interactive branch is unit-testable
// with a Fake prompter (PersistentPreRunE's setStyle would otherwise reset Gate).
func (a *App) resolveInitTarget(planningRepo string, flagSet bool) (pointer bool, repo string, err error) {
	if flagSet {
		return true, planningRepo, nil
	}
	if !a.Gate.On() {
		return false, "", nil
	}
	where, err := a.Prompt.SelectOne("Where does this repo's planning live?", []prompt.Option{
		{Label: "Here — scaffold a planning tree in this repo", Value: "here"},
		{Label: "Another repo — point at an external planning repo", Value: "elsewhere"},
	})
	if err != nil {
		return false, "", err
	}
	if where != "elsewhere" {
		return false, "", nil
	}
	repo, err = a.Prompt.Text("Path to the planning repo (relative or absolute)", "../planning")
	if err != nil {
		return false, "", err
	}
	return true, repo, nil
}

// runInitScaffold writes a full planning tree + config under abs (today's init).
func runInitScaffold(app *App, abs string) error {
	created, err := config.Init(abs, app.DryRun)
	if err != nil {
		return err
	}
	if app.JSON {
		return render.InitJSON(app.Out, "scaffold", abs, "", created, app.DryRun)
	}
	if len(created) == 0 {
		fmt.Fprintf(app.Out, "%s already initialized: %s\n", app.Style.Dim("·"), abs)
		return nil
	}
	verb := "initialized"
	if app.DryRun {
		verb = "would initialize"
	}
	fmt.Fprintf(app.Out, "%s %s %s\n", app.Style.Green("✔"), verb, app.Style.Bold(abs))
	for _, c := range created {
		fmt.Fprintf(app.Out, "  %s %s\n", app.Style.Dim("+"), c)
	}
	fmt.Fprintf(app.Out, "\n%s\n", app.Style.Dim(`→ next: tskflwctl epic new "Title" --description "..."`))
	return nil
}

// runInitPointer writes a pointer config under abs (no tree), validating the
// external planning repo first.
func runInitPointer(app *App, abs, planningRepo string) error {
	created, err := config.InitPointer(abs, planningRepo, app.DryRun)
	if err != nil {
		return err
	}
	if app.JSON {
		return render.InitJSON(app.Out, "pointer", abs, planningRepo, created, app.DryRun)
	}
	if len(created) == 0 {
		fmt.Fprintf(app.Out, "%s already initialized: %s\n", app.Style.Dim("·"), abs)
		return nil
	}
	verb := "pointed"
	if app.DryRun {
		verb = "would point"
	}
	fmt.Fprintf(app.Out, "%s %s %s at planning repo %s\n",
		app.Style.Green("✔"), verb, app.Style.Bold(abs), app.Style.Bold(planningRepo))
	fmt.Fprintf(app.Out, "\n%s\n", app.Style.Dim("→ next: run tskflwctl from here — planning resolves to the pointed repo"))
	return nil
}
