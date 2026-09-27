package store

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// ReadLintTasks adapts the body-carrying local task scan to lint's portable
// failed-record contract without adding another filesystem pass.
func (s *FS) ReadLintTasks() ([]core.LoadedRecord[core.TaskWithBody], []core.LoadProblem, error) {
	records, problems, err := s.ListTasksWithBodies()
	loaded := make([]core.LoadedRecord[core.TaskWithBody], 0, len(records))
	for _, record := range records {
		loaded = append(loaded, taskBodyRecord(record))
	}
	return loaded, loadedProblems(core.EntityTask, problems), err
}

func (s *FS) ReadLintEpics() ([]core.LoadedRecord[domain.Epic], []core.LoadProblem, error) {
	records, problems, err := s.ListEpics()
	loaded := make([]core.LoadedRecord[domain.Epic], 0, len(records))
	for _, record := range records {
		loaded = append(loaded, epicRecord(record))
	}
	return loaded, loadedProblems(core.EntityEpic, problems), err
}

func (s *FS) ReadAuditSnapshot(selector string) (core.AuditSnapshot, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return core.AuditSnapshot{}, err
	}
	if selector != "" {
		path, err := s.resolveAudit(selector)
		if err != nil {
			return core.AuditSnapshot{}, err
		}
		content, err := s.auditReadFile(path)
		if err != nil {
			return core.AuditSnapshot{}, fmt.Errorf("read audit %s: %w", path, err)
		}
		a, findings, nearMisses, candidateIssues, err := parseAuditWithFindings(content, path)
		if err != nil {
			return core.AuditSnapshot{}, fmt.Errorf("%s: %w", path, err)
		}
		record := core.AuditWithFindings{
			Audit: a, Findings: findings, NearMisses: nearMisses, CandidateIssues: candidateIssues,
		}
		return core.AuditSnapshot{Audits: []core.LoadedRecord[core.AuditWithFindings]{{
			Value: record, Source: core.RecordSource{ID: a.FilenameID, Location: a.Path},
		}}}, nil
	}
	records, problems, err := s.ListAuditsWithFindings()
	loaded := make([]core.LoadedRecord[core.AuditWithFindings], 0, len(records))
	for _, record := range records {
		loaded = append(loaded, core.LoadedRecord[core.AuditWithFindings]{
			Value:  record,
			Source: core.RecordSource{ID: record.Audit.FilenameID, Location: record.Audit.Path},
		})
	}
	return core.AuditSnapshot{
		Audits: loaded, Problems: loadedProblems(core.EntityAudit, problems),
	}, err
}

func (s *FS) ReadLintResearch() ([]core.LoadedRecord[domain.Research], []core.LoadProblem, error) {
	records, problems, err := s.ListResearch()
	loaded := make([]core.LoadedRecord[domain.Research], 0, len(records))
	for _, record := range records {
		loaded = append(loaded, researchRecord(record))
	}
	return loaded, loadedProblems(core.EntityResearch, problems), err
}
