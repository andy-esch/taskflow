package tui

import (
	"path/filepath"
	"testing"
)

// H3 (2026-06-22 audit): when no directory could be watched, newWatcher must
// return an error (closing the underlying watcher) so Run takes the watchOff
// branch and the footer honestly reports live-reload as unavailable — instead of
// returning a live-looking watcher that can never deliver an event.
func TestNewWatcher_NoWatchableDirsReturnsError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	w, err := newWatcher([]string{missing})
	if err == nil {
		if w != nil {
			_ = w.close()
		}
		t.Fatal("newWatcher must error when no directory could be watched")
	}

	// A real, watchable directory still succeeds.
	w2, err := newWatcher([]string{t.TempDir()})
	if err != nil {
		t.Fatalf("newWatcher should succeed on a real directory: %v", err)
	}
	_ = w2.close()
}

// An empty path set is also unwatchable — nothing to observe means live reload
// is off, not silently on.
func TestNewWatcher_EmptyPathsReturnsError(t *testing.T) {
	if w, err := newWatcher(nil); err == nil {
		if w != nil {
			_ = w.close()
		}
		t.Fatal("newWatcher(nil) must error: no directories to watch")
	}
}
