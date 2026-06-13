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

// GetEpic returns one epic plus its markdown body. The id resolves exact
// first, then fuzzy (unique prefix/substring), like task and audit slugs.
func (s *FS) GetEpic(id string) (domain.Epic, string, error) {
	var cands []candidate
	entries, err := os.ReadDir(s.epicsDir)
	if err != nil && !os.IsNotExist(err) {
		return domain.Epic{}, "", fmt.Errorf("read epics dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		cands = append(cands, candidate{
			id:   strings.TrimSuffix(e.Name(), ".md"),
			path: filepath.Join(s.epicsDir, e.Name()),
		})
	}
	c, err := resolveID("epic", id, cands)
	if err != nil {
		return domain.Epic{}, "", err
	}
	content, err := os.ReadFile(c.path)
	if err != nil {
		return domain.Epic{}, "", fmt.Errorf("read epic %s: %w", c.path, err)
	}
	path := c.path
	ep, err := parseEpic(content, path)
	if err != nil {
		return domain.Epic{}, "", fmt.Errorf("%s: %w", path, err)
	}
	_, body := splitFrontmatter(content)
	return ep, string(body), nil
}

func parseEpic(content []byte, path string) (domain.Epic, error) {
	fm, _, err := splitFrontmatterStrict(content)
	if err != nil {
		return domain.Epic{}, err
	}
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
