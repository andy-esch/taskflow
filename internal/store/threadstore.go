package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

var _ core.ThreadStore = (*FS)(nil)

func (s *FS) ListThreads() ([]domain.Thread, []domain.FileProblem, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return nil, nil, err
	}
	return scanDir(s.threadsDir, func(path string, content []byte) (domain.Thread, error) {
		return parseThread(content, path)
	})
}

func (s *FS) GetThread(ref string) (domain.Thread, string, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return domain.Thread{}, "", err
	}
	path, err := s.resolveThread(ref)
	if err != nil {
		return domain.Thread{}, "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Thread{}, "", fmt.Errorf("read Thread %s: %w", path, err)
	}
	thread, err := parseThread(content, path)
	if err != nil {
		return domain.Thread{}, "", fmt.Errorf("%s: %w", path, err)
	}
	_, body := splitFrontmatter(content)
	return thread, string(body), nil
}

func (s *FS) threadCandidates() ([]candidate, error) { return flatCandidates(s.threadsDir) }

func (s *FS) resolveThread(ref string) (string, error) {
	candidates, err := s.threadCandidates()
	if err != nil {
		return "", err
	}
	match, err := resolveID("Thread", ref, candidates)
	if err != nil {
		return "", err
	}
	return match.path, nil
}

func parseThread(content []byte, path string) (domain.Thread, error) {
	base := filepath.Base(path)
	filenameID, slug, ok := splitFlatName(strings.TrimSuffix(base, ".md"))
	if !ok {
		reason, kind := entityNameProblem(base)
		return domain.Thread{}, fmt.Errorf("%w: %q %s", kind, base, reason)
	}
	fm, _, err := splitFrontmatterStrict(content)
	if err != nil {
		return domain.Thread{}, err
	}
	if fm == nil {
		return domain.Thread{}, missingFrontmatterErr("Thread", "id, status, description, goal, created, tasks; see `tskflwctl schema thread`")
	}
	var thread domain.Thread
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &thread); err != nil {
			return domain.Thread{}, fmt.Errorf("%w: %s", errBadFrontmatter, frontmatterError(fm, err))
		}
	}
	thread.Slug = slug
	thread.FilenameID = filenameID
	thread.Path = path
	thread.SourceVersion = hashContent(content)
	return thread, nil
}
