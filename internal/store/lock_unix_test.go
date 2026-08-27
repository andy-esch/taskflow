//go:build unix

package store

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

const (
	lockHelperMode              = "TSKFLWCTL_TEST_LOCK_HELPER"
	lockHelperRoot              = "TSKFLWCTL_TEST_LOCK_ROOT"
	graphMutationHelperMode     = "TSKFLWCTL_TEST_GRAPH_MUTATION_HELPER"
	graphMutationHelperTask     = "TSKFLWCTL_TEST_GRAPH_MUTATION_TASK"
	graphMutationHelperRequires = "TSKFLWCTL_TEST_GRAPH_MUTATION_REQUIRES"
	graphMutationHelperReady    = "TSKFLWCTL_TEST_GRAPH_MUTATION_READY"
	graphMutationHelperStart    = "TSKFLWCTL_TEST_GRAPH_MUTATION_START"
)

// TestRepositoryLockHelperProcess is re-executed by
// TestRepositoryLockReleasesWhenProcessTerminates. It deliberately holds the
// repository lock until the parent kills it; ordinary test runs return at once.
func TestRepositoryLockHelperProcess(t *testing.T) {
	if os.Getenv(lockHelperMode) != "1" {
		return
	}
	unlock, err := NewFS(os.Getenv(lockHelperRoot)).writeLock()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer unlock()
	fmt.Println("locked")
	<-time.After(time.Hour)
}

func TestRepositoryLockReleasesWhenProcessTerminates(t *testing.T) {
	root := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestRepositoryLockHelperProcess$")
	cmd.Env = append(os.Environ(), lockHelperMode+"=1", lockHelperRoot+"="+root)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	stopped := false
	defer func() {
		if !stopped {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	if scanner := bufio.NewScanner(stdout); !scanner.Scan() || scanner.Text() != "locked" {
		t.Fatalf("lock helper did not acquire the repository lock")
	}

	probe, err := os.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(probe.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err == nil {
		_ = syscall.Flock(int(probe.Fd()), syscall.LOCK_UN)
		_ = probe.Close()
		t.Fatal("product lock path did not hold the cross-process repository flock")
	} else if err != syscall.EWOULDBLOCK && err != syscall.EAGAIN {
		_ = probe.Close()
		t.Fatalf("nonblocking lock probe = %v", err)
	}
	_ = probe.Close()

	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err == nil {
		t.Fatal("killed lock helper exited successfully")
	}
	stopped = true
	unlocked, err := NewFS(root).writeLock()
	if err != nil {
		t.Fatalf("repository lock failed after holder exit: %v", err)
	}
	unlocked()
}

func TestRepositoryGraphMutationHelperProcess(t *testing.T) {
	if os.Getenv(graphMutationHelperMode) != "1" {
		return
	}
	if err := os.WriteFile(os.Getenv(graphMutationHelperReady), []byte("ready"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(os.Getenv(graphMutationHelperStart)); err == nil {
			break
		}
		if time.Now().After(deadline) {
			fmt.Fprintln(os.Stderr, "timed out waiting for graph mutation start")
			os.Exit(2)
		}
		time.Sleep(time.Millisecond)
	}
	dependent := os.Getenv(graphMutationHelperTask)
	prerequisite := os.Getenv(graphMutationHelperRequires)
	_, err := NewFS(os.Getenv(lockHelperRoot)).MutateTaskGraph(graphMutationNow, false, func(graph *core.TaskGraph) (core.TaskGraphMutationPlan, error) {
		task, ok := graph.Task(dependent)
		if !ok {
			return core.TaskGraphMutationPlan{}, fmt.Errorf("missing task %s", dependent)
		}
		time.Sleep(100 * time.Millisecond) // widens the cross-process contention window
		return core.TaskGraphMutationPlan{TaskWrites: []core.TaskDependencyWrite{{
			TaskID: dependent, DependsOn: append(task.DependsOn, prerequisite),
		}}}, nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	fmt.Println("applied")
	os.Exit(0)
}

func TestMutateTaskGraphSerializesOppositeEdgesAcrossProcesses(t *testing.T) {
	root := t.TempDir()
	aID, bID := testutil.TaskID("alpha"), testutil.TaskID("beta")
	writeGraphMutationTask(t, root, "alpha", domain.StatusReadyToStart, nil, "")
	writeGraphMutationTask(t, root, "beta", domain.StatusReadyToStart, nil, "")
	coordination := t.TempDir()
	start := filepath.Join(coordination, "start")

	type child struct {
		cmd    *exec.Cmd
		stdout bytes.Buffer
		stderr bytes.Buffer
		ready  string
	}
	newChild := func(dependent, prerequisite, readyName string) *child {
		c := &child{ready: filepath.Join(coordination, readyName)}
		c.cmd = exec.Command(os.Args[0], "-test.run=^TestRepositoryGraphMutationHelperProcess$")
		c.cmd.Env = append(os.Environ(),
			graphMutationHelperMode+"=1",
			lockHelperRoot+"="+root,
			graphMutationHelperTask+"="+dependent,
			graphMutationHelperRequires+"="+prerequisite,
			graphMutationHelperReady+"="+c.ready,
			graphMutationHelperStart+"="+start,
		)
		c.cmd.Stdout, c.cmd.Stderr = &c.stdout, &c.stderr
		return c
	}
	children := []*child{newChild(aID, bID, "first-ready"), newChild(bID, aID, "second-ready")}
	for _, c := range children {
		if err := c.cmd.Start(); err != nil {
			t.Fatal(err)
		}
	}
	deadline := time.Now().Add(5 * time.Second)
	for _, c := range children {
		for {
			if _, err := os.Stat(c.ready); err == nil {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("child did not become ready")
			}
			time.Sleep(time.Millisecond)
		}
	}
	if err := os.WriteFile(start, []byte("start"), 0o644); err != nil {
		t.Fatal(err)
	}
	succeeded, rejected := 0, 0
	for _, c := range children {
		err := c.cmd.Wait()
		switch {
		case err == nil && strings.Contains(c.stdout.String(), "applied"):
			succeeded++
		case err != nil && strings.Contains(c.stderr.String(), "dependency cycle:"):
			rejected++
		default:
			t.Fatalf("unexpected child result err=%v stdout=%q stderr=%q", err, c.stdout.String(), c.stderr.String())
		}
	}
	if succeeded != 1 || rejected != 1 {
		t.Fatalf("cross-process outcomes: succeeded=%d rejected=%d", succeeded, rejected)
	}
	graph, err := core.LoadTaskGraph(NewFS(root))
	if err != nil || graph.Health() != core.GraphHealthy {
		t.Fatalf("final graph health=%v err=%v", graph.Health(), err)
	}
	edges := 0
	for _, taskID := range []string{aID, bID} {
		task, _ := graph.Task(taskID)
		edges += len(task.DependsOn)
	}
	if edges != 1 {
		t.Fatalf("final edge count = %d", edges)
	}
}

func TestGraphMutationLockAcquisitionErrorIsAttributable(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-planning-root")
	_, err := NewFS(missing).MutateTaskGraph(graphMutationNow, false, func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
		return core.TaskGraphMutationPlan{}, nil
	})
	if err == nil || !strings.Contains(err.Error(), "open repo root for write lock") {
		t.Fatalf("lock acquisition error = %v", err)
	}
}

func TestRepositoryLockSerializesIndependentStoresInOneProcess(t *testing.T) {
	root := t.TempDir()
	unlocked := processRepositoryLock(root)
	guard := repositoryGuardFor(root)
	if guard.write.TryLock() {
		guard.write.Unlock()
		unlocked()
		t.Fatal("same-process repository guard did not hold its keyed mutex")
	}
	unlocked()
	if !guard.write.TryLock() {
		t.Fatal("same-process repository guard did not release its keyed mutex")
	}
	guard.write.Unlock()
}

func TestRepositoryLockKeyCanonicalizesEquivalentRoots(t *testing.T) {
	realRoot := t.TempDir()
	parent := t.TempDir()
	alias := filepath.Join(parent, "planning-alias")
	if err := os.Symlink(realRoot, alias); err != nil {
		t.Fatal(err)
	}
	want := repositoryLockKey(realRoot)
	for _, candidate := range []string{
		realRoot + string(filepath.Separator),
		filepath.Join(realRoot, "."),
		filepath.Join(realRoot, "..", filepath.Base(realRoot)),
		alias,
	} {
		if got := repositoryLockKey(candidate); got != want {
			t.Errorf("repositoryLockKey(%q) = %q, want %q", candidate, got, want)
		}
	}
	if got := normalizeRepositoryLockKey(realRoot, true); got != strings.ToLower(want) {
		t.Errorf("case-insensitive lock key = %q, want %q", got, strings.ToLower(want))
	}
}
