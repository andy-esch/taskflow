package store

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/andy-esch/taskflow/internal/domain"
)

// repositoryGuard supplies the same-process half of the repository guard and
// records the guarded-planner callback phase at the same canonical-root scope.
type repositoryGuard struct {
	write          sync.Mutex
	callbackMu     sync.RWMutex
	callbackActive bool
}

// repositoryGuards supplies the same-process half of the repository guard.
// Platform file-lock semantics vary for independently opened handles owned by one
// process, and a future long-lived adapter may create more than one *FS. The OS
// lock remains authoritative across processes; this keyed mutex makes the combined
// contract identical within one. CLI processes normally retain one entry; a future
// long-lived multi-space adapter should reference-count and evict idle entries.
var repositoryGuards = struct {
	sync.Mutex
	byRoot map[string]*repositoryGuard
}{byRoot: make(map[string]*repositoryGuard)}

func repositoryLockKey(root string) string {
	return normalizeRepositoryLockKey(root, runtime.GOOS == "windows")
}

func normalizeRepositoryLockKey(root string, caseInsensitive bool) string {
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	if absolute, err := filepath.Abs(root); err == nil {
		root = absolute
	}
	root = filepath.Clean(root)
	if caseInsensitive {
		root = strings.ToLower(root)
	}
	return root
}

func repositoryGuardFor(root string) *repositoryGuard {
	key := repositoryLockKey(root)
	repositoryGuards.Lock()
	guard := repositoryGuards.byRoot[key]
	if guard == nil {
		guard = new(repositoryGuard)
		repositoryGuards.byRoot[key] = guard
	}
	repositoryGuards.Unlock()
	return guard
}

func processRepositoryLock(root string) func() {
	guard := repositoryGuardFor(root)
	guard.write.Lock()
	return guard.write.Unlock
}

func (s *FS) enterRepositoryPlanner() (func(), error) {
	guard := repositoryGuardFor(s.root)
	guard.callbackMu.Lock()
	if guard.callbackActive {
		guard.callbackMu.Unlock()
		return nil, repositoryPlannerReentryError()
	}
	guard.callbackActive = true
	guard.callbackMu.Unlock()
	return func() {
		guard.callbackMu.Lock()
		guard.callbackActive = false
		guard.callbackMu.Unlock()
	}, nil
}

func repositoryPlannerReentryError() error {
	return fmt.Errorf("repository mutation planner is active; Store access from its callback or a concurrent caller is unavailable: %w", domain.ErrConflict)
}

func (s *FS) rejectRepositoryPlannerCall() error {
	guard := repositoryGuardFor(s.root)
	guard.callbackMu.RLock()
	active := guard.callbackActive
	guard.callbackMu.RUnlock()
	if active {
		return repositoryPlannerReentryError()
	}
	return nil
}

// writeLock is the ordinary Store mutation entry point. The platform lock owns
// cooperating-writer serialization. The process-local planner check turns an
// invalid callback re-entry into an attributable conflict instead of a self-deadlock.
func (s *FS) writeLock() (func(), error) {
	release, err := s.checkedWriteLock()
	if err != nil {
		return nil, err
	}
	return func() { _ = release() }, nil
}

// checkedWriteLock is the guarded-mutation form of writeLock: release failures
// are returned so a control-inverted boundary can attribute them to the operation.
// Ordinary legacy call sites retain their no-error unlock function until they are
// migrated; both paths use the same process + platform serialization.
func (s *FS) checkedWriteLock() (func() error, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return nil, err
	}
	unlockProcess := processRepositoryLock(s.root)
	unlockPlatform, err := s.platformWriteLockChecked()
	if err != nil {
		unlockProcess()
		return nil, err
	}
	return func() error {
		platformErr := unlockPlatform()
		unlockProcess()
		var hookErr error
		if testHookRepositoryUnlockError != nil {
			hookErr = testHookRepositoryUnlockError()
		}
		return errors.Join(platformErr, hookErr)
	}, nil
}

var testHookRepositoryUnlockError func() error
