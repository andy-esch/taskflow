package store

import (
	"path/filepath"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// These projections are the filesystem adapter's only ordinary-read translation
// boundary. Core receives canonical identity and opaque location explicitly and
// never has to parse a path or inspect filename-derived domain fields.
func taskSource(path string) core.RecordSource {
	id, _, _ := splitFlatName(strings.TrimSuffix(filepath.Base(path), ".md"))
	return core.RecordSource{ID: id, Location: path, LocationIsPath: true}
}

func epicRecord(epic domain.Epic, path string) core.LoadedRecord[domain.Epic] {
	return core.LoadedRecord[domain.Epic]{
		Value:  epic,
		Source: core.RecordSource{ID: epic.ID, Location: path, LocationIsPath: true},
	}
}

func auditSource(path string) core.RecordSource {
	id, _, _ := splitFlatName(strings.TrimSuffix(filepath.Base(path), ".md"))
	return core.RecordSource{ID: id, Location: path, LocationIsPath: true}
}

func auditRecord(audit domain.Audit, path string) core.LoadedRecord[domain.Audit] {
	return core.LoadedRecord[domain.Audit]{
		Value:  audit,
		Source: auditSource(path),
	}
}

func researchSource(path string) core.RecordSource {
	id, _, _ := splitFlatName(strings.TrimSuffix(filepath.Base(path), ".md"))
	return core.RecordSource{ID: id, Location: path, LocationIsPath: true}
}

func researchRecord(research domain.Research, path string) core.LoadedRecord[domain.Research] {
	return core.LoadedRecord[domain.Research]{
		Value:  research,
		Source: researchSource(path),
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
	return s.readTask(ref)
}

func (s *FS) ReadEpics() (core.EpicRead, error) {
	records, problems, err := s.scanEpics()
	if err != nil {
		return core.EpicRead{}, err
	}
	return core.EpicRead{Records: records, Problems: loadedProblems(core.EntityEpic, problems)}, nil
}

func (s *FS) ReadEpic(ref string) (core.LoadedRecord[core.EpicWithBody], error) {
	return s.readEpic(ref)
}

func (s *FS) ReadAudits() (core.AuditRead, error) {
	records, problems, err := s.scanAudits()
	if err != nil {
		return core.AuditRead{}, err
	}
	return core.AuditRead{Records: records, Problems: loadedProblems(core.EntityAudit, problems)}, nil
}

func (s *FS) ReadAudit(ref string) (core.LoadedRecord[core.AuditWithBody], error) {
	return s.readAudit(ref)
}

func (s *FS) ReadResearch() (core.ResearchRead, error) {
	records, problems, err := s.scanResearch()
	if err != nil {
		return core.ResearchRead{}, err
	}
	return core.ResearchRead{Records: records, Problems: loadedProblems(core.EntityResearch, problems)}, nil
}

func (s *FS) ReadResearchDocument(ref string) (core.LoadedRecord[core.ResearchWithBody], error) {
	return s.readResearch(ref)
}

func taskGraphRecords(read core.TaskGraphRead) []core.LoadedRecord[domain.Task] {
	records := make([]core.LoadedRecord[domain.Task], 0, len(read.GuardedRecords))
	for _, guarded := range read.GuardedRecords {
		records = append(records, guarded.Record)
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
