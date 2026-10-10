package render

import (
	"io"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TasksPageJSON(w io.Writer, items []core.LoadedRecord[domain.Task], problems []core.LoadProblem, page *core.ListPage) error {
	envelope := wire.ToTasksEnvelope(items, problems)
	envelope.Page = wire.ToListPageJSON(page)
	return wire.EncodeJSON(w, envelope)
}

func EpicsPageJSON(w io.Writer, items []core.EpicSummary, problems []core.LoadProblem, page *core.ListPage) error {
	envelope := wire.ToEpicsEnvelope(items, problems)
	envelope.Page = wire.ToListPageJSON(page)
	return wire.EncodeJSON(w, envelope)
}

func AuditsPageJSON(w io.Writer, items []core.LoadedRecord[domain.Audit], problems []core.LoadProblem, page *core.ListPage) error {
	envelope := wire.ToAuditsEnvelope(items, problems)
	envelope.Page = wire.ToListPageJSON(page)
	return wire.EncodeJSON(w, envelope)
}

func ResearchPageJSON(w io.Writer, items []core.LoadedRecord[domain.Research], problems []core.LoadProblem, page *core.ListPage) error {
	envelope := wire.ToResearchListEnvelope(items, problems)
	envelope.Page = wire.ToListPageJSON(page)
	return wire.EncodeJSON(w, envelope)
}

func FindingsPageJSON(w io.Writer, items []core.AuditFinding, problems []core.LoadProblem, page *core.ListPage) error {
	envelope := wire.ToFindingsEnvelope(items, problems)
	envelope.Page = wire.ToListPageJSON(page)
	return wire.EncodeJSON(w, envelope)
}

func ThreadsPageJSON(w io.Writer, list core.ThreadListView, problems []core.ThreadReadProblem, page *core.ListPage) error {
	envelope := wire.ToThreadsEnvelope(list, problems)
	envelope.Page = wire.ToListPageJSON(page)
	return wire.EncodeJSON(w, envelope)
}

// ThreadsProjectedPageJSON preserves repository graph context independently of
// selected columns and matching rows. A nil page retains the legacy projection.
func ThreadsProjectedPageJSON(w io.Writer, cols []Column[core.ThreadView], list core.ThreadListView, problems []core.ThreadReadProblem, page *core.ListPage) error {
	var context []orderedField
	if page != nil {
		context = []orderedField{
			{"graph_health", string(list.GraphHealth)},
			{"graph_problems", wire.ToGraphProblemsJSON(list.GraphProblems)},
		}
	}
	return projectedPageJSONWithContext(w, "threads", cols, list.Threads, wire.ToThreadReadProblemsJSON(problems), page, context)
}
