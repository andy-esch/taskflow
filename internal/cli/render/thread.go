package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

func ThreadsJSON(w io.Writer, list core.ThreadListView, problems []domain.FileProblem) error {
	return wire.EncodeJSON(w, wire.ToThreadsEnvelope(list, problems))
}

func ThreadShowJSON(w io.Writer, view core.ThreadView, body string) error {
	return wire.EncodeJSON(w, wire.ToThreadShowEnvelope(view, body))
}

func ThreadFrontierJSON(w io.Writer, view core.ThreadView) error {
	return wire.EncodeJSON(w, wire.ToThreadFrontierEnvelope(view))
}

func ThreadGraphJSON(w io.Writer, projection core.ThreadGraphProjection) error {
	return wire.EncodeJSON(w, wire.ToThreadGraphEnvelope(projection))
}

func ThreadPlanJSON(w io.Writer, projection core.ThreadGraphProjection) error {
	return wire.EncodeJSON(w, wire.ToThreadPlanEnvelope(projection))
}

func ThreadMutationJSON(w io.Writer, receipt core.ThreadCreationReceipt, path string, workspace wire.WorkspaceJSON) error {
	return wire.EncodeJSON(w, wire.ToThreadMutationEnvelope(receipt, path, workspace))
}

func ThreadUpdateJSON(w io.Writer, receipt core.ThreadMutationReceipt, path string, workspace wire.WorkspaceJSON) error {
	return wire.EncodeJSON(w, wire.ToThreadUpdateEnvelope(receipt, path, workspace))
}

func ThreadApplyComposeJSON(w io.Writer, plan core.ThreadApplyPlan, planPath string, dryRun bool, workspace wire.WorkspaceJSON) error {
	return wire.EncodeJSON(w, wire.ToThreadApplyComposeEnvelope(plan, planPath, dryRun, workspace))
}

func ThreadApplyJSON(w io.Writer, receipt core.ThreadApplyReceipt, planPath string, workspace wire.WorkspaceJSON) error {
	return wire.EncodeJSON(w, wire.ToThreadApplyEnvelope(receipt, planPath, workspace))
}

func ThreadApplyHuman(w io.Writer, st Style, receipt core.ThreadApplyReceipt) {
	for _, operation := range receipt.Operations {
		marker := st.Dim("•")
		if operation.Kind == "dependency" {
			verb := "skipped"
			suffix := " (already present)"
			switch operation.State {
			case core.ThreadApplyApplied:
				marker, verb, suffix = st.Green("✔"), "added", ""
			case core.ThreadApplyPending:
				verb, suffix = "pending", ""
				if receipt.DryRun {
					marker, verb = st.Dim("◇"), "would add"
				}
			}
			fmt.Fprintf(w, "%s %s dependency %s -> %s%s\n", marker, verb, operation.PrerequisiteID, operation.DependentID, suffix)
			continue
		}
		verb := "skipped"
		suffix := " (already exists identically)"
		switch operation.State {
		case core.ThreadApplyApplied:
			marker, verb, suffix = st.Green("✔"), "created", ""
		case core.ThreadApplyPending:
			verb, suffix = "pending", ""
			if receipt.DryRun {
				marker, verb = st.Dim("◇"), "would create"
			}
		}
		fmt.Fprintf(w, "%s %s Thread %s%s\n", marker, verb, operation.ThreadID, suffix)
	}
	if receipt.Complete {
		fmt.Fprintf(w, "%s Thread apply complete\n", st.Green("✔"))
	}
}

func ThreadsHuman(w io.Writer, st Style, list core.ThreadListView) error {
	rows := make([][]string, 0, len(list.Threads))
	for _, view := range list.Threads {
		status := string(view.Thread.Status)
		if view.Inconsistent {
			status += " ⚠"
		}
		rows = append(rows, []string{
			status,
			st.Bold(view.Thread.Slug),
			fmt.Sprintf("%d/%d done", view.Rollup.Done, view.Rollup.Total),
			fmt.Sprintf("%d/%d drained", view.Rollup.Drained, view.Rollup.Total),
			fmt.Sprintf("%d eligible", len(view.Frontier)),
			string(view.GraphHealth) + "/" + string(view.ProjectionHealth),
			view.Thread.Description,
		})
	}
	if len(rows) > 0 {
		writeTable(w, st.width, []string{
			st.Dim("STATUS"), st.Dim("THREAD"), st.Dim("PROGRESS"), st.Dim("SOUND"),
			st.Dim("FRONTIER"), st.Dim("GRAPH/VIEW"), st.Dim("DESCRIPTION"),
		}, rows)
	}
	for _, view := range list.Threads {
		if len(view.Problems) == 0 && len(view.GraphProblems) == 0 {
			continue
		}
		fmt.Fprintf(w, "\n%s  %s\n", st.Bold("Diagnostics"), view.Thread.Slug)
		threadDiagnosticsHuman(w, st, view)
	}
	if len(list.GraphProblems) > 0 {
		fmt.Fprintf(w, "\n%s  repository graph\n", st.Bold("Diagnostics"))
		graphDiagnosticsHuman(w, st, list.GraphProblems, nil)
	}
	return nil
}

func ThreadShowHuman(w io.Writer, st Style, view core.ThreadView, body string) error {
	fmt.Fprintf(w, "%s  %s\n", st.Bold(view.Thread.Slug), st.Dim("("+view.Thread.ID+")"))
	fmt.Fprintf(w, "%s  %s", st.Dim("status:"), view.Thread.Status)
	if view.Inconsistent {
		fmt.Fprint(w, "  "+st.Warn("⚠ inconsistent"))
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s  %d/%d done · %d/%d drained · %d deprecated\n",
		st.Dim("progress:"), view.Rollup.Done, view.Rollup.Total, view.Rollup.Drained, view.Rollup.Total, view.Rollup.Deprecated)
	fmt.Fprintf(w, "%s  %s · projection %s · %d frontier · %d external gate(s)\n",
		st.Dim("graph:"), view.GraphHealth, view.ProjectionHealth, len(view.Frontier), len(view.ExternalGates))
	fmt.Fprintf(w, "%s  %s\n", st.Dim("goal:"), view.Thread.Goal)
	if view.Thread.TargetDate != "" {
		fmt.Fprintf(w, "%s  %s\n", st.Dim("target:"), view.Thread.TargetDate)
	}
	if len(view.Members) > 0 {
		fmt.Fprintln(w, "\n"+st.Bold("Members"))
		for _, member := range view.Members {
			name := member.Task.Slug
			if name == "" {
				name = member.State.TaskID
			}
			fmt.Fprintf(w, "%s %s  %s/%s\n", st.Dim("•"), name, member.State.Role, member.State.Gate)
		}
	}
	if len(view.ExternalGates) > 0 {
		fmt.Fprintln(w, "\n"+st.Bold("External gates"))
		for _, gate := range view.ExternalGates {
			name := gate.Task.Slug
			if name == "" {
				name = gate.State.TaskID
			}
			state := "satisfied"
			if gate.Outstanding {
				state = "outstanding"
			}
			fmt.Fprintf(w, "%s %s  %s  %s/%s\n", st.Dim("•"), name, state, gate.State.Role, gate.State.Gate)
		}
	}
	threadDiagnosticsHuman(w, st, view)
	if strings.TrimSpace(body) != "" {
		fmt.Fprintln(w, "\n"+body)
	}
	return nil
}

func ThreadFrontierHuman(w io.Writer, st Style, view core.ThreadView) error {
	fmt.Fprintf(w, "%s  %s\n", st.Bold(view.Thread.Slug), st.Dim("("+view.Thread.ID+")"))
	fmt.Fprintf(w, "%s  %s · projection %s · %d eligible member(s)\n",
		st.Dim("graph:"), view.GraphHealth, view.ProjectionHealth, len(view.Frontier))
	for _, member := range view.Frontier {
		fmt.Fprintf(w, "%s %s  %s\n", st.Green("✔"), st.Bold(member.Task.Slug), member.Task.ID)
	}
	if len(view.Frontier) == 0 {
		fmt.Fprintf(w, "%s no dispatchable member tasks\n", st.Dim("•"))
	}
	threadDiagnosticsHuman(w, st, view)
	return nil
}

func ThreadPlanHuman(w io.Writer, st Style, projection core.ThreadGraphProjection) error {
	view := projection.View
	fmt.Fprintf(w, "%s  %s\n", st.Bold(view.Thread.Slug), st.Dim("("+view.Thread.ID+")"))
	topology := "partial"
	if projection.TopologyComplete {
		topology = "complete"
	}
	fmt.Fprintf(w, "%s  %s · projection %s · topology %s\n",
		st.Dim("graph:"), view.GraphHealth, view.ProjectionHealth, topology)

	byID := make(map[string]core.ThreadGraphNode, len(projection.Nodes))
	for _, node := range projection.Nodes {
		byID[node.TaskID] = node
	}
	if len(view.ExternalGates) > 0 {
		fmt.Fprintln(w, "\n"+st.Bold("External gates"))
		for _, gate := range view.ExternalGates {
			node := byID[gate.State.TaskID]
			state := "satisfied"
			if gate.Outstanding {
				state = "outstanding"
			}
			fmt.Fprintf(w, "%s %s  %s  %s\n", st.Dim("•"), node.Label, node.TaskID, state)
		}
	}

	ranked := make(map[string]bool, len(view.Members))
	for _, wave := range projection.Waves {
		fmt.Fprintf(w, "\n%s\n", st.Bold(fmt.Sprintf("Wave %d", wave.Index)))
		for _, taskID := range wave.TaskIDs {
			node := byID[taskID]
			ranked[taskID] = true
			fmt.Fprintf(w, "%s %s  %s  %s/%s\n", st.Dim("•"), node.Label, taskID, node.State.Role, node.State.Gate)
		}
	}
	unranked := make([]core.ThreadGraphNode, 0)
	for _, node := range projection.Nodes {
		if node.Role == core.ThreadTaskMember && !ranked[node.TaskID] {
			unranked = append(unranked, node)
		}
	}
	if len(unranked) > 0 {
		fmt.Fprintln(w, "\n"+st.Bold("Unranked members"))
		for _, node := range unranked {
			fmt.Fprintf(w, "%s %s  %s  %s/%s\n", st.Warn("⚠"), node.Label, node.TaskID, node.State.Role, node.State.Gate)
		}
	}
	if len(projection.Waves) == 0 && len(unranked) == 0 {
		fmt.Fprintf(w, "\n%s no member tasks to rank\n", st.Dim("•"))
	}
	threadDiagnosticsHuman(w, st, view)
	return nil
}

func ThreadCreatedHuman(w io.Writer, st Style, receipt core.ThreadCreationReceipt, path string) {
	verb, marker := "created", "✔"
	if receipt.DryRun {
		verb, marker = "would create", "◇"
	}
	fmt.Fprintf(w, "%s %s Thread %s at %s\n", st.Green(marker), verb, st.Bold(receipt.Thread.ID), path)
}

func ThreadMutationHuman(w io.Writer, st Style, receipt core.ThreadMutationReceipt, path string) {
	marker, verb := "✔", "updated"
	if receipt.DryRun {
		marker, verb = "◇", "would update"
	} else if !receipt.Changed {
		marker, verb = "•", "already satisfied"
	}
	fmt.Fprintf(w, "%s %s Thread %s (%s)\n", st.Green(marker), verb, st.Bold(receipt.Thread.Slug), receipt.Operation)
	for _, outcome := range receipt.MemberOutcomes {
		prefix := st.Green("✔")
		if receipt.DryRun {
			prefix = st.Dim("◇")
		} else if outcome.Outcome == "skipped" {
			prefix = st.Dim("•")
		}
		fmt.Fprintf(w, "  %s %s %s (%s)\n", prefix, outcome.Action, outcome.TaskID, outcome.Outcome)
	}
	if receipt.Before.Thread.Status != receipt.After.Thread.Status {
		fmt.Fprintf(w, "  %s status %s -> %s\n", st.Dim("•"), receipt.Before.Thread.Status, receipt.After.Thread.Status)
	}
	if receipt.Changed {
		fmt.Fprintf(w, "  %s members %d -> %d · frontier %d -> %d · drained %d/%d -> %d/%d\n",
			st.Dim("projection:"), len(receipt.Before.Members), len(receipt.After.Members),
			len(receipt.Before.Frontier), len(receipt.After.Frontier),
			receipt.Before.Rollup.Drained, receipt.Before.Rollup.Total,
			receipt.After.Rollup.Drained, receipt.After.Rollup.Total)
	}
	if receipt.Remedy != "" {
		fmt.Fprintf(w, "  %s %s\n", st.Dim("remedy:"), receipt.Remedy)
	}
	if path != "" {
		fmt.Fprintf(w, "  %s %s\n", st.Dim("path:"), path)
	}
}

func threadDiagnosticsHuman(w io.Writer, st Style, view core.ThreadView) {
	for _, problem := range view.Problems {
		fmt.Fprintf(w, "%s %s: %s\n", st.Warn("⚠"), problem.Code, problem.Message)
	}
	graphDiagnosticsHuman(w, st, view.GraphProblems, nil)
}
