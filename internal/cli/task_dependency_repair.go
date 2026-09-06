package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
	"github.com/andy-esch/taskflow/internal/wire"
)

type graphRepairCommandFailure struct {
	cause     error
	receipt   core.TaskGraphRepairReceipt
	workspace wire.WorkspaceJSON
}

func (e *graphRepairCommandFailure) Error() string { return e.cause.Error() }
func (e *graphRepairCommandFailure) Unwrap() error { return e.cause }

type graphRepairManifest struct {
	Schema     int                            `yaml:"schema"`
	Operations []graphRepairManifestOperation `yaml:"operations"`
}

type graphRepairManifestOperation struct {
	Action     string `yaml:"action"`
	Task       string `yaml:"task"`
	Location   string `yaml:"location"`
	Field      string `yaml:"field"`
	Value      string `yaml:"value"`
	Occurrence int    `yaml:"occurrence"`
}

func newTaskDependencyRepairCmd(app *App) *cobra.Command {
	var auto bool
	var drops, dedupes []string
	var manifestPath string
	cmd := &cobra.Command{
		Use:     "repair",
		Short:   "Diagnose or repair broken graph-owned declarations",
		Long:    "With no selector, diagnose exact graph-owned source defects and print copyable repair commands. --auto applies only canonical deduplication, self-edge removal, and empty legacy-key cleanup. Use repeatable --drop/--dedupe selectors or a YAML --plan for explicit, reauthorized source removals. Invalid and dangling values, cycle choices, and ambiguous legacy intent are never guessed.",
		Example: "  tskflwctl task depend repair\n  tskflwctl task depend repair --auto --dry-run\n  tskflwctl task depend repair --drop 'tasks/ID-task.md:depends_on=RAW#0'\n  tskflwctl task depend repair --plan repair.yaml --json",
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			mutationSelected := auto || len(drops) > 0 || len(dedupes) > 0 || manifestPath != ""
			if !mutationSelected {
				receipt, err := app.Svc.InspectTaskGraphRepair()
				if err != nil {
					return err
				}
				return emitTaskGraphRepair(app, receipt)
			}
			request := core.TaskGraphRepairRequest{Auto: auto}
			for _, selector := range drops {
				edit, err := parseGraphRepairSelector(selector, core.TaskGraphSourceDropDeclaration)
				if err != nil {
					return err
				}
				request.Edits = append(request.Edits, normalizeGraphRepairLocation(app, edit))
			}
			for _, selector := range dedupes {
				edit, err := parseGraphRepairSelector(selector, core.TaskGraphSourceDedupe)
				if err != nil {
					return err
				}
				request.Edits = append(request.Edits, normalizeGraphRepairLocation(app, edit))
			}
			if manifestPath != "" {
				manifestEdits, err := readGraphRepairManifest(manifestPath)
				if err != nil {
					return err
				}
				for _, edit := range manifestEdits {
					request.Edits = append(request.Edits, normalizeGraphRepairLocation(app, edit))
				}
			}
			receipt, err := app.Svc.RepairTaskGraph(request, app.DryRun)
			if err != nil {
				if len(receipt.AppliedFiles) == 0 {
					return err
				}
				return &graphRepairCommandFailure{cause: err, receipt: receipt, workspace: app.workspace()}
			}
			return emitTaskGraphRepair(app, receipt)
		},
	}
	cmd.Annotations = map[string]string{"safety": "mutating"}
	cmd.Flags().BoolVar(&auto, "auto", false, "apply only canonical dedupe, self-edge, and empty legacy-key repairs")
	cmd.Flags().StringArrayVar(&drops, "drop", nil, "drop an exact defective source declaration: <task-or-path>:<field>=<raw-value>[#occurrence]")
	cmd.Flags().StringArrayVar(&dedupes, "dedupe", nil, "deduplicate one canonical source value: <task-or-path>:depends_on=<task-id>")
	cmd.Flags().StringVar(&manifestPath, "plan", "", "read additional convergent repair operations from YAML")
	return cmd
}

func normalizeGraphRepairLocation(app *App, edit core.TaskGraphSourceEdit) core.TaskGraphSourceEdit {
	if edit.Source.Location != "" && !filepath.IsAbs(edit.Source.Location) && app.Cfg != nil {
		edit.Source.Location = filepath.Join(app.Cfg.Root, edit.Source.Location)
	}
	return edit
}

func emitTaskGraphRepair(app *App, receipt core.TaskGraphRepairReceipt) error {
	if app.JSON {
		return render.TaskGraphRepairJSON(app.Out, receipt, app.workspace())
	}
	return render.TaskGraphRepairHuman(app.Out, app.Style, receipt, app.workspace())
}

func parseGraphRepairSelector(selector string, action core.TaskGraphSourceEditAction) (core.TaskGraphSourceEdit, error) {
	sourceText, field, value, ok := splitGraphRepairSelector(selector)
	if !ok {
		return core.TaskGraphSourceEdit{}, fmt.Errorf("%w: repair selector %q must be <task-or-path>:<field>=<raw-value>[#occurrence]", domain.ErrValidation, selector)
	}
	occurrence := 0
	if action == core.TaskGraphSourceDropDeclaration {
		if hash := strings.LastIndex(value, "#"); hash >= 0 && hash < len(value)-1 {
			if parsed, err := strconv.Atoi(value[hash+1:]); err == nil {
				if parsed < 0 {
					return core.TaskGraphSourceEdit{}, fmt.Errorf("%w: repair declaration occurrence must be non-negative", domain.ErrValidation)
				}
				occurrence, value = parsed, value[:hash]
			}
		}
	}
	source := core.TaskGraphSourceRef{}
	if filepath.IsAbs(sourceText) || strings.ContainsRune(sourceText, filepath.Separator) {
		source.Location = sourceText
	} else if id.Valid(sourceText) {
		source.TaskID = sourceText
	} else {
		source.TaskSlug = sourceText
	}
	return core.TaskGraphSourceEdit{Action: action, Source: source, Field: field, Value: value, Occurrence: occurrence}, nil
}

func splitGraphRepairSelector(selector string) (string, core.TaskDependencyField, string, bool) {
	fields := []core.TaskDependencyField{
		core.TaskDependencyDependsOn, core.TaskDependencyBlockedBy,
		core.TaskDependencyDependencies, core.TaskDependencyBlocks,
	}
	best := -1
	var matched core.TaskDependencyField
	for _, field := range fields {
		marker := ":" + string(field) + "="
		if index := strings.Index(selector, marker); index > 0 && (best < 0 || index < best) {
			best, matched = index, field
		}
	}
	if best < 0 {
		return "", "", "", false
	}
	markerLength := len(matched) + 2
	return selector[:best], matched, selector[best+markerLength:], true
}

func validRepairField(field core.TaskDependencyField) bool {
	switch field {
	case core.TaskDependencyDependsOn, core.TaskDependencyBlockedBy, core.TaskDependencyDependencies, core.TaskDependencyBlocks:
		return true
	default:
		return false
	}
}

func readGraphRepairManifest(path string) ([]core.TaskGraphSourceEdit, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read graph repair plan %s: %w", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	var manifest graphRepairManifest
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("%w: decode graph repair plan %s: %v", domain.ErrValidation, path, err)
	}
	if manifest.Schema != 1 {
		return nil, fmt.Errorf("%w: graph repair plan schema must be 1, got %d", domain.ErrValidation, manifest.Schema)
	}
	if len(manifest.Operations) == 0 {
		return nil, fmt.Errorf("%w: graph repair plan contains no operations", domain.ErrValidation)
	}
	edits := make([]core.TaskGraphSourceEdit, 0, len(manifest.Operations))
	for index, operation := range manifest.Operations {
		action := core.TaskGraphSourceEditAction(operation.Action)
		switch action {
		case core.TaskGraphSourceDropDeclaration, core.TaskGraphSourceDedupe, core.TaskGraphSourceDropEmptyField:
		default:
			return nil, fmt.Errorf("%w: graph repair plan operation %d has unsupported action %q", domain.ErrValidation, index+1, operation.Action)
		}
		if !validRepairField(core.TaskDependencyField(operation.Field)) {
			return nil, fmt.Errorf("%w: graph repair plan operation %d has unsupported field %q", domain.ErrValidation, index+1, operation.Field)
		}
		if operation.Occurrence < 0 {
			return nil, fmt.Errorf("%w: graph repair plan operation %d occurrence must be non-negative", domain.ErrValidation, index+1)
		}
		source := core.TaskGraphSourceRef{Location: operation.Location}
		if operation.Task != "" {
			if id.Valid(operation.Task) {
				source.TaskID = operation.Task
			} else {
				source.TaskSlug = operation.Task
			}
		}
		if source.TaskID == "" && source.TaskSlug == "" && source.Location == "" {
			return nil, fmt.Errorf("%w: graph repair plan operation %d requires task or location", domain.ErrValidation, index+1)
		}
		edits = append(edits, core.TaskGraphSourceEdit{
			Action: action, Source: source, Field: core.TaskDependencyField(operation.Field),
			Value: operation.Value, Occurrence: operation.Occurrence,
		})
	}
	return edits, nil
}
