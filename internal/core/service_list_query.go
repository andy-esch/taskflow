package core

import (
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// Query methods compose existing entity filters with the common query boundary.
// They keep the full diagnostic set, even for an empty or one-row page; pagination
// never turns a partial adapter read into a healthy result.
func (s *Service) QueryTasks(f TaskFilter, q ListQuery) (ListResult[LoadedRecord[domain.Task]], []LoadProblem, error) {
	if err := validateListQuery(q, true); err != nil {
		return ListResult[LoadedRecord[domain.Task]]{}, nil, err
	}
	items, problems, err := s.ListTasks(f)
	if err != nil {
		return ListResult[LoadedRecord[domain.Task]]{}, problems, err
	}
	result, err := queryList(items, q, "id", struct {
		Kind   string
		Filter TaskFilter
	}{"tasks", f}, canonicalLoadProblems(problems),
		func(r LoadedRecord[domain.Task]) listRow {
			t := r.Value
			return listRow{r.Source.ID, t.Slug, t.Created, t.Updated, strings.Join([]string{t.Slug, t.Title, t.Description, strings.Join(t.Tags, " ")}, "\n")}
		})
	return result, problems, err
}

func (s *Service) QueryEpics(status string, q ListQuery) (ListResult[EpicSummary], []LoadProblem, error) {
	if err := validateListQuery(q, true); err != nil {
		return ListResult[EpicSummary]{}, nil, err
	}
	if status != "" {
		if err := domain.ValidateEpicStatus(status); err != nil {
			return ListResult[EpicSummary]{}, nil, err
		}
	}
	items, problems, err := s.ListEpics()
	if err != nil {
		return ListResult[EpicSummary]{}, problems, err
	}
	filtered := make([]EpicSummary, 0, len(items))
	for _, r := range items {
		if status == "" || r.Epic.Status == status {
			filtered = append(filtered, r)
		}
	}
	result, err := queryList(filtered, q, "id", struct {
		Kind, Status string
	}{"epics", status}, canonicalLoadProblems(problems),
		func(r EpicSummary) listRow {
			e := r.Epic
			return listRow{r.Source.ID, e.ID, e.Created, e.Updated, strings.Join([]string{e.ID, e.Description, strings.Join(e.Tags, " ")}, "\n")}
		})
	return result, problems, err
}

func (s *Service) QueryAudits(bucket string, all bool, q ListQuery) (ListResult[LoadedRecord[domain.Audit]], []LoadProblem, error) {
	if err := validateListQuery(q, true); err != nil {
		return ListResult[LoadedRecord[domain.Audit]]{}, nil, err
	}
	items, problems, err := s.ListAudits(bucket, all)
	if err != nil {
		return ListResult[LoadedRecord[domain.Audit]]{}, problems, err
	}
	result, err := queryList(items, q, "id", struct {
		Kind, Bucket string
		All          bool
	}{"audits", bucket, all}, canonicalLoadProblems(problems),
		func(r LoadedRecord[domain.Audit]) listRow {
			a := r.Value
			return listRow{r.Source.ID, a.Slug, a.Date, a.Updated, a.Slug + "\n" + a.Area}
		})
	return result, problems, err
}

func (s *Service) QueryResearch(tag string, q ListQuery) (ListResult[LoadedRecord[domain.Research]], []LoadProblem, error) {
	if err := validateListQuery(q, true); err != nil {
		return ListResult[LoadedRecord[domain.Research]]{}, nil, err
	}
	items, problems, err := s.ListResearch(tag)
	if err != nil {
		return ListResult[LoadedRecord[domain.Research]]{}, problems, err
	}
	result, err := queryList(items, q, "-created", struct {
		Kind, Tag string
	}{"research", tag}, canonicalLoadProblems(problems),
		func(r LoadedRecord[domain.Research]) listRow {
			d := r.Value
			return listRow{r.Source.ID, d.Slug, d.Created, d.Updated, strings.Join([]string{d.Slug, d.Description, strings.Join(d.Tags, " ")}, "\n")}
		})
	return result, problems, err
}

func (s *Service) QueryThreadViews(status string, q ListQuery) (ThreadListView, *ListPage, []ThreadReadProblem, error) {
	if err := validateListQuery(q, true); err != nil {
		return ThreadListView{}, nil, nil, err
	}
	if status != "" {
		if err := domain.ValidateThreadStatus(domain.ThreadStatus(status)); err != nil {
			return ThreadListView{}, nil, nil, err
		}
	}
	list, problems, err := s.ListThreadViews()
	if err != nil {
		return list, nil, problems, err
	}
	filtered := make([]ThreadView, 0, len(list.Threads))
	for _, r := range list.Threads {
		if status == "" || string(r.Thread.Status) == status {
			filtered = append(filtered, r)
		}
	}
	result, err := queryList(filtered, q, "id", struct {
		Kind, Status string
	}{"threads", status}, struct {
		Problems      []ThreadReadProblem
		GraphHealth   GraphHealth
		GraphProblems []GraphProblem
	}{problems, list.GraphHealth, list.GraphProblems},
		func(r ThreadView) listRow {
			t := r.Thread
			return listRow{r.Source.ID, t.Slug, t.Created, t.Updated, strings.Join([]string{t.Slug, t.Description, t.Goal, strings.Join(t.Tags, " ")}, "\n")}
		})
	list.Threads = result.Items
	return list, result.Page, problems, err
}

func (s *Service) QueryFindingPage(f FindingFilter, q ListQuery) (ListResult[AuditFinding], []LoadProblem, error) {
	if err := validateListQuery(q, false); err != nil {
		return ListResult[AuditFinding]{}, nil, err
	}
	items, problems, err := s.QueryFindings(f)
	if err != nil {
		return ListResult[AuditFinding]{}, problems, err
	}
	result, err := queryList(items, q, "id", struct {
		Kind   string
		Filter FindingFilter
	}{"findings", f}, canonicalLoadProblems(problems),
		func(r AuditFinding) listRow {
			return listRow{id: r.AuditID + "." + r.Code, name: r.Title, text: strings.Join([]string{r.Audit, r.Code, r.Title, r.Component, r.File, r.Note}, "\n")}
		})
	return result, problems, err
}
