package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

type threadApplyCommandFailure struct {
	cause     error
	receipt   core.ThreadApplyReceipt
	planPath  string
	workspace wire.WorkspaceJSON
}

func (e *threadApplyCommandFailure) Error() string { return e.cause.Error() }
func (e *threadApplyCommandFailure) Unwrap() error { return e.cause }

func newThreadComposeCmd(app *App) *cobra.Command {
	var from, outPath string
	cmd := &cobra.Command{
		Use:   "compose",
		Short: "Compile existing tasks and dependency edges into a durable Thread apply plan",
		Long: "Read one strict literal YAML/JSON manifest, resolve its exact stable task IDs, and validate the proposed global DAG without mutation. " +
			"Nodes marked member: false may supply transitive upstream graph context without entering Thread membership; ordinary Thread views expose only direct membership-boundary prerequisites as external gates. " +
			"A real run creates a no-clobber materialized plan; --dry-run prints the same plan without creating the output file.",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "mutating"},
		RunE: func(_ *cobra.Command, _ []string) error {
			if strings.TrimSpace(from) == "" || strings.TrimSpace(outPath) == "" {
				return fmt.Errorf("%w: --from and --out are required", domain.ErrValidation)
			}
			if outPath == "-" {
				return fmt.Errorf("%w: --out must name a durable file path, not stdout", domain.ErrValidation)
			}
			content, err := readThreadApplyInput(app.In, from)
			if err != nil {
				return err
			}
			var manifest core.ThreadComposeManifest
			if err := decodeStrictThreadYAML(content, &manifest, "Thread authoring manifest"); err != nil {
				return err
			}
			plan, err := app.Svc.ComposeThreadApply(app.Cfg.ID, manifest)
			if err != nil {
				return err
			}
			encoded, err := yaml.Marshal(plan)
			if err != nil {
				return fmt.Errorf("encode Thread apply plan: %w", err)
			}
			if app.DryRun {
				if err := requireThreadApplyPlanAbsent(outPath); err != nil {
					return err
				}
			} else {
				if err := writeThreadApplyPlan(outPath, encoded); err != nil {
					return err
				}
			}
			if app.JSON {
				return render.ThreadApplyComposeJSON(app.Out, plan, outPath, app.DryRun, app.workspace())
			}
			if app.DryRun {
				fmt.Fprintf(app.Out, "would compose Thread %s (%s) -> %s\n", plan.Thread.ID, plan.Thread.Slug, outPath)
				_, err = app.Out.Write(encoded)
				return err
			}
			fmt.Fprintf(app.Out, "composed Thread %s (%s) -> %s\n", plan.Thread.ID, plan.Thread.Slug, outPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "strict authoring manifest path, or - for stdin")
	cmd.Flags().StringVar(&outPath, "out", "", "new durable apply-plan path (must not already exist)")
	return cmd
}

func newThreadApplyCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "apply <materialized-plan>",
		Short: "Converge dependencies and one new Thread from a durable plan",
		Long: "Revalidate planning identity and repository health under one mutation guard, apply additive dependency writes first, and create the Thread last. " +
			"An interrupted plan is safe to retry from the same durable file.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"safety": "mutating"},
		RunE: func(_ *cobra.Command, args []string) error {
			if args[0] == "-" {
				return fmt.Errorf("%w: Thread apply requires a durable plan path; stdin is not accepted", domain.ErrValidation)
			}
			content, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read Thread apply plan %s: %w", args[0], err)
			}
			var plan core.ThreadApplyPlan
			if err := decodeStrictThreadYAML(content, &plan, "Thread apply plan"); err != nil {
				return err
			}
			receipt, err := app.Svc.ApplyThreadPlan(plan, app.DryRun)
			if err != nil {
				var failure *core.ThreadApplyFailure
				if errors.As(err, &failure) {
					return &threadApplyCommandFailure{
						cause: err, receipt: failure.Receipt, planPath: args[0], workspace: app.workspace(),
					}
				}
				return err
			}
			if app.JSON {
				return render.ThreadApplyJSON(app.Out, receipt, args[0], app.workspace())
			}
			render.ThreadApplyHuman(app.Out, app.Style, receipt)
			return nil
		},
	}
}

func readThreadApplyInput(in io.Reader, path string) ([]byte, error) {
	if path == "-" {
		content, err := io.ReadAll(in)
		if err != nil {
			return nil, fmt.Errorf("read Thread manifest from stdin: %w", err)
		}
		return content, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Thread manifest %s: %w", path, err)
	}
	return content, nil
}

func decodeStrictThreadYAML(content []byte, target any, kind string) error {
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: parse %s: %v", domain.ErrValidation, kind, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("%w: %s must contain exactly one YAML/JSON document", domain.ErrValidation, kind)
		}
		return fmt.Errorf("%w: parse trailing %s content: %v", domain.ErrValidation, kind, err)
	}
	return nil
}

func requireThreadApplyPlanAbsent(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("thread apply plan %s already exists: %w", path, domain.ErrConflict)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect Thread apply plan %s: %w", path, err)
	}
	return nil
}

func writeThreadApplyPlan(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create Thread apply-plan directory %s: %w", dir, err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("thread apply plan %s already exists: %w", path, domain.ErrConflict)
		}
		return fmt.Errorf("create Thread apply plan %s: %w", path, err)
	}
	remove := true
	defer func() {
		_ = file.Close()
		if remove {
			_ = os.Remove(path)
		}
	}()
	if _, err := file.Write(content); err != nil {
		return fmt.Errorf("write Thread apply plan %s: %w", path, err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync Thread apply plan %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close Thread apply plan %s: %w", path, err)
	}
	// Match entity-create durability: the recovery token's directory entry is
	// best-effort fsync'd after its contents, so compose never reports success
	// before the plan is as durable as the mutations it is meant to recover.
	if directory, openErr := os.Open(dir); openErr == nil {
		_ = directory.Sync()
		_ = directory.Close()
	}
	remove = false
	return nil
}
