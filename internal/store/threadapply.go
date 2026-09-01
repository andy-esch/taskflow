package store

import (
	"errors"
	"fmt"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

var _ core.ThreadApplyMutationStore = (*FS)(nil)

// MutateThreadApply owns the full bulk-link critical section. The operation is
// intentionally resumable rather than transactionally atomic: each task file is
// replaced atomically, every durable graph prefix is valid, and the new Thread
// is created last with no-clobber semantics.
func (s *FS) MutateThreadApply(now time.Time, dryRun bool, planner core.ThreadApplyPlanner) (result core.ThreadApplyMutationResult, err error) {
	result.DryRun = dryRun
	if planner == nil {
		return result, fmt.Errorf("%w: Thread apply planner is required", domain.ErrValidation)
	}
	if now.IsZero() {
		return result, fmt.Errorf("%w: Thread apply time is required", domain.ErrValidation)
	}
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return result, err
	}

	unlock, err := s.checkedWriteLock()
	if err != nil {
		return result, err
	}
	defer func() {
		if releaseErr := unlock(); releaseErr != nil {
			wrapped := fmt.Errorf("release repository Thread apply guard: %w", releaseErr)
			if err == nil {
				err = wrapped
			} else {
				err = errors.Join(err, wrapped)
			}
		}
	}()

	repoID, err := s.currentPlanningIdentity()
	if err != nil {
		return result, err
	}
	graph, err := core.LoadTaskGraph(s)
	if err != nil {
		return result, fmt.Errorf("load authoritative task graph: %w", err)
	}
	threads, bodies, problems, err := s.listThreadApplyThreads()
	if err != nil {
		return result, fmt.Errorf("load authoritative Threads: %w", err)
	}
	if err := core.ValidateThreadCreationSource(graph, threads, threadReadProblemsFromFiles(problems)); err != nil {
		return result, err
	}
	snapshot := core.ThreadApplySnapshot{
		PlanningRepoID: repoID,
		Graph:          graph,
		Threads:        clonePlannerThreads(threads),
		ThreadBodies:   cloneStringMap(bodies),
	}
	plan, err := callThreadApplyPlanner(s, planner, snapshot)
	if err != nil {
		return result, err
	}
	decision, err := core.PrepareThreadApply(snapshot, plan)
	if err != nil {
		return result, err
	}
	result.Plan = decision.Plan
	result.Operations = append([]core.ThreadApplyOperation(nil), decision.Operations...)
	result.Changed = hasPendingThreadApplyOperation(result.Operations)

	graphWrites, err := s.materializeTaskGraphPlan(graph, decision.GraphPlan, now)
	if err != nil {
		return result, err
	}
	if len(graphWrites) != len(decision.GraphPlan.TaskWrites) {
		return result, fmt.Errorf("%w: Thread apply dependency plan did not materialize its complete write set", domain.ErrConflict)
	}
	var threadWrite *materializedThreadCreation
	if decision.ThreadPlan != nil {
		materialized, materializeErr := s.materializeThreadCreation(*decision.ThreadPlan)
		if materializeErr != nil {
			return result, materializeErr
		}
		threadWrite = &materialized
	}
	if dryRun {
		result.Complete = !result.Changed
		return result, nil
	}
	if !result.Changed {
		result.Complete = true
		return result, nil
	}

	// All-or-nothing preflight before the first durable file. The advisory guard
	// excludes cooperating writers; these source hashes and the repeated identity
	// discovery catch raw edits and config/pointer changes.
	if testHookBeforeThreadApplyVerify != nil {
		testHookBeforeThreadApplyVerify()
	}
	currentID, err := s.currentPlanningIdentity()
	if err != nil {
		return result, err
	}
	if currentID != repoID {
		return result, fmt.Errorf("planning repository identity changed while preparing Thread apply: %w", domain.ErrConflict)
	}
	currentGraph, err := core.LoadTaskGraph(s)
	if err != nil {
		return result, fmt.Errorf("re-read authoritative task graph before Thread apply: %w", err)
	}
	currentThreads, _, currentProblems, err := s.listThreadApplyThreads()
	if err != nil {
		return result, fmt.Errorf("re-read authoritative Threads before Thread apply: %w", err)
	}
	if !graph.SameSourceSnapshot(currentGraph) || !sameThreadSourceSnapshot(threads, problems, currentThreads, currentProblems) {
		return result, fmt.Errorf("repository tasks or Threads changed while preparing Thread apply; retry: %w", domain.ErrConflict)
	}

	for _, write := range graphWrites {
		if testHookBeforeThreadApplyWrite != nil {
			testHookBeforeThreadApplyWrite("dependency", write.taskID)
		}
		if err := verifyUnchanged(s.resolvePath, write.taskID, write.path, write.ifVersion, "task", "Thread apply dependency update"); err != nil {
			return result, err
		}
		if err := writeFileAtomic(write.path, write.content, 0o644); err != nil {
			return result, fmt.Errorf("write Thread apply dependencies for task %s: %w", write.taskID, err)
		}
		result.Committed = true
		markThreadApplyDependenciesApplied(result.Operations, write.taskID)
		result.Complete = !hasPendingThreadApplyOperation(result.Operations)
		if testHookAfterThreadApplyWrite != nil {
			if hookErr := testHookAfterThreadApplyWrite("dependency", write.taskID); hookErr != nil {
				return result, fmt.Errorf("after Thread apply dependency update for task %s: %w", write.taskID, hookErr)
			}
		}
	}

	// A raw editor can act despite the advisory guard. Re-prepare after every
	// changed dependency prefix, including recovery runs where the Thread already
	// exists, so an edge undone after its target CAS cannot produce a false
	// complete receipt. A concurrently-created exact Thread becomes a skip.
	if len(graphWrites) > 0 || threadWrite != nil {
		if testHookBeforeThreadApplyWrite != nil {
			testHookBeforeThreadApplyWrite("thread", decision.Plan.Thread.ID)
		}
		result.Complete = false
		finalDecision, prepareErr := s.reprepareThreadApply(decision.Plan, repoID)
		if prepareErr != nil {
			return result, prepareErr
		}
		if pendingDependencyOperation(finalDecision.Operations) {
			markThreadApplyDependenciesPending(result.Operations, finalDecision.Operations)
			return result, fmt.Errorf("planned dependencies changed before final Thread convergence; retry the same plan: %w", domain.ErrConflict)
		}
		finalThreadOperation := finalDecision.Operations[len(finalDecision.Operations)-1]
		if finalThreadOperation.State == core.ThreadApplySkipped {
			markThreadApplyThreadState(result.Operations, core.ThreadApplySkipped)
			result.Changed = hasAppliedThreadApplyOperation(result.Operations)
			result.Complete = true
			return result, nil
		}
		if finalDecision.ThreadPlan == nil {
			return result, fmt.Errorf("%w: Thread apply lost its final creation plan", domain.ErrConflict)
		}
		materialized, materializeErr := s.materializeThreadCreation(*finalDecision.ThreadPlan)
		if materializeErr != nil {
			return result, materializeErr
		}
		threadWrite = &materialized
		if err := s.writeNewFileUnlocked(s.threadsDir, threadWrite.path, threadWrite.content, "Thread", threadWrite.thread.ID+"-"+threadWrite.thread.Slug); err != nil {
			return result, err
		}
		result.Committed = true
		markThreadApplyThreadState(result.Operations, core.ThreadApplyApplied)
		result.Complete = true
		if testHookAfterThreadApplyWrite != nil {
			if hookErr := testHookAfterThreadApplyWrite("thread", threadWrite.thread.ID); hookErr != nil {
				return result, fmt.Errorf("after Thread apply create for %s: %w", threadWrite.thread.ID, hookErr)
			}
		}
	}
	result.Complete = !hasPendingThreadApplyOperation(result.Operations)
	return result, nil
}

func callThreadApplyPlanner(store *FS, planner core.ThreadApplyPlanner, snapshot core.ThreadApplySnapshot) (core.ThreadApplyPlan, error) {
	leave, err := store.enterRepositoryPlanner()
	if err != nil {
		return core.ThreadApplyPlan{}, err
	}
	defer leave()
	return planner(snapshot)
}

func (s *FS) currentPlanningIdentity() (string, error) {
	if s.planningIdentityReader == nil {
		return "", fmt.Errorf("%w: planning repository identity cannot be re-read for Thread apply", domain.ErrValidation)
	}
	root, repoID, err := s.planningIdentityReader()
	if err != nil {
		return "", fmt.Errorf("re-read planning repository identity: %w", err)
	}
	if repositoryLockKey(root) != repositoryLockKey(s.root) {
		return "", fmt.Errorf("planning configuration now resolves to %s instead of guarded root %s: %w", root, s.root, domain.ErrConflict)
	}
	if repoID == "" {
		return "", fmt.Errorf("%w: planning repository has no durable id; run `tskflwctl config migrate` before applying a Thread plan", domain.ErrValidation)
	}
	return repoID, nil
}

type threadApplyDocument struct {
	thread domain.Thread
	body   string
}

func (s *FS) listThreadApplyThreads() ([]domain.Thread, map[string]string, []domain.FileProblem, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return nil, nil, nil, err
	}
	documents, problems, err := scanDir(s.threadsDir, func(path string, content []byte) (threadApplyDocument, error) {
		thread, parseErr := parseThread(content, path)
		if parseErr != nil {
			return threadApplyDocument{}, parseErr
		}
		_, body := splitFrontmatter(content)
		return threadApplyDocument{thread: thread, body: string(body)}, nil
	})
	if err != nil {
		return nil, nil, nil, err
	}
	threads := make([]domain.Thread, 0, len(documents))
	bodies := make(map[string]string, len(documents))
	for _, document := range documents {
		threads = append(threads, document.thread)
		bodies[document.thread.ID] = document.body
	}
	return threads, bodies, problems, nil
}

func (s *FS) reprepareThreadApply(plan core.ThreadApplyPlan, expectedRepoID string) (core.ThreadApplyDecision, error) {
	currentID, err := s.currentPlanningIdentity()
	if err != nil {
		return core.ThreadApplyDecision{}, err
	}
	if currentID != expectedRepoID {
		return core.ThreadApplyDecision{}, fmt.Errorf("planning repository identity changed during Thread apply: %w", domain.ErrConflict)
	}
	graph, err := core.LoadTaskGraph(s)
	if err != nil {
		return core.ThreadApplyDecision{}, fmt.Errorf("re-read task graph before final Thread convergence: %w", err)
	}
	threads, bodies, problems, err := s.listThreadApplyThreads()
	if err != nil {
		return core.ThreadApplyDecision{}, fmt.Errorf("re-read Threads before final Thread convergence: %w", err)
	}
	if err := core.ValidateThreadCreationSource(graph, threads, threadReadProblemsFromFiles(problems)); err != nil {
		return core.ThreadApplyDecision{}, err
	}
	return core.PrepareThreadApply(core.ThreadApplySnapshot{
		PlanningRepoID: currentID, Graph: graph,
		Threads: clonePlannerThreads(threads), ThreadBodies: cloneStringMap(bodies),
	}, plan)
}

func cloneStringMap(values map[string]string) map[string]string {
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

func hasPendingThreadApplyOperation(operations []core.ThreadApplyOperation) bool {
	for _, operation := range operations {
		if operation.State == core.ThreadApplyPending {
			return true
		}
	}
	return false
}

func hasAppliedThreadApplyOperation(operations []core.ThreadApplyOperation) bool {
	for _, operation := range operations {
		if operation.State == core.ThreadApplyApplied {
			return true
		}
	}
	return false
}

func pendingDependencyOperation(operations []core.ThreadApplyOperation) bool {
	for _, operation := range operations {
		if operation.Kind == "dependency" && operation.State == core.ThreadApplyPending {
			return true
		}
	}
	return false
}

func markThreadApplyDependenciesApplied(operations []core.ThreadApplyOperation, taskID string) {
	for index := range operations {
		if operations[index].Kind == "dependency" && operations[index].DependentID == taskID && operations[index].State == core.ThreadApplyPending {
			operations[index].State = core.ThreadApplyApplied
		}
	}
}

func markThreadApplyDependenciesPending(operations, current []core.ThreadApplyOperation) {
	for _, observed := range current {
		if observed.Kind != "dependency" || observed.State != core.ThreadApplyPending {
			continue
		}
		for index := range operations {
			if operations[index].Kind == "dependency" &&
				operations[index].DependentID == observed.DependentID &&
				operations[index].PrerequisiteID == observed.PrerequisiteID {
				operations[index].State = core.ThreadApplyPending
				break
			}
		}
	}
}

func markThreadApplyThreadState(operations []core.ThreadApplyOperation, state core.ThreadApplyOperationState) {
	for index := range operations {
		if operations[index].Kind == "thread" {
			operations[index].State = state
			return
		}
	}
}

// Hooks inject raw-editor races and post-commit interruption in focused tests.
var testHookBeforeThreadApplyVerify func()
var testHookBeforeThreadApplyWrite func(kind, id string)
var testHookAfterThreadApplyWrite func(kind, id string) error
