package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/wire"
)

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
		fmt.Fprintf(w, "%s legacy %s on %s; run task depend migrate\n", st.Warn("⚠"), diagnostic.Field, diagnostic.TaskID)
	}
}
