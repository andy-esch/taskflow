package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
	"github.com/andy-esch/taskflow/internal/threadspike"
)

// ThreadSpikeRepository is an explicitly experimental adapter used only to
// exercise ADR-0006 against the real Markdown parser, atomic writes, CAS, and
// repository lock. It is not wired into core.Service or a public CLI.
type ThreadSpikeRepository struct {
	fs     *FS
	repoID string
}

func NewThreadSpikeRepository(root, repoID string) *ThreadSpikeRepository {
	return &ThreadSpikeRepository{fs: NewFS(root), repoID: repoID}
}

func (r *ThreadSpikeRepository) Snapshot() (threadspike.Snapshot, error) {
	return r.snapshot()
}

func (r *ThreadSpikeRepository) snapshot() (threadspike.Snapshot, error) {
	snapshot := threadspike.Snapshot{
		RepoID:  r.repoID,
		Tasks:   map[string]threadspike.Task{},
		Threads: map[string]threadspike.Thread{},
		Epics:   map[string]bool{},
	}
	tasks, taskProblems, err := scanDir(r.fs.tasksDir, parseThreadSpikeTask)
	if err != nil {
		return threadspike.Snapshot{}, err
	}
	snapshot.Problems = append(snapshot.Problems, taskProblems...)
	for _, task := range tasks {
		if !id.Valid(task.Record.ID) || task.Record.ID != task.Record.FilenameID {
			snapshot.Problems = append(snapshot.Problems, domain.FileProblem{
				Path: task.Record.Path, Message: fmt.Sprintf("task frontmatter id %q does not match filename id %q", task.Record.ID, task.Record.FilenameID),
			})
			continue
		}
		if !task.Record.Status.Valid() {
			snapshot.Problems = append(snapshot.Problems, domain.FileProblem{
				Path: task.Record.Path, Message: fmt.Sprintf("task has invalid status %q", task.Record.Status),
			})
			continue
		}
		if _, duplicate := snapshot.Tasks[task.Record.ID]; duplicate {
			snapshot.Problems = append(snapshot.Problems, domain.FileProblem{
				Path: task.Record.Path, Message: "duplicate task id " + task.Record.ID,
			})
			continue
		}
		snapshot.Tasks[task.Record.ID] = task
	}
	epics, epicProblems, err := r.fs.ListEpics()
	if err != nil {
		return threadspike.Snapshot{}, err
	}
	snapshot.Problems = append(snapshot.Problems, epicProblems...)
	for _, epic := range epics {
		snapshot.Epics[epic.ID] = true
	}
	threadsDir := filepath.Join(r.fs.root, threadspike.Dir)
	threads, threadProblems, err := scanDir(threadsDir, parseThreadSpikeThread)
	if err != nil {
		return threadspike.Snapshot{}, err
	}
	snapshot.Problems = append(snapshot.Problems, threadProblems...)
	for _, thread := range threads {
		if _, duplicate := snapshot.Threads[thread.ID]; duplicate {
			snapshot.Problems = append(snapshot.Problems, domain.FileProblem{
				Path: thread.Path, Message: "duplicate Thread id " + thread.ID,
			})
			continue
		}
		snapshot.Threads[thread.ID] = thread
	}
	for _, thread := range snapshot.Threads {
		for _, memberID := range thread.Tasks {
			if _, exists := snapshot.Tasks[memberID]; !exists {
				snapshot.Problems = append(snapshot.Problems, domain.FileProblem{
					Path: thread.Path, Message: fmt.Sprintf("Thread member %s does not exist", memberID),
				})
			}
		}
	}
	sort.Slice(snapshot.Problems, func(i, j int) bool {
		if snapshot.Problems[i].Path != snapshot.Problems[j].Path {
			return snapshot.Problems[i].Path < snapshot.Problems[j].Path
		}
		return snapshot.Problems[i].Message < snapshot.Problems[j].Message
	})
	return snapshot, nil
}

func parseThreadSpikeTask(path string, content []byte) (threadspike.Task, error) {
	record, err := parseTask(content, path)
	if err != nil {
		return threadspike.Task{}, err
	}
	fm, body, err := splitFrontmatterStrict(content)
	if err != nil {
		return threadspike.Task{}, err
	}
	var proposed struct {
		DependsOn []string `yaml:"depends_on"`
	}
	if err := yaml.Unmarshal(fm, &proposed); err != nil {
		return threadspike.Task{}, fmt.Errorf("%w: %s", errBadFrontmatter, frontmatterError(fm, err))
	}
	return threadspike.Task{Record: record, DependsOn: proposed.DependsOn, Body: string(body)}, nil
}

func parseThreadSpikeThread(path string, content []byte) (threadspike.Thread, error) {
	base := filepath.Base(path)
	filenameID, slug, ok := splitFlatName(strings.TrimSuffix(base, ".md"))
	if !ok {
		reason, kind := entityNameProblem(base)
		return threadspike.Thread{}, fmt.Errorf("%w: %q %s", kind, base, reason)
	}
	fm, body, err := splitFrontmatterStrict(content)
	if err != nil {
		return threadspike.Thread{}, err
	}
	if fm == nil {
		return threadspike.Thread{}, missingFrontmatterErr("Thread", "id, status, description, goal, created, tasks")
	}
	var thread threadspike.Thread
	if err := yaml.Unmarshal(fm, &thread); err != nil {
		return threadspike.Thread{}, fmt.Errorf("%w: %s", errBadFrontmatter, frontmatterError(fm, err))
	}
	if thread.ID != filenameID {
		return threadspike.Thread{}, fmt.Errorf("%w: Thread id %q does not match filename id %q", domain.ErrValidation, thread.ID, filenameID)
	}
	if !thread.Status.Valid() {
		return threadspike.Thread{}, fmt.Errorf("%w: invalid Thread status %q", domain.ErrValidation, thread.Status)
	}
	if strings.TrimSpace(thread.Description) == "" || strings.TrimSpace(thread.Goal) == "" || strings.TrimSpace(thread.Created) == "" {
		return threadspike.Thread{}, fmt.Errorf("%w: Thread description, goal, and created are required", domain.ErrValidation)
	}
	if len(thread.Tasks) != len(uniqueStrings(thread.Tasks)) {
		return threadspike.Thread{}, fmt.Errorf("%w: Thread tasks contains duplicate ids", domain.ErrValidation)
	}
	thread.Slug = slug
	thread.Path = path
	thread.Body = string(body)
	return thread, nil
}

// Apply revalidates and writes while holding the same repository lock used by
// production mutations. Writes are atomic one file at a time, not transactional
// as a group; the materialized plan and prefix receipt make interruption resumable.
func (r *ThreadSpikeRepository) Apply(plan threadspike.MaterializedPlan, options threadspike.ApplyOptions) (threadspike.Receipt, error) {
	if options.DryRun {
		snapshot, err := r.snapshot()
		if err != nil {
			return threadspike.Receipt{}, err
		}
		decision, err := threadspike.PrepareApply(snapshot, plan)
		if err != nil {
			return threadspike.Receipt{}, err
		}
		return previewReceipt(decision), nil
	}

	unlock, err := r.fs.writeLock()
	if err != nil {
		return threadspike.Receipt{}, err
	}
	defer unlock()
	snapshot, err := r.snapshot()
	if err != nil {
		return threadspike.Receipt{}, err
	}
	decision, err := threadspike.PrepareApply(snapshot, plan)
	if err != nil {
		return threadspike.Receipt{}, err
	}
	receipt := threadspike.Receipt{Entries: append([]threadspike.Operation(nil), decision.Skipped...)}
	record := func(operation threadspike.Operation) error {
		receipt.Entries = append(receipt.Entries, operation)
		if options.AfterWrite != nil {
			if err := options.AfterWrite(operation); err != nil {
				return fmt.Errorf("injected interruption after %s %s: %w", operation.Kind, operation.ID, err)
			}
		}
		return nil
	}
	for _, create := range decision.CreateTasks {
		if err := r.createTask(create); err != nil {
			return receipt, err
		}
		if err := record(threadspike.Operation{Kind: "task", ID: create.Task.ID, Action: "created"}); err != nil {
			return receipt, err
		}
	}
	for _, update := range decision.UpdateDependencies {
		if err := r.updateDependencies(update); err != nil {
			return receipt, err
		}
		if err := record(threadspike.Operation{Kind: "dependency-set", ID: update.TaskID, Action: "updated"}); err != nil {
			return receipt, err
		}
	}
	if decision.CreateThread != nil {
		if err := r.createThread(*decision.CreateThread); err != nil {
			return receipt, err
		}
		if err := record(threadspike.Operation{Kind: "thread", ID: decision.CreateThread.ID, Action: "created"}); err != nil {
			return receipt, err
		}
	}
	receipt.Complete = true
	return receipt, nil
}

func previewReceipt(decision threadspike.ApplyDecision) threadspike.Receipt {
	receipt := threadspike.Receipt{Complete: true, Entries: append([]threadspike.Operation(nil), decision.Skipped...)}
	for _, create := range decision.CreateTasks {
		receipt.Entries = append(receipt.Entries, threadspike.Operation{Kind: "task", ID: create.Task.ID, Action: "would-create"})
	}
	for _, update := range decision.UpdateDependencies {
		receipt.Entries = append(receipt.Entries, threadspike.Operation{Kind: "dependency-set", ID: update.TaskID, Action: "would-update"})
	}
	if decision.CreateThread != nil {
		receipt.Entries = append(receipt.Entries, threadspike.Operation{Kind: "thread", ID: decision.CreateThread.ID, Action: "would-create"})
	}
	return receipt
}

func (r *ThreadSpikeRepository) createTask(create threadspike.TaskCreate) error {
	task := create.Task.Task(create.DependsOn)
	if err := domain.ActiveTaskFieldErr(task.Record); err != nil {
		return err
	}
	fields := taskFields(task.Record)
	if len(task.DependsOn) > 0 {
		fields = append(fields, fmField{"depends_on", task.DependsOn})
	}
	content, err := buildFile(fields, task.Body)
	if err != nil {
		return err
	}
	path := filepath.Join(r.fs.tasksDir, task.Record.ID+"-"+task.Record.Slug+".md")
	return r.fs.writeNewFile(r.fs.tasksDir, path, content, "task", task.Record.ID, false)
}

func (r *ThreadSpikeRepository) updateDependencies(update threadspike.DependencyUpdate) error {
	path, err := r.fs.resolvePath(update.TaskID)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read task %s: %w", path, err)
	}
	newContent, err := updateFrontmatter(content, map[string]any{"depends_on": update.DependsOn})
	if err != nil {
		return err
	}
	if _, err := parseThreadSpikeTask(path, newContent); err != nil {
		return fmt.Errorf("%w: dependency update would not reload: %v", domain.ErrValidation, err)
	}
	if err := verifyUnchanged(r.fs.resolvePath, update.TaskID, path, hashContent(content), "task", "dependency update"); err != nil {
		return err
	}
	return writeFileAtomic(path, newContent, 0o644)
}

func (r *ThreadSpikeRepository) createThread(planned threadspike.PlannedThread) error {
	thread := planned.Thread()
	fields := []fmField{
		{"schema", domain.FileSchemaVersion},
		{"id", thread.ID},
		{"status", string(thread.Status)},
		{"description", thread.Description},
		{"goal", thread.Goal},
		{"created", thread.Created},
		{"tags", thread.Tags},
		{"tasks", thread.Tasks},
	}
	content, err := buildFile(fields, thread.Body)
	if err != nil {
		return err
	}
	dir := filepath.Join(r.fs.root, threadspike.Dir)
	path := filepath.Join(dir, thread.ID+"-"+thread.Slug+".md")
	return r.fs.writeNewFile(dir, path, content, "Thread", thread.ID, false)
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	out := make([]string, 0, len(seen))
	for value := range seen {
		out = append(out, value)
	}
	return out
}
