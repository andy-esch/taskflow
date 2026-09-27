package store

import (
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// These projections are the filesystem adapter's only ordinary-read translation
// boundary. Core receives canonical identity and opaque location explicitly and
// never has to parse a path or inspect filename-derived domain fields.
func taskRecord(task domain.Task) core.LoadedRecord[domain.Task] {
	return core.LoadedRecord[domain.Task]{
		Value:  task,
		Source: core.RecordSource{ID: task.FilenameID, Location: task.Path},
	}
}

func taskBodyRecord(record core.TaskWithBody) core.LoadedRecord[core.TaskWithBody] {
	return core.LoadedRecord[core.TaskWithBody]{
		Value:  record,
		Source: core.RecordSource{ID: record.Task.FilenameID, Location: record.Task.Path},
	}
}

func epicRecord(epic domain.Epic) core.LoadedRecord[domain.Epic] {
	return core.LoadedRecord[domain.Epic]{
		Value:  epic,
		Source: core.RecordSource{ID: epic.ID, Location: epic.Path},
	}
}

func auditRecord(audit domain.Audit) core.LoadedRecord[domain.Audit] {
	return core.LoadedRecord[domain.Audit]{
		Value:  audit,
		Source: core.RecordSource{ID: audit.FilenameID, Location: audit.Path},
	}
}

func researchRecord(research domain.Research) core.LoadedRecord[domain.Research] {
	return core.LoadedRecord[domain.Research]{
		Value:  research,
		Source: core.RecordSource{ID: research.FilenameID, Location: research.Path},
	}
}

func loadedProblems(kind core.EntityKind, problems []domain.FileProblem) []core.LoadProblem {
	out := make([]core.LoadProblem, 0, len(problems))
	for _, problem := range problems {
		out = append(out, core.LoadProblem{
			EntityKind: kind, EntityID: problem.EntityID, EntitySlug: problem.EntitySlug,
			Location: problem.Path, LocalPath: problem.Path, Message: problem.Message,
		})
	}
	return out
}

func (s *FS) ReadTasks() (core.TaskRead, error) {
	read, err := s.ReadTaskGraph()
	if err != nil {
		return core.TaskRead{}, err
	}
	return core.TaskRead{Records: taskGraphRecords(read), Problems: taskGraphProblems(read.Problems)}, nil
}

func (s *FS) ReadTask(ref string) (core.LoadedRecord[core.TaskWithBody], error) {
	task, body, err := s.GetTask(ref)
	if err != nil {
		return core.LoadedRecord[core.TaskWithBody]{}, err
	}
	return taskBodyRecord(core.TaskWithBody{Task: task, Body: body}), nil
}

func (s *FS) ReadEpics() (core.EpicRead, error) {
	epics, problems, err := s.ListEpics()
	if err != nil {
		return core.EpicRead{}, err
	}
	records := make([]core.LoadedRecord[domain.Epic], 0, len(epics))
	for _, epic := range epics {
		records = append(records, epicRecord(epic))
	}
	return core.EpicRead{Records: records, Problems: loadedProblems(core.EntityEpic, problems)}, nil
}

func (s *FS) ReadEpic(ref string) (core.LoadedRecord[core.EpicWithBody], error) {
	epic, body, err := s.GetEpic(ref)
	if err != nil {
		return core.LoadedRecord[core.EpicWithBody]{}, err
	}
	return core.LoadedRecord[core.EpicWithBody]{
		Value:  core.EpicWithBody{Epic: epic, Body: body},
		Source: core.RecordSource{ID: epic.ID, Location: epic.Path},
	}, nil
}

func (s *FS) ReadAudits() (core.AuditRead, error) {
	audits, problems, err := s.ListAudits()
	if err != nil {
		return core.AuditRead{}, err
	}
	records := make([]core.LoadedRecord[domain.Audit], 0, len(audits))
	for _, audit := range audits {
		records = append(records, auditRecord(audit))
	}
	return core.AuditRead{Records: records, Problems: loadedProblems(core.EntityAudit, problems)}, nil
}

func (s *FS) ReadAudit(ref string) (core.LoadedRecord[core.AuditWithBody], error) {
	audit, body, err := s.GetAudit(ref)
	if err != nil {
		return core.LoadedRecord[core.AuditWithBody]{}, err
	}
	return core.LoadedRecord[core.AuditWithBody]{
		Value:  core.AuditWithBody{Audit: audit, Body: body},
		Source: core.RecordSource{ID: audit.FilenameID, Location: audit.Path},
	}, nil
}

func (s *FS) ReadResearch() (core.ResearchRead, error) {
	docs, problems, err := s.ListResearch()
	if err != nil {
		return core.ResearchRead{}, err
	}
	records := make([]core.LoadedRecord[domain.Research], 0, len(docs))
	for _, research := range docs {
		records = append(records, researchRecord(research))
	}
	return core.ResearchRead{Records: records, Problems: loadedProblems(core.EntityResearch, problems)}, nil
}

func (s *FS) ReadResearchDocument(ref string) (core.LoadedRecord[core.ResearchWithBody], error) {
	research, body, err := s.GetResearch(ref)
	if err != nil {
		return core.LoadedRecord[core.ResearchWithBody]{}, err
	}
	return core.LoadedRecord[core.ResearchWithBody]{
		Value:  core.ResearchWithBody{Research: research, Body: body},
		Source: core.RecordSource{ID: research.FilenameID, Location: research.Path},
	}, nil
}

func taskGraphRecords(read core.TaskGraphRead) []core.LoadedRecord[domain.Task] {
	if read.Records != nil {
		return append([]core.LoadedRecord[domain.Task](nil), read.Records...)
	}
	records := make([]core.LoadedRecord[domain.Task], 0, len(read.Tasks))
	for _, task := range read.Tasks {
		records = append(records, taskRecord(task))
	}
	return records
}

func taskGraphProblems(problems []core.TaskGraphLoadProblem) []core.LoadProblem {
	out := make([]core.LoadProblem, 0, len(problems))
	for _, problem := range problems {
		localPath := problem.Path
		if localPath == "" && problem.LocationIsPath {
			localPath = problem.Location
		}
		out = append(out, core.LoadProblem{
			EntityKind: core.EntityTask, EntityID: problem.TaskID, EntitySlug: problem.TaskSlug,
			Location: problem.Location, LocalPath: localPath, Message: problem.Message,
		})
	}
	return out
}
