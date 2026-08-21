package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/theme"
	"github.com/andy-esch/taskflow/internal/wire"
)

// SummaryHuman renders the at-a-glance dashboard.
func SummaryHuman(w io.Writer, st Style, s core.Summary) error {
	// Status counts — active line, then archived line, only non-zero buckets.
	active, archived := splitCounts(s.Counts)
	fmt.Fprintf(w, "%s\n", st.Bold("Tasks"))
	if line := countLine(st, active); line != "" {
		fmt.Fprintf(w, "  %s  %s\n", st.Dim("active  "), line)
	}
	if line := countLine(st, archived); line != "" {
		fmt.Fprintf(w, "  %s  %s\n", st.Dim("archived"), line)
	}

	if len(s.InProgress) > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Bold(fmt.Sprintf("In progress (%d)", len(s.InProgress))))
		rows := make([][]string, 0, len(s.InProgress))
		for _, t := range s.InProgress {
			rows = append(rows, []string{"  " + st.Bold(t.Slug), st.Dim(theme.RelativeDate(theme.TaskDate(t))), t.Description})
		}
		writeTable(w, st.width, nil, rows)
	}

	if len(s.Epics) > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Bold("Epics"))
		rows := make([][]string, 0, len(s.Epics))
		for _, e := range s.Epics {
			bar := fmt.Sprintf("%s %s", st.Bar(e.Percent(), 10), st.Percent(e.Percent()))
			rows = append(rows, []string{"  " + st.Bold(e.Epic.ID), bar, theme.Counts(e.Done, e.Total), e.Epic.Description})
		}
		writeTable(w, st.width, nil, rows)
	}

	// Only open audits, only when there are any — the actionable subset, rendered
	// with the same bar treatment as epics so the dashboard reads from one vocabulary.
	if len(s.OpenAudits) > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Bold(fmt.Sprintf("Open audits (%d)", len(s.OpenAudits))))
		rows := make([][]string, 0, len(s.OpenAudits))
		for _, a := range s.OpenAudits {
			bar := fmt.Sprintf("%s %s", st.SegmentBar(a.DoneFindings, a.ActiveFindings, a.DroppedFindings, a.Findings, 10), st.AuditPercent(a.Percent()))
			counts := theme.Counts(a.Resolved(), a.Findings)
			if note := auditStateNote(st, a, false); note != "" {
				counts += "  " + note
			}
			rows = append(rows, []string{"  " + st.Bold(a.Slug), bar, counts, a.Area})
		}
		writeTable(w, st.width, nil, rows)
	}

	// Audit findings — the actionable cross-audit inbox, triaged. Same source as the
	// TUI dashboard's widget (core.Summary.Findings), so the two surfaces agree.
	if fr := s.Findings; fr.Open+fr.InProgress > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Bold(fmt.Sprintf("Audit findings (%d open · %d in progress)", fr.Open, fr.InProgress)))
		if line := countByLine(st, fr.ByUrgency); line != "" {
			fmt.Fprintf(w, "  %s  %s\n", st.Dim("by urgency"), line)
		}
		if line := countByLine(st, fr.ByComponent); line != "" {
			fmt.Fprintf(w, "  %s  %s\n", st.Dim("by area  "), line)
		}
	}
	if s.ReadyToClose > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Green(fmt.Sprintf("✓ %d audit(s) ready to close (all findings resolved; `audit close <slug>`)", s.ReadyToClose)))
	}

	if s.RevisitDue > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Warn(fmt.Sprintf("↻ %d deferred due to revisit (snooze date reached; `task ready`/`task next` to resume)", s.RevisitDue)))
	}
	if s.BadEpicStatus > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Warn(fmt.Sprintf("⚠ %d epic(s) with unrecognized status (set active/retired/deprecated; run `lint`)", s.BadEpicStatus)))
	}
	if len(s.Problems) > 0 {
		fmt.Fprintf(w, "\n%s\n", st.Red(fmt.Sprintf("! %d unreadable file(s) (run `lint`)", len(s.Problems))))
	}
	return nil
}

func splitCounts(counts []core.StatusCount) (active, archived []core.StatusCount) {
	for _, c := range counts {
		if c.Status.IsActive() {
			active = append(active, c)
		} else {
			archived = append(archived, c)
		}
	}
	return active, archived
}

// countByLine renders a finding breakdown ("1 acute · 12 soon · 23 eventually"),
// the dim-separated, uncolored counterpart of the dashboard's by-urgency / by-area
// lines. Shares the iterate/format/join STRUCTURE with them via theme.Breakdown;
// only this surface's plain segment format + dim separator differ (audit M10).
func countByLine(st Style, cs []core.CountBy) string {
	return theme.Breakdown(cs, st.Dim(" · "), 0,
		func(c core.CountBy) string { return fmt.Sprintf("%d %s", c.Count, c.Key) }, nil)
}

// countLine renders "3 next-up · 1 in-progress", skipping zero buckets.
func countLine(st Style, counts []core.StatusCount) string {
	var parts []string
	for _, c := range counts {
		if c.Count == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d %s", c.Count, st.Status(c.Status)))
	}
	return strings.Join(parts, st.Dim(" · "))
}

// SummaryJSON writes the dashboard as a versioned envelope.
func SummaryJSON(w io.Writer, s core.Summary) error {
	return wire.EncodeJSON(w, wire.ToSummaryEnvelope(s))
}

// StatusAllHuman renders one compact block per logical planning identity, followed by
// the cross-space working set. Entry-point diagnoses stay inside their group so a dead
// checkout is visible without drowning out summaries loaded through healthy alternates.
func StatusAllHuman(w io.Writer, st Style, overview core.SpaceOverview) error {
	for i, space := range overview.Spaces {
		if i > 0 {
			fmt.Fprintln(w)
		}
		heading := space.ID
		if space.Selected != nil && space.Selected.Label != "" && space.Selected.Label != space.ID {
			heading += "  " + st.Dim(space.Selected.Label)
		}
		fmt.Fprintln(w, st.Bold(heading))
		if space.Summary != nil {
			renderCompactSpaceSummary(w, st, *space.Summary)
		} else {
			fmt.Fprintf(w, "  %s %s\n", st.Red("!"), space.LoadError)
		}
		for _, entry := range space.Entries {
			if entry.Healthy() {
				continue
			}
			fmt.Fprintf(w, "  %s %s  %s — %s\n", st.Red("!"), st.Bold(entry.ID), st.Dim(string(entry.State)), entry.Detail)
			if entry.Remedy != "" {
				fmt.Fprintf(w, "      %s %s\n", st.Dim("remedy:"), entry.Remedy)
			}
		}
	}

	if len(overview.InProgress) == 0 {
		fmt.Fprintf(w, "\n%s\n", st.Dim("No tasks in progress across registered spaces."))
		return nil
	}
	fmt.Fprintf(w, "\n%s\n", st.Bold(fmt.Sprintf("In progress across spaces (%d)", len(overview.InProgress))))
	rows := make([][]string, 0, len(overview.InProgress))
	for _, item := range overview.InProgress {
		rows = append(rows, []string{
			"  " + st.Dim("["+item.SpaceID+"]"),
			st.Bold(item.Task.Slug),
			st.Dim(theme.RelativeDate(theme.TaskDate(item.Task))),
			item.Task.Description,
		})
	}
	writeTable(w, st.width, nil, rows)
	return nil
}

func renderCompactSpaceSummary(w io.Writer, st Style, summary core.Summary) {
	tasks := countLine(st, summary.Counts)
	if tasks == "" {
		tasks = st.Dim("none")
	}
	fmt.Fprintf(w, "  %s  %s\n", st.Dim("tasks "), tasks)

	var details []string
	if len(summary.Epics) > 0 {
		done, total := 0, 0
		for _, epic := range summary.Epics {
			done += epic.Done
			total += epic.Total
		}
		details = append(details, fmt.Sprintf("%s · %d/%d member tasks done", plural(len(summary.Epics), "epic"), done, total))
	}
	if len(summary.OpenAudits) > 0 {
		details = append(details, plural(len(summary.OpenAudits), "open audit"))
	}
	if findings := summary.Findings.Open + summary.Findings.InProgress; findings > 0 {
		details = append(details, plural(findings, "actionable finding"))
	}
	if len(details) > 0 {
		fmt.Fprintf(w, "  %s  %s\n", st.Dim("other "), strings.Join(details, st.Dim(" · ")))
	}

	var warnings []string
	if summary.ReadyToClose > 0 {
		warnings = append(warnings, plural(summary.ReadyToClose, "audit")+" ready to close")
	}
	if summary.RevisitDue > 0 {
		warnings = append(warnings, plural(summary.RevisitDue, "deferred task")+" due to revisit")
	}
	if summary.BadEpicStatus > 0 {
		warnings = append(warnings, plural(summary.BadEpicStatus, "epic")+" with unrecognized status")
	}
	if len(summary.Problems) > 0 {
		warnings = append(warnings, plural(len(summary.Problems), "unreadable file"))
	}
	if len(warnings) > 0 {
		fmt.Fprintf(w, "  %s %s\n", st.Warn("!"), strings.Join(warnings, st.Dim(" · ")))
	}
}

// StatusAllJSON emits the shared cross-space status envelope.
func StatusAllJSON(w io.Writer, overview core.SpaceOverview) error {
	return wire.EncodeJSON(w, wire.ToStatusAllEnvelope(overview))
}
