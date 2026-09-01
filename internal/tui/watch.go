package tui

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fsnotify/fsnotify"
)

// fsDebounce is the quiet period a change must be followed by before the TUI
// reloads — long enough to coalesce an editor's write/rename/chmod storm into a
// single refresh, short enough to feel live.
const fsDebounce = 200 * time.Millisecond

type watchHealth uint8

const (
	watchHealthy watchHealth = iota
	watchDegraded
	watchUnavailable
)

// watcher owns the desired layout for the active planning space, not merely the
// directories that happened to exist when it started. Direct leaf watches catch
// entity writes; nearest-existing-parent sentinels catch a missing leaf appearing
// and an attached leaf being replaced. It is replaced after a successful atlas
// navigation and closed as that workspace becomes inactive.
type watcher struct {
	fsw *fsnotify.Watcher

	mu       sync.Mutex
	rawPaths []string
	desired  []string
	attached map[string]os.FileInfo
	// addPath is bound to fsnotify.Add; the seam makes transient attachment
	// failure and recovery deterministic in tests without replacing the backend.
	addPath func(string) error
	health  watchHealth
	closed  bool
}

// newWatcher normalizes and de-duplicates the Layout port's desired leaf paths,
// then attaches both present leaves and parent sentinels. A missing leaf is a
// degraded but recoverable state when an existing parent can be watched. A valid
// desired set whose initial Add calls all fail retains an unavailable watcher so
// manual refresh can retry; only an invalid/unbounded set or backend construction
// failure returns no watcher.
func newWatcher(paths []string) (*watcher, error) {
	return newWatcherWithPathAdder(paths, nil)
}

func newWatcherWithPathAdder(paths []string, addPath func(string) error) (*watcher, error) {
	rawPaths := cleanWatchPaths(paths)
	if len(rawPaths) == 0 {
		return nil, errors.New("no watchable directories: live reload unavailable")
	}
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &watcher{
		fsw: fsw, rawPaths: rawPaths, attached: make(map[string]os.FileInfo),
		addPath: addPath, health: watchUnavailable,
	}
	if w.addPath == nil {
		w.addPath = fsw.Add
	}
	w.reconcile()
	return w, nil
}

func (w *watcher) watchHealth() watchHealth {
	if w == nil {
		return watchUnavailable
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.health
}

// reconcile makes the concrete fsnotify set converge on the current desired
// layout. File identity, not path spelling alone, decides whether an attachment
// is still current so an atomically replaced directory is re-added.
func (w *watcher) reconcile() watchHealth {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		w.health = watchUnavailable
		return w.health
	}

	w.desired = normalizeWatchPaths(w.rawPaths)
	required := make(map[string]struct{}, len(w.desired)+1)
	complete := len(w.desired) > 0
	for _, desired := range w.desired {
		if info, err := os.Stat(desired); err == nil && info.IsDir() {
			required[desired] = struct{}{}
		} else {
			complete = false
		}
		parent, ok := nearestExistingWatchDirectory(filepath.Dir(desired))
		if !ok {
			complete = false
			continue
		}
		required[parent] = struct{}{}
	}
	symlinkSentinels, hasSymlink := symlinkParentSentinels(w.rawPaths)
	if hasSymlink {
		// fsnotify follows symlinks, and some backends do not report a same-name
		// lexical retarget through the parent directory. Keep ordinary target
		// changes live, but report the portable retarget blind spot honestly.
		complete = false
	}
	for _, sentinel := range symlinkSentinels {
		required[sentinel] = struct{}{}
	}

	for path, previous := range w.attached {
		_, stillRequired := required[path]
		current, err := os.Stat(path)
		if stillRequired && err == nil && current.IsDir() && os.SameFile(previous, current) {
			continue
		}
		_ = w.fsw.Remove(path)
		delete(w.attached, path)
	}

	targets := make([]string, 0, len(required))
	for path := range required {
		targets = append(targets, path)
	}
	sort.Strings(targets)
	for _, path := range targets {
		if _, ok := w.attached[path]; ok {
			continue
		}
		before, err := os.Stat(path)
		if err != nil || !before.IsDir() {
			continue
		}
		if err := w.addPath(path); err != nil {
			continue
		}
		after, err := os.Stat(path)
		if err != nil || !after.IsDir() || !os.SameFile(before, after) {
			_ = w.fsw.Remove(path)
			continue
		}
		w.attached[path] = after
	}

	if len(w.attached) == 0 {
		w.health = watchUnavailable
		return w.health
	}
	for path := range required {
		if _, ok := w.attached[path]; !ok {
			complete = false
		}
	}
	if complete {
		w.health = watchHealthy
	} else {
		w.health = watchDegraded
	}
	return w.health
}

func cleanWatchPaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		absolute, err := filepath.Abs(filepath.Clean(path))
		if err != nil {
			continue
		}
		if absolute == filepath.Dir(absolute) {
			continue
		}
		seen[absolute] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for path := range seen {
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

func normalizeWatchPaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	for _, path := range cleanWatchPaths(paths) {
		canonical := canonicalWatchPath(path)
		if canonical == filepath.Dir(canonical) {
			continue
		}
		seen[canonical] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for path := range seen {
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

// canonicalWatchPath resolves the deepest existing prefix, then restores any
// missing suffix. That collapses symlink aliases even before a desired leaf has
// been created, without requiring the leaf itself to exist.
func canonicalWatchPath(path string) string {
	probe := filepath.Clean(path)
	missing := make([]string, 0, 2)
	for {
		if resolved, err := filepath.EvalSymlinks(probe); err == nil {
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return filepath.Clean(resolved)
		}
		parent := filepath.Dir(probe)
		if parent == probe {
			return filepath.Clean(path)
		}
		missing = append(missing, filepath.Base(probe))
		probe = parent
	}
}

// symlinkParentSentinels watches the directory entry that names each symlink in
// a raw Layout path where the backend exposes such changes. Direct watches follow
// symlinks to their targets; because same-name retarget delivery is not portable,
// any such path remains degraded and manual refresh re-resolves it.
func symlinkParentSentinels(paths []string) ([]string, bool) {
	seen := make(map[string]struct{})
	hasSymlink := false
	cleaned := cleanWatchPaths(paths)
	boundary := commonWatchBoundary(cleaned)
	for _, path := range cleaned {
		pathBoundary := boundary
		if pathBoundary == filepath.Dir(pathBoundary) {
			// Disjoint layouts have no useful shared lexical boundary. Limit each
			// scan to its leaf parent rather than treating stable operating-system
			// aliases such as macOS /var as planning-space degradation.
			pathBoundary = filepath.Dir(path)
		}
		for probe := filepath.Clean(path); ; probe = filepath.Dir(probe) {
			info, err := os.Lstat(probe)
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				hasSymlink = true
				parent := canonicalWatchPath(filepath.Dir(probe))
				if parent != filepath.Dir(parent) {
					seen[parent] = struct{}{}
				}
			}
			if probe == pathBoundary || probe == filepath.Dir(probe) {
				break
			}
		}
	}
	out := make([]string, 0, len(seen))
	for path := range seen {
		out = append(out, path)
	}
	sort.Strings(out)
	return out, hasSymlink
}

// commonWatchBoundary returns the lowest lexical directory shared by the
// desired leaves. Symlinks above it belong to the host path namespace, not to
// the planning layout whose retargeting this watcher can usefully diagnose.
func commonWatchBoundary(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	boundary := filepath.Dir(paths[0])
	for _, path := range paths[1:] {
		for !pathWithin(boundary, path) && boundary != filepath.Dir(boundary) {
			boundary = filepath.Dir(boundary)
		}
	}
	return boundary
}

func pathWithin(directory, path string) bool {
	relative, err := filepath.Rel(directory, path)
	return err == nil && relative != ".." && !filepath.IsAbs(relative) &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

// nearestExistingWatchDirectory deliberately refuses a filesystem root. Watching
// an entire volume because a Layout supplied no useful ancestor would turn one
// broken planning path into an unbounded event source.
func nearestExistingWatchDirectory(path string) (string, bool) {
	for candidate := filepath.Clean(path); ; candidate = filepath.Dir(candidate) {
		if candidate == filepath.Dir(candidate) {
			return "", false
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, true
		}
	}
}

func (w *watcher) close() error {
	if w == nil || w.fsw == nil {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.closed = true
	w.health = watchUnavailable
	w.attached = nil
	return w.fsw.Close()
}

// waitForFS blocks until the next filesystem change, reconciles the desired
// attachment set, and returns the resulting health with the reload nudge. The
// model re-issues it after each event while recovery remains possible.
func waitForFS(w *watcher) tea.Cmd {
	if w == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case _, ok := <-w.fsw.Events:
			if !ok {
				return nil
			}
			return fsEventMsg{health: w.reconcile()}
		case _, ok := <-w.fsw.Errors:
			if !ok {
				return nil
			}
			return fsEventMsg{health: w.reconcile()}
		}
	}
}

func reconcileWatcher(w *watcher) tea.Cmd {
	if w == nil {
		return nil
	}
	return func() tea.Msg { return watcherReconciledMsg{health: w.reconcile()} }
}

// debounceTick performs one final reconciliation after the quiet period. This
// closes the race where a nested directory tree is created faster than parent
// sentinels can be attached one level at a time.
func debounceTick(w *watcher, gen int) tea.Cmd {
	return tea.Tick(fsDebounce, func(time.Time) tea.Msg {
		health := watchHealthy
		if w != nil {
			health = w.reconcile()
		}
		return debounceMsg{gen: gen, health: health}
	})
}
