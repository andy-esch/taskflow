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
var _ core.ThreadPathSource = (*FS)(nil)

// ReadThreads adapts the filesystem's resilient entity scan to the portable
// Thread read contract in one pass. Filename identity recovery belongs here at
// the concrete adapter boundary; core never parses Location for meaning. An
// opaque revision of the exact bytes accompanies unreadable records so guarded
// snapshot comparison cannot mistake an unchanged diagnostic for unchanged data.
func (s *FS) ReadThreads() (core.ThreadRead, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return core.ThreadRead{}, err
	}
	threads, problems, err := scanDirWithSourceVersions(s.threadsDir, func(path string, content []byte) (domain.Thread, error) {
		return parseThread(content, path)
	})
	if err != nil {
		return core.ThreadRead{}, err
	}
	return threadReadFromSourceFiles(threads, problems), nil
}

func threadReadFromSourceFiles(threads []domain.Thread, problems []sourceFileProblem) core.ThreadRead {
	read := core.ThreadRead{
		Threads:  append([]domain.Thread(nil), threads...),
		Problems: make([]core.ThreadReadProblem, 0, len(problems)),
	}
	for _, problem := range problems {
		read.Problems = append(read.Problems, threadReadProblemFromFile(problem.problem, problem.sourceVersion))
	}
	return read
}

func threadReadProblemFromFile(problem domain.FileProblem, sourceVersion string) core.ThreadReadProblem {
	threadID, threadSlug := "", ""
	base := filepath.Base(problem.Path)
	if id, slug, ok := splitFlatName(strings.TrimSuffix(base, ".md")); ok {
		threadID, threadSlug = id, slug
	}
	return core.ThreadReadProblem{
		ThreadID: threadID, ThreadSlug: threadSlug, Location: problem.Path,
		Message: problem.Message, SourceVersion: sourceVersion,
	}
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
