package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/domain"
)

// readSurface identifies one independently reloadable asynchronous projection.
// It deliberately describes adapter surfaces rather than storage: the same
// bounded contention policy can therefore serve entity tabs, the dashboard, and
// the future Thread UI without teaching the reducer about filesystem details.
type readSurface uint8

const (
	readEntityList readSurface = iota
	readEntityDetail
	readDashboard
	readThreadList
	readThreadDetail
	readSurfaceCount
)

// readRequest is the stable identity of one asynchronous read generation. A
// retry is valid only while this request is still current; any later reload or
// selection change supersedes it before it can touch the model.
type readRequest struct {
	surface readSurface
	kind    entityKind
	id      string
	gen     int
}

// readResult makes participation in the contention policy explicit beside a
// result message's declaration. Load results without an error deliberately do
// not implement it and pass through untouched.
type readResult interface {
	readError() error
}

// withReadConflictRetry is the shared transient-contention policy for reads. A
// guarded repository mutation excludes concurrent Store access for the width of
// its planner window (domain.ErrConflict), so a watcher refresh that lands
// inside one has not found a broken repository — it has found a busy one, and
// the model it already holds is still the last coherent read.
//
// Only the FIRST conflict per request is held. retry is deliberately the
// unwrapped command, so its own result is final for this generation: a
// repository that stays contended surfaces an ordinary durable error instead of
// spinning. A read whose message type carries no error (readMessageError)
// simply opts out.
//
// This is the read half of conflict handling. A conflict from a MUTATION means
// the file changed under a compare-and-swap, which is not transient at all —
// that path flashes and reloads to show the current state (see actionErrMsg).
func withReadConflictRetry(request readRequest, read tea.Cmd) tea.Cmd {
	if read == nil {
		return nil
	}
	return func() tea.Msg {
		msg := read()
		if domain.Classify(readMessageError(msg)) == domain.ClassConflict {
			return readConflictMsg{request: request, retry: read}
		}
		return msg
	}
}

// readMessageError extracts the failure from a result that opted into the
// shared policy at its declaration site.
func readMessageError(msg tea.Msg) error {
	result, ok := msg.(readResult)
	if !ok {
		return nil
	}
	return result.readError()
}

// deferReadRetry waits one filesystem quiet period before re-reading — the same
// window the watcher coalesces a save-storm into, which is also about as long as
// a planner window takes to close. Retrying sooner would mostly re-enter it.
func deferReadRetry(request readRequest, retry tea.Cmd) tea.Cmd {
	return tea.Tick(fsDebounce, func(time.Time) tea.Msg {
		return readRetryMsg{request: request, retry: retry}
	})
}
