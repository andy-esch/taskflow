package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
)

// ListEpics parses every epics/*.md file. Unreadable epics are skipped and
// reported as FileProblems (resilient, like ListTasks).
func (s *FS) ListEpics() ([]domain.Epic, []domain.FileProblem, error) {
	entries, err := os.ReadDir(s.epicsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("read epics dir: %w", err)
	}
	var epics []domain.Epic
	var problems []domain.FileProblem
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(s.epicsDir, e.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("read epic %s: %w", path, err)
		}
		ep, err := parseEpic(content, path)
		if err != nil {
			problems = append(problems, domain.FileProblem{Path: path, Message: err.Error()})
			continue
		}
		epics = append(epics, ep)
	}
	return epics, problems, nil
}

// GetEpic returns one epic plus its markdown body.
func (s *FS) GetEpic(id string) (domain.Epic, string, error) {
	path := filepath.Join(s.epicsDir, id+".md")
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.Epic{}, "", fmt.Errorf("epic %q: %w", id, domain.ErrNotFound)
		}
		return domain.Epic{}, "", fmt.Errorf("read epic %s: %w", path, err)
	}
	ep, err := parseEpic(content, path)
	if err != nil {
		return domain.Epic{}, "", fmt.Errorf("%s: %w", path, err)
	}
	_, body := splitFrontmatter(content)
	return ep, string(body), nil
}

func parseEpic(content []byte, path string) (domain.Epic, error) {
	fm, _ := splitFrontmatter(content)
	var ep domain.Epic
	if len(fm) > 0 {
		if err := yaml.Unmarshal(fm, &ep); err != nil {
			return domain.Epic{}, fmt.Errorf("%w: %s", errBadFrontmatter, frontmatterError(fm, err))
		}
	}
	ep.ID = strings.TrimSuffix(filepath.Base(path), ".md")
	ep.Path = path
	return ep, nil
}
