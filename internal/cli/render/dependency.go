package render

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/wire"
)

// TaskGraphRepairJSON writes the guarded source-declaration repair receipt.
func TaskGraphRepairJSON(w io.Writer, receipt core.TaskGraphRepairReceipt, workspace wire.WorkspaceJSON) error {
	return wire.EncodeJSON(w, wire.ToTaskGraphRepairEnvelope(receipt, workspace))
}

// TaskGraphRepairHuman keeps diagnosis terse while making every destructive
// choice copyable. Exact source paths are preferred over potentially ambiguous
// duplicate task IDs.
func TaskGraphRepairHuman(w io.Writer, st Style, receipt core.TaskGraphRepairReceipt, workspace wire.WorkspaceJSON) error {
	fmt.Fprintf(w, "%s %s -> %s · changed=%t · committed=%t\n", st.Dim("graph:"), receipt.InitialHealth, receipt.FinalHealth, receipt.Changed, receipt.Committed)
	if workspace.PlanningRoot != "" {
		fmt.Fprintf(w, "%s %s\n", st.Dim("workspace:"), workspace.PlanningRoot)
	}
	if len(receipt.Selected) == 0 {
		auto, explicit, direct := 0, 0, 0
		for _, defect := range receipt.Residual {
			switch {
			case !defect.Repairable:
				direct++
			case defect.Automatic:
				auto++
			default:
				explicit++
			}
		}
		if auto == 0 && explicit == 0 && direct == 0 {
			fmt.Fprintf(w, "%s no broken graph-owned declarations found\n", st.Green("✔"))
			return nil
		}
		if auto > 0 {
			fmt.Fprintf(w, "%s %d inferable repair(s); apply with `tskflwctl task depend repair --auto`\n", st.Bold("inferable:"), auto)
		}
		for _, defect := range receipt.Residual {
			if !defect.Repairable {
				continue
			}
			detail := repairEditDisplay(defect.Target)
			if defect.HasProjectedEdge {
				detail += fmt.Sprintf("  edge %s -> %s", defect.ProjectedEdge.From, defect.ProjectedEdge.To)
			}
			if len(defect.CandidateIDs) > 0 {
				detail += "  candidates " + strings.Join(defect.CandidateIDs, ", ")
			}
			fmt.Fprintf(w, "%s %s  %s\n", st.Dim("•"), defect.Reason, detail)
			if !defect.Automatic {
				flag := "--drop"
				if defect.Target.Action == core.TaskGraphSourceDedupe {
					flag = "--dedupe"
				}
				fmt.Fprintf(w, "  %s\n", st.Dim("tskflwctl task depend repair "+flag+" "+shellQuote(repairEditSelector(defect.Target))))
			}
		}
		for _, defect := range receipt.Residual {
			if defect.Repairable {
				continue
			}
			fmt.Fprintf(w, "%s %s/%s: %s\n", st.Warn("⚠"), defect.Reason, defect.Problem.Code, defect.Problem.Message)
		}
		if len(receipt.IncompleteThreads) > 0 {
			fmt.Fprintf(w, "%s %d unreadable Thread document(s); impact evidence is incomplete\n", st.Warn("⚠"), len(receipt.IncompleteThreads))
			for _, problem := range receipt.IncompleteThreads {
				location := problem.Location
				if location == "" {
					location = problem.ThreadID
				}
				fmt.Fprintf(w, "  %s %s: %s\n", st.Dim("•"), location, problem.Message)
			}
		}
		fmt.Fprintln(w, st.Dim("nothing was written; select --auto, --drop, --dedupe, or --plan"))
		return nil
	}
	operations := make(map[core.TaskGraphSourceEdit]core.TaskGraphRepairOperation, len(receipt.Operations))
	for _, operation := range receipt.Operations {
		operations[operation.Edit] = operation
	}
	fmt.Fprintf(w, "%s %d intent(s), %d active operation(s)\n", st.Dim("selected:"), len(receipt.Selected), len(receipt.Operations))
	for _, selection := range receipt.Selected {
		operation, active := operations[selection]
		mode := "explicit"
		reason := "already satisfied"
		if active {
			reason = string(operation.Reason)
			if operation.Automatic {
				mode = "auto"
			}
		} else {
			mode = "satisfied"
		}
		fmt.Fprintf(w, "%s %s %s  %s\n", st.Dim("•"), mode, repairEditDisplay(selection), st.Dim(reason))
	}
	verb, prefix := "removed", st.Green("✔")
	if receipt.DryRun {
		verb, prefix = "would remove", st.Dim("◇")
	}
	for _, declaration := range receipt.Removed {
		fmt.Fprintf(w, "%s %s %s:%s=%s#%d\n", prefix, verb, repairSourceDisplay(declaration.Source), declaration.Field, strconv.Quote(declaration.Value), declaration.Occurrence)
	}
	if !receipt.Changed {
		fmt.Fprintf(w, "%s selected repair intent is already satisfied\n", st.Dim("•"))
	}
	for _, files := range []struct {
		label  string
		values []string
	}{
		{label: "planned files", values: receipt.PlannedFiles},
		{label: "applied files", values: receipt.AppliedFiles},
		{label: "remaining files", values: receipt.RemainingFiles},
	} {
		if len(files.values) > 0 {
			fmt.Fprintf(w, "%s\n", st.Dim(files.label+": "+strings.Join(files.values, ", ")))
		}
	}
	for _, impact := range receipt.Impacts {
		fmt.Fprintf(w, "%s task %s state %s -> %s\n", st.Dim("•"), impact.TaskID,
			taskGraphStateSummary(impact.Before), taskGraphStateSummary(impact.After))
	}
	for _, impact := range receipt.ThreadImpacts {
		name := impact.Slug
		if name == "" {
			name = impact.ThreadID
		}
		fmt.Fprintf(w, "%s Thread %s (%s) projection %s -> %s; changed tasks: %s\n", st.Dim("•"), name,
			impact.ThreadID, impact.Before.ProjectionHealth, impact.After.ProjectionHealth, strings.Join(impact.ChangedTaskIDs, ", "))
	}
	if len(receipt.Residual) > 0 {
		fmt.Fprintf(w, "%s %d residual defect(s); run `tskflwctl task depend repair` to inspect\n", st.Warn("⚠"), len(receipt.Residual))
		for _, defect := range receipt.Residual {
			if defect.Repairable {
				fmt.Fprintf(w, "  %s %s  %s\n", st.Dim("•"), defect.Reason, repairEditDisplay(defect.Target))
			} else {
				fmt.Fprintf(w, "  %s %s/%s: %s\n", st.Warn("⚠"), defect.Reason, defect.Problem.Code, defect.Problem.Message)
			}
		}
	}
	if len(receipt.IncompleteThreads) > 0 {
		fmt.Fprintf(w, "%s %d unreadable Thread document(s); impact evidence is incomplete\n", st.Warn("⚠"), len(receipt.IncompleteThreads))
		for _, problem := range receipt.IncompleteThreads {
			location := problem.Location
			if location == "" {
				location = problem.ThreadID
			}
			fmt.Fprintf(w, "  %s %s: %s\n", st.Dim("•"), location, problem.Message)
		}
	}
	return nil
}

func taskGraphStateSummary(state core.TaskGraphState) string {
	return fmt.Sprintf("%s/%s sound=%t eligible=%t drained=%t inconsistent=%t",
		state.Role, state.Gate, state.SoundlyCompleted, state.Eligible, state.Drained, state.Inconsistent)
}

func repairEditDisplay(edit core.TaskGraphSourceEdit) string {
	return fmt.Sprintf("%s:%s=%s#%d", repairSourceDisplay(edit.Source), edit.Field, strconv.Quote(edit.Value), edit.Occurrence)
}

func repairEditSelector(edit core.TaskGraphSourceEdit) string {
	selector := fmt.Sprintf("%s:%s=%s", repairSourceDisplay(edit.Source), edit.Field, edit.Value)
	if edit.Action == core.TaskGraphSourceDropDeclaration {
		selector += fmt.Sprintf("#%d", edit.Occurrence)
	}
	return selector
}

func repairSourceDisplay(source core.TaskGraphSourceRef) string {
	if source.Location != "" {
		return source.Location
	}
	if source.TaskID != "" {
		return source.TaskID
	}
	return source.TaskSlug
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

// DependencyMutationJSON writes the stable guarded dependency receipt.
func DependencyMutationJSON(w io.Writer, receipt core.DependencyMutationReceipt, workspace wire.WorkspaceJSON) error {
	return wire.EncodeJSON(w, wire.ToDependencyMutationEnvelope(receipt, workspace))
}

// DependencyMutationHuman explains every edge outcome and migration prefix.
func DependencyMutationHuman(w io.Writer, st Style, receipt core.DependencyMutationReceipt) error {
	if !receipt.Changed {
		fmt.Fprintf(w, "%s dependency graph already satisfies the %s request\n", st.Dim("•"), receipt.Operation)
	}
	for _, edge := range receipt.Edges {
		if edge.Outcome == "skipped" {
			fmt.Fprintf(w, "%s %s %s -> %s (already satisfied)\n", st.Dim("•"), edge.Action, edge.PrerequisiteID, edge.DependentID)
			continue
		}
		prefix := "✔"
		verb := edge.Outcome
		if receipt.DryRun {
			prefix, verb = "◇", "would be "+edge.Outcome
		}
		fmt.Fprintf(w, "%s %s %s -> %s\n", st.Green(prefix), verb, edge.PrerequisiteID, edge.DependentID)
	}
	for _, impact := range receipt.Impacts {
		prefix := st.Dim("•")
		newlyUnsafe := (impact.Before.Gate != impact.After.Gate && impact.After.Gate != core.GateClear) ||
			(!impact.Before.Inconsistent && impact.After.Inconsistent)
		if newlyUnsafe {
			prefix = st.Warn("⚠")
		}
		fmt.Fprintf(w, "%s %s state %s/%s -> %s/%s\n", prefix, impact.TaskID,
			impact.Before.Role, impact.Before.Gate, impact.After.Role, impact.After.Gate)
		if newlyUnsafe {
			fmt.Fprintf(w, "  %s inspect with `tskflwctl task blockers %s`; restore a sound prerequisite or remove the edge\n", st.Dim("remedy:"), impact.TaskID)
		}
	}
	if len(receipt.ClearedLegacyFields) > 0 {
		verb := "cleared"
		prefix := "✔"
		if receipt.DryRun {
			verb = "would clear"
			prefix = "◇"
		}
		for _, clear := range receipt.ClearedLegacyFields {
			fmt.Fprintf(w, "%s %s %s on %s\n", st.Green(prefix), verb, clear.Field, clear.TaskID)
		}
	}
	if len(receipt.AppliedTaskIDs) > 0 {
		fmt.Fprintf(w, "%s\n", st.Dim("applied task files: "+strings.Join(receipt.AppliedTaskIDs, ", ")))
	} else if receipt.DryRun && len(receipt.PlannedTaskIDs) > 0 {
		fmt.Fprintf(w, "%s\n", st.Dim("planned task files: "+strings.Join(receipt.PlannedTaskIDs, ", ")))
	}
	return nil
}

// TaskBlockersJSON writes the blocker diagnostic envelope.
func TaskBlockersJSON(w io.Writer, result core.TaskBlockersResult) error {
	return wire.EncodeJSON(w, wire.ToTaskBlockersEnvelope(result))
}

// TaskBlockersHuman renders a compact explanatory blocker list.
func TaskBlockersHuman(w io.Writer, st Style, result core.TaskBlockersResult) error {
	graphQueryHeader(w, st, result.TaskID, result.Task.Slug, result.State, result.Health, result.Projection)
	if len(result.Blockers) == 0 {
		fmt.Fprintf(w, "%s no blockers\n", st.Green("✔"))
	}
	for _, detail := range result.Blockers {
		name := detail.Task.Slug
		if name == "" {
			name = detail.Blocker.TaskID
		}
		direct := "transitive"
		if detail.Blocker.Direct {
			direct = "direct"
		}
		fmt.Fprintf(w, "%s %s  %s  %s\n", st.Dim("•"), st.Bold(name), detail.Blocker.Reason, direct)
		fmt.Fprintf(w, "  %s\n", st.Dim(strings.Join(detail.Blocker.Path, " -> ")))
	}
	graphDiagnosticsHuman(w, st, result.Problems, result.Legacy)
	return nil
}

// TaskUnblocksJSON writes the downstream-impact envelope.
func TaskUnblocksJSON(w io.Writer, result core.TaskUnblocksResult) error {
	return wire.EncodeJSON(w, wire.ToTaskUnblocksEnvelope(result))
}

// TaskUnblocksHuman renders transitive downstream impact without implying
// counterfactual eligibility.
func TaskUnblocksHuman(w io.Writer, st Style, result core.TaskUnblocksResult) error {
	graphQueryHeader(w, st, result.TaskID, result.Task.Slug, result.State, result.Health, "downstream impact")
	if len(result.Unblocks) == 0 {
		fmt.Fprintf(w, "%s no downstream tasks\n", st.Dim("•"))
	}
	for _, detail := range result.Unblocks {
		name := detail.Task.Slug
		if name == "" {
			name = detail.Impact.TaskID
		}
		direct := "transitive"
		if detail.Impact.Direct {
			direct = "direct"
		}
		fmt.Fprintf(w, "%s %s  %s/%s  %s\n", st.Dim("•"), st.Bold(name), detail.State.Role, detail.State.Gate, direct)
		fmt.Fprintf(w, "  %s\n", st.Dim(strings.Join(detail.Impact.Path, " -> ")))
	}
	graphDiagnosticsHuman(w, st, result.Problems, result.Legacy)
	return nil
}

func graphQueryHeader(w io.Writer, st Style, taskID, slug string, state core.TaskGraphState, health core.GraphHealth, projection string) {
	name := slug
	if name == "" {
		name = taskID
	}
	fmt.Fprintf(w, "%s  %s\n", st.Bold(name), st.Dim("("+taskID+")"))
	fmt.Fprintf(w, "%s  %s\n", st.Dim("graph:"), health)
	fmt.Fprintf(w, "%s  %s/%s  eligible=%t\n", st.Dim("state:"), state.Role, state.Gate, state.Eligible)
	fmt.Fprintf(w, "%s  %s\n", st.Dim("view:"), projection)
}

func graphDiagnosticsHuman(w io.Writer, st Style, problems []core.GraphProblem, legacy []core.LegacyDependencyDiagnostic) {
	seen := make(map[string]bool, len(problems))
	for _, problem := range problems {
		key := string(problem.Code) + "\x00" + problem.Message
		if seen[key] {
			continue
		}
		seen[key] = true
		fmt.Fprintf(w, "%s %s: %s\n", st.Warn("⚠"), problem.Code, problem.Message)
	}
	for _, diagnostic := range legacy {
		remedy := "run task depend migrate"
		if !diagnostic.MigrationReady() {
			remedy = "run task depend repair, then task depend migrate"
		}
		fmt.Fprintf(w, "%s legacy %s on %s; %s\n", st.Warn("⚠"), diagnostic.Field, diagnostic.TaskID, remedy)
	}
}
