package store

import "github.com/andy-esch/taskflow/internal/core"

// Path resolvers: slug/id → absolute file path WITHOUT reading or parsing the file.
// The `<entity> path` commands use these so they still work on a file whose
// frontmatter won't parse — the case where you most need the path (to open and
// repair it). GetTask/GetEpic/GetAudit would fail at the parse step first.

// ResolveTaskPath returns a task's file path from its slug/id, parse-free.
func (s *FS) ResolveTaskPath(slug string) (string, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return "", err
	}
	return s.resolve(slug)
}

// ResolveEpicPath returns an epic's file path from its id, parse-free.
func (s *FS) ResolveEpicPath(id string) (string, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return "", err
	}
	return s.resolveEpicPath(id)
}

// ResolveAuditPath returns an audit's file path from its slug/id, parse-free.
func (s *FS) ResolveAuditPath(slug string) (string, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return "", err
	}
	return s.resolveAudit(slug)
}

// ResolveResearchPath returns a research doc's file path from its slug/id, parse-free.
func (s *FS) ResolveResearchPath(slug string) (string, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return "", err
	}
	return s.resolveResearch(slug)
}

// ResolveThreadPath returns a Thread's path without parsing its frontmatter.
func (s *FS) ResolveThreadPath(ref string) (string, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return "", err
	}
	return s.resolveThread(ref)
}

// Selected local navigation is parse-free but cannot fall back to a human label.
// Keep exact lookup separate from the ordinary CLI reference resolver above.
func (s *FS) ResolveTaskPathByID(id string) (string, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return "", err
	}
	return s.resolveTaskMutationTarget(core.TaskMutationTarget{ID: id})
}

func (s *FS) ResolveEpicPathByID(id string) (string, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return "", err
	}
	return s.resolveEpicMutationPath(id)
}

func (s *FS) ResolveAuditPathByID(id string) (string, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return "", err
	}
	return s.resolveAuditMutationTarget(core.AuditMutationTarget{ID: id})
}

func (s *FS) ResolveResearchPathByID(id string) (string, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return "", err
	}
	return s.resolveResearchMutationTarget(core.ResearchMutationTarget{ID: id})
}

func (s *FS) ResolveThreadPathByID(id string) (string, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return "", err
	}
	return s.resolveThreadPathExact(id)
}

func (s *FS) resolveThreadPathExact(id string) (string, error) {
	if err := validEntityID(id); err != nil {
		return "", err
	}
	candidates, err := s.threadCandidates()
	if err != nil {
		return "", err
	}
	selected, err := resolveExactID(candidates, id)
	if err != nil {
		return "", err
	}
	return selected.path, nil
}
