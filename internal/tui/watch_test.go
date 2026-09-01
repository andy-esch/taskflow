package tui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func nextWatcherEvent(t *testing.T, w *watcher, action func() error) fsEventMsg {
	t.Helper()
	result := make(chan any, 1)
	go func() { result <- waitForFS(w)() }()
	if action != nil {
		if err := action(); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case msg := <-result:
		event, ok := msg.(fsEventMsg)
		if !ok {
			t.Fatalf("watch result = %T, want fsEventMsg", msg)
		}
		return event
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for filesystem event")
		return fsEventMsg{}
	}
}

func queuedWatcherEvent(t *testing.T, w *watcher, action func() error) fsEventMsg {
	t.Helper()
	if err := action(); err != nil {
		t.Fatal(err)
	}
	return nextWatcherEvent(t, w, nil)
}

func watcherHealthAfter(t *testing.T, w *watcher, action func() error, want watchHealth) {
	t.Helper()
	event := nextWatcherEvent(t, w, action)
	if event.health != want {
		event.health = w.reconcile()
	}
	if event.health != want {
		t.Fatalf("watch health = %v, want %v", event.health, want)
	}
}

func TestNewWatcherRejectsEmptyAndFilesystemRootPaths(t *testing.T) {
	for name, paths := range map[string][]string{
		"empty": nil,
		"root":  {string(filepath.Separator)},
	} {
		t.Run(name, func(t *testing.T) {
			if w, err := newWatcher(paths); err == nil {
				_ = w.close()
				t.Fatal("unbounded watcher must be unavailable")
			}
		})
	}
}

func TestWatcherRecoversMissingDesiredLeafAndObservesItsFiles(t *testing.T) {
	root := t.TempDir()
	tasks := filepath.Join(root, "tasks")
	threads := filepath.Join(root, "threads")
	if err := os.Mkdir(tasks, 0o755); err != nil {
		t.Fatal(err)
	}
	w, err := newWatcher([]string{tasks, threads})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = w.close() })
	if got := w.watchHealth(); got != watchDegraded {
		t.Fatalf("initial health = %v, want degraded", got)
	}

	watcherHealthAfter(t, w, func() error { return os.Mkdir(threads, 0o755) }, watchHealthy)
	threadPath := filepath.Join(threads, "6g5w000000q1-new-thread.md")
	watcherHealthAfter(t, w, func() error {
		return os.WriteFile(threadPath, []byte("thread\n"), 0o644)
	}, watchHealthy)
}

func TestWatcherRecoversNestedDesiredLeafWhenParentTreeAppears(t *testing.T) {
	root := t.TempDir()
	tasks := filepath.Join(root, "tasks")
	threads := filepath.Join(root, "future", "planning", "threads")
	if err := os.Mkdir(tasks, 0o755); err != nil {
		t.Fatal(err)
	}
	w, err := newWatcher([]string{tasks, threads})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = w.close() })
	if got := w.watchHealth(); got != watchDegraded {
		t.Fatalf("initial health = %v, want degraded", got)
	}
	watcherHealthAfter(t, w, func() error { return os.MkdirAll(threads, 0o755) }, watchHealthy)
}

func TestDebounceTickPerformsLoadBearingQuietPeriodReconciliation(t *testing.T) {
	root := t.TempDir()
	tasks := filepath.Join(root, "tasks")
	firstParent := filepath.Join(root, "future")
	threads := filepath.Join(firstParent, "planning", "threads")
	if err := os.Mkdir(tasks, 0o755); err != nil {
		t.Fatal(err)
	}
	w, err := newWatcher([]string{tasks, threads})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = w.close() })

	event := queuedWatcherEvent(t, w, func() error { return os.Mkdir(firstParent, 0o755) })
	if event.health != watchDegraded {
		t.Fatalf("first-parent event health = %v, want degraded", event.health)
	}
	if err := os.MkdirAll(threads, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(threads, "immediate.md"), []byte("thread\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	msg, ok := debounceTick(w, 7)().(debounceMsg)
	if !ok || msg.gen != 7 || msg.health != watchHealthy {
		t.Fatalf("quiet-period result = %#v, want gen 7 healthy", msg)
	}
	w.mu.Lock()
	_, attached := w.attached[canonicalWatchPath(threads)]
	w.mu.Unlock()
	if !attached {
		t.Fatal("quiet-period reconciliation did not attach the completed leaf")
	}
}

func TestNewWatcherRetainsUnavailableStateForManualRecovery(t *testing.T) {
	tasks := filepath.Join(t.TempDir(), "tasks")
	if err := os.Mkdir(tasks, 0o755); err != nil {
		t.Fatal(err)
	}
	w, err := newWatcherWithPathAdder([]string{tasks}, func(string) error {
		return errors.New("temporary watch exhaustion")
	})
	if err != nil || w == nil {
		t.Fatalf("retryable watcher = %v, %v", w, err)
	}
	t.Cleanup(func() { _ = w.close() })
	if got := w.watchHealth(); got != watchUnavailable {
		t.Fatalf("initial health = %v, want unavailable", got)
	}
	w.mu.Lock()
	w.addPath = w.fsw.Add
	w.mu.Unlock()
	msg, ok := reconcileWatcher(w)().(watcherReconciledMsg)
	if !ok || msg.health != watchHealthy {
		t.Fatalf("manual recovery = %#v, want healthy", msg)
	}
}

func TestWatcherReportsTransientAddFailureAndExplicitReconcileRecovers(t *testing.T) {
	root := t.TempDir()
	current := filepath.Join(root, "threads")
	replacement := filepath.Join(root, "replacement")
	retired := filepath.Join(root, "retired")
	for _, path := range []string{current, replacement} {
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	w, err := newWatcher([]string{current})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = w.close() })

	w.mu.Lock()
	realAdd := w.addPath
	w.addPath = func(path string) error {
		if path == canonicalWatchPath(current) {
			return errors.New("temporary watch limit")
		}
		return realAdd(path)
	}
	w.mu.Unlock()
	if err := os.Rename(current, retired); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, current); err != nil {
		t.Fatal(err)
	}
	if got := w.reconcile(); got != watchDegraded {
		t.Fatalf("health after direct Add failure = %v, want degraded", got)
	}

	w.mu.Lock()
	w.addPath = realAdd
	w.mu.Unlock()
	msg, ok := reconcileWatcher(w)().(watcherReconciledMsg)
	if !ok || msg.health != watchHealthy {
		t.Fatalf("explicit recovery = %#v, want healthy", msg)
	}
}

func TestWatcherRejectsAttachmentReplacedDuringAdd(t *testing.T) {
	root := t.TempDir()
	current := filepath.Join(root, "threads")
	replacement := filepath.Join(root, "replacement")
	retired := filepath.Join(root, "retired")
	if err := os.Mkdir(replacement, 0o755); err != nil {
		t.Fatal(err)
	}
	w, err := newWatcher([]string{current})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = w.close() })

	w.mu.Lock()
	realAdd := w.addPath
	w.addPath = func(path string) error {
		if err := realAdd(path); err != nil {
			return err
		}
		if path != canonicalWatchPath(current) {
			return nil
		}
		if err := os.Rename(current, retired); err != nil {
			return err
		}
		return os.Rename(replacement, current)
	}
	w.mu.Unlock()
	if err := os.Mkdir(current, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := w.reconcile(); got != watchDegraded {
		t.Fatalf("health after replacement during Add = %v, want degraded", got)
	}
	w.mu.Lock()
	_, staleAttached := w.attached[canonicalWatchPath(current)]
	w.addPath = realAdd
	w.mu.Unlock()
	if staleAttached {
		t.Fatal("watcher recorded an attachment whose identity changed during Add")
	}
	if got := w.reconcile(); got != watchHealthy {
		t.Fatalf("health after stable retry = %v, want healthy", got)
	}
}

func TestWatcherReattachesRemovedAndRecreatedDirectory(t *testing.T) {
	root := t.TempDir()
	threads := filepath.Join(root, "threads")
	if err := os.Mkdir(threads, 0o755); err != nil {
		t.Fatal(err)
	}
	w, err := newWatcher([]string{threads})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = w.close() })

	watcherHealthAfter(t, w, func() error { return os.Remove(threads) }, watchDegraded)
	watcherHealthAfter(t, w, func() error { return os.Mkdir(threads, 0o755) }, watchHealthy)
	watcherHealthAfter(t, w, func() error {
		return os.WriteFile(filepath.Join(threads, "after-recreate.md"), []byte("ok\n"), 0o644)
	}, watchHealthy)
}

func TestWatcherReattachesAtomicallyReplacedDirectory(t *testing.T) {
	root := t.TempDir()
	current := filepath.Join(root, "threads")
	replacement := filepath.Join(root, "replacement")
	retired := filepath.Join(root, "retired")
	for _, path := range []string{current, replacement} {
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	w, err := newWatcher([]string{current})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = w.close() })

	event := queuedWatcherEvent(t, w, func() error {
		if err := os.Rename(current, retired); err != nil {
			return err
		}
		return os.Rename(replacement, current)
	})
	if event.health != watchHealthy {
		t.Fatalf("replacement event health = %v, want healthy", event.health)
	}

	w.mu.Lock()
	attached := w.attached[canonicalWatchPath(current)]
	w.mu.Unlock()
	currentInfo, err := os.Stat(current)
	if err != nil {
		t.Fatal(err)
	}
	if attached == nil {
		t.Fatalf("replacement path is not attached; health=%v", w.watchHealth())
	}
	if !os.SameFile(attached, currentInfo) {
		t.Fatalf("watcher retained the replaced directory inode: attached=%#v current=%#v health=%v", attached.Sys(), currentInfo.Sys(), w.watchHealth())
	}
	watcherHealthAfter(t, w, func() error {
		return os.WriteFile(filepath.Join(current, "after-replacement.md"), []byte("ok\n"), 0o644)
	}, watchHealthy)
}

func TestWatcherRecoversRetargetedSymlink(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first")
	second := filepath.Join(root, "second")
	for _, target := range []string{first, second} {
		if err := os.MkdirAll(filepath.Join(target, "tasks"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	current := filepath.Join(root, "current")
	if err := os.Symlink(first, current); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	next := filepath.Join(root, "current-next")
	if err := os.Symlink(second, next); err != nil {
		t.Skipf("second symlink unavailable: %v", err)
	}
	w, err := newWatcher([]string{filepath.Join(current, "tasks")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = w.close() })

	if got := w.watchHealth(); got != watchDegraded {
		t.Fatalf("symlink-backed initial health = %v, want degraded", got)
	}
	if err := os.Rename(next, current); err != nil {
		t.Fatal(err)
	}
	// kqueue does not reliably report replacing a same-name symlink entry. A
	// manual refresh must therefore re-resolve it while health remains honestly
	// degraded instead of promising portable retarget coverage.
	if got := w.reconcile(); got != watchDegraded {
		t.Fatalf("retarget reconciliation health = %v, want degraded", got)
	}
	w.mu.Lock()
	_, oldAttached := w.attached[canonicalWatchPath(filepath.Join(first, "tasks"))]
	_, newAttached := w.attached[canonicalWatchPath(filepath.Join(second, "tasks"))]
	w.mu.Unlock()
	if oldAttached || !newAttached {
		t.Fatalf("retarget attachments: old=%v new=%v desired=%v", oldAttached, newAttached, w.desired)
	}
	watcherHealthAfter(t, w, func() error {
		return os.WriteFile(filepath.Join(current, "tasks", "after-retarget.md"), []byte("ok\n"), 0o644)
	}, watchDegraded)
}

func TestWatcherNormalizesSymlinkAliasesBeforeAttaching(t *testing.T) {
	realRoot := t.TempDir()
	tasks := filepath.Join(realRoot, "tasks")
	if err := os.Mkdir(tasks, 0o755); err != nil {
		t.Fatal(err)
	}
	aliasRoot := filepath.Join(t.TempDir(), "planning-alias")
	if err := os.Symlink(realRoot, aliasRoot); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	w, err := newWatcher([]string{tasks, filepath.Join(aliasRoot, "tasks")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = w.close() })
	if len(w.desired) != 1 || w.desired[0] != canonicalWatchPath(tasks) {
		t.Fatalf("normalized desired paths = %v", w.desired)
	}
	w.mu.Lock()
	attached := len(w.attached)
	w.mu.Unlock()
	if attached != 3 { // canonical leaf + target parent + lexical symlink-parent sentinel
		t.Fatalf("attached watch count = %d, want 3", attached)
	}
	if got := w.watchHealth(); got != watchDegraded {
		t.Fatalf("symlink-backed health = %v, want degraded", got)
	}
}

func TestWatcherCloseReleasesLeafAndSentinelWatches(t *testing.T) {
	tasks := filepath.Join(t.TempDir(), "tasks")
	if err := os.Mkdir(tasks, 0o755); err != nil {
		t.Fatal(err)
	}
	w, err := newWatcher([]string{tasks})
	if err != nil {
		t.Fatal(err)
	}
	if got := len(w.fsw.WatchList()); got != 2 {
		t.Fatalf("watch count before close = %d, want leaf plus sentinel", got)
	}
	if err := w.close(); err != nil {
		t.Fatal(err)
	}
	if got := len(w.fsw.WatchList()); got != 0 {
		t.Fatalf("watch count after close = %d, want 0", got)
	}
}
