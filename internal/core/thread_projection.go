package core

import (
	"fmt"
	"sort"

	"github.com/andy-esch/taskflow/internal/domain"
)

// ThreadTaskRole distinguishes persisted membership from a boundary task found
// through the repository-global dependency graph.
type ThreadTaskRole string

const (
	ThreadTaskMember       ThreadTaskRole = "member"
	ThreadTaskExternalGate ThreadTaskRole = "external-gate"
)

// ThreadProblemCode is stable Thread-local diagnostic vocabulary: document and
// membership defects plus explanatory completed-Thread consistency reasons.
type ThreadProblemCode string

const (
	ThreadProblemInvalidDocument            ThreadProblemCode = "invalid-thread-document"
	ThreadProblemIDDrift                    ThreadProblemCode = "thread-id-drift"
	ThreadProblemDuplicateID                ThreadProblemCode = "duplicate-thread-id"
	ThreadProblemTaskIDCollision            ThreadProblemCode = "thread-task-id-collision"
	ThreadProblemMissingMember              ThreadProblemCode = "missing-thread-member"
	ThreadProblemCompletedEmpty             ThreadProblemCode = "completed-thread-empty"
	ThreadProblemCompletedUndrained         ThreadProblemCode = "completed-thread-undrained"
	ThreadProblemCompletedExternalGate      ThreadProblemCode = "completed-thread-external-gate"
	ThreadProblemCompletedUnhealthyEvidence ThreadProblemCode = "completed-thread-unhealthy-evidence"
)

type ThreadProblem struct {
	Code     ThreadProblemCode
	ThreadID string
	TaskID   string
	Path     string
	Message  string
}

type ThreadTaskView struct {
	Role  ThreadTaskRole
	Task  domain.Task
	State TaskGraphState
}

type ThreadExternalGate struct {
	ThreadTaskView
	Outstanding bool
}

type ThreadRollup struct {
	Done       int
	Total      int
	Drained    int
	Deprecated int
}

// ThreadView is the one adapter-neutral runtime projection shared by Thread
// list/show/frontier and their human and machine renderers.
type ThreadView struct {
	Thread           domain.Thread
	Source           RecordSource
	Rollup           ThreadRollup
	Members          []ThreadTaskView
	ExternalGates    []ThreadExternalGate
	Frontier         []ThreadTaskView
	GraphHealth      GraphHealth
	ProjectionHealth GraphHealth
	GraphProblems    []GraphProblem
	Problems         []ThreadProblem
	Inconsistent     bool
}

// ThreadListView hoists repository-global graph diagnostics out of individual
// Thread rows. A list may therefore report a broken empty repository without
// repeating the same graph-problem set once per Thread.
type ThreadListView struct {
	Threads       []ThreadView
	GraphHealth   GraphHealth
	GraphProblems []GraphProblem
}

// ProjectThread joins one persisted membership set to the immutable repository
// task graph. It never reads storage and never derives dependency rules itself.
func ProjectThread(thread domain.Thread, graph *TaskGraph) ThreadView {
	return projectThread(thread, RecordSource{ID: thread.CanonicalID(), Location: thread.Path, LocationIsPath: thread.Path != ""}, graph)
}

// ProjectLoadedThread uses the identity of the selected source, including when
// its frontmatter ID has drifted. A readable location is diagnostic context,
// not a local path unless the adapter supplied that presentation evidence.
func ProjectLoadedThread(record LoadedRecord[domain.Thread], graph *TaskGraph) ThreadView {
	return projectThread(record.Value, record.Source, graph)
}

func projectThread(thread domain.Thread, source RecordSource, graph *TaskGraph) ThreadView {
	path := ""
	if source.LocationIsPath {
		path = source.Location
	}
	view := ThreadView{Thread: cloneThread(thread), Source: source, GraphHealth: GraphBroken, ProjectionHealth: GraphBroken}
	if graph == nil {
		view.Problems = append(view.Problems, ThreadProblem{
			Code: ThreadProblemInvalidDocument, ThreadID: thread.ID, Path: path,
			Message: "repository task graph is unavailable",
		})
		if thread.Status == domain.ThreadStatusCompleted {
			view.Inconsistent = true
			view.Problems = append(view.Problems, ThreadProblem{
				Code: ThreadProblemCompletedUnhealthyEvidence, ThreadID: thread.ID, Path: path,
				Message: "completed Thread has broken projection evidence",
			})
		}
		return view
	}
	view.GraphHealth = graph.Health()
	view.ProjectionHealth = view.GraphHealth
	view.GraphProblems = graph.Problems()
	if err := domain.ValidateThreadDocument(thread); err != nil {
		view.ProjectionHealth = GraphBroken
		view.Problems = append(view.Problems, ThreadProblem{
			Code: ThreadProblemInvalidDocument, ThreadID: thread.ID, Path: path,
			Message: err.Error(),
		})
	}
	if source.ID != "" && source.ID != thread.ID {
		view.ProjectionHealth = GraphBroken
		view.Problems = append(view.Problems, ThreadProblem{
			Code: ThreadProblemIDDrift, ThreadID: thread.ID, Path: path,
			Message: fmt.Sprintf("Thread id drift: frontmatter=%q source=%q", thread.ID, source.ID),
		})
	}
	collisionID := thread.ID
	if _, collision := graph.Task(collisionID); !collision {
		collisionID = source.ID
	}
	if collisionID != "" {
		if _, collision := graph.Task(collisionID); collision {
			view.ProjectionHealth = GraphBroken
			view.Problems = append(view.Problems, ThreadProblem{
				Code: ThreadProblemTaskIDCollision, ThreadID: thread.ID, TaskID: collisionID, Path: path,
				Message: fmt.Sprintf("stable id %s is used by both a task and a Thread", collisionID),
			})
		}
	}

	memberIDs := append([]string(nil), thread.Tasks...)
	sort.Strings(memberIDs)
	members := make(map[string]bool, len(memberIDs))
	for _, taskID := range memberIDs {
		if members[taskID] {
			continue
		}
		members[taskID] = true
		task, exists := graph.Task(taskID)
		state := graph.State(taskID)
		member := ThreadTaskView{Role: ThreadTaskMember, Task: task, State: state}
		if !exists {
			member.Task.ID = taskID
			view.ProjectionHealth = GraphBroken
			view.Problems = append(view.Problems, ThreadProblem{
				Code: ThreadProblemMissingMember, ThreadID: thread.ID, TaskID: taskID, Path: path,
				Message: fmt.Sprintf("thread %s references missing task %s", thread.ID, taskID),
			})
			// A declared but unreadable/missing member remains real unfinished work;
			// excluding it would manufacture a misleading 0/0 or 100% rollup.
			view.Rollup.Total++
			view.Members = append(view.Members, member)
			continue
		}
		view.Members = append(view.Members, member)
		if task.Status == domain.StatusDeprecated {
			view.Rollup.Deprecated++
			continue
		}
		view.Rollup.Total++
		if task.Status == domain.StatusCompleted {
			view.Rollup.Done++
		}
		if state.Drained {
			view.Rollup.Drained++
		}
	}

	external := make(map[string]bool)
	for _, taskID := range memberIDs {
		// Withdrawn work is excluded from the Thread completion denominator and
		// cannot be dispatched. Its own prerequisites therefore do not gate the
		// initiative or appear as actionable external work.
		if task, exists := graph.Task(taskID); exists && task.Status == domain.StatusDeprecated {
			continue
		}
		for _, prerequisite := range graph.Prerequisites(taskID) {
			if !members[prerequisite] {
				external[prerequisite] = true
			}
		}
	}
	externalIDs := make([]string, 0, len(external))
	for taskID := range external {
		externalIDs = append(externalIDs, taskID)
	}
	sort.Strings(externalIDs)
	for _, taskID := range externalIDs {
		task, _ := graph.Task(taskID)
		state := graph.State(taskID)
		view.ExternalGates = append(view.ExternalGates, ThreadExternalGate{
			ThreadTaskView: ThreadTaskView{Role: ThreadTaskExternalGate, Task: task, State: state},
			Outstanding:    !state.SoundlyCompleted,
		})
	}

	if view.ProjectionHealth == GraphHealthy {
		for _, member := range view.Members {
			if member.State.Eligible {
				view.Frontier = append(view.Frontier, member)
			}
		}
	}
	if thread.Status == domain.ThreadStatusCompleted {
		if view.ProjectionHealth != GraphHealthy {
			view.Inconsistent = true
			view.Problems = append(view.Problems, ThreadProblem{
				Code: ThreadProblemCompletedUnhealthyEvidence, ThreadID: thread.ID, Path: path,
				Message: fmt.Sprintf("completed Thread has %s projection evidence", view.ProjectionHealth),
			})
		}
		if view.Rollup.Total == 0 {
			view.Inconsistent = true
			view.Problems = append(view.Problems, ThreadProblem{
				Code: ThreadProblemCompletedEmpty, ThreadID: thread.ID, Path: path,
				Message: "completed Thread has no non-deprecated members",
			})
		}
		if view.Rollup.Drained != view.Rollup.Total {
			view.Inconsistent = true
			view.Problems = append(view.Problems, ThreadProblem{
				Code: ThreadProblemCompletedUndrained, ThreadID: thread.ID, Path: path,
				Message: fmt.Sprintf("completed Thread has %d of %d members soundly completed", view.Rollup.Drained, view.Rollup.Total),
			})
		}
		for _, gate := range view.ExternalGates {
			if !gate.Outstanding {
				continue
			}
			view.Inconsistent = true
			view.Problems = append(view.Problems, ThreadProblem{
				Code: ThreadProblemCompletedExternalGate, ThreadID: thread.ID, TaskID: gate.State.TaskID, Path: path,
				Message: fmt.Sprintf("completed Thread has outstanding external gate %s", gate.State.TaskID),
			})
		}
	}
	return view
}

func cloneThread(thread domain.Thread) domain.Thread {
	thread.Tags = append([]string(nil), thread.Tags...)
	thread.Tasks = append([]string(nil), thread.Tasks...)
	return thread
}
