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
