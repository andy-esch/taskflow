package cli

import (
	"errors"
	"io/fs"
	"os"
	"syscall"

	"github.com/andy-esch/taskflow/internal/wire"
)

// filesystemDetails extracts the recovery-relevant shape of an OS failure, or nil when
// err did not come from the filesystem. It reads the structure Go already carries —
// *fs.PathError and *os.LinkError name the failing syscall and path — rather than
// matching on message text.
//
// The four classes below are the ones with genuinely different next actions:
// permission (grant access), not-found (create the parent / fix the path), read-only
// (remount or choose another tree), no-space (free space). Everything else is `io`,
// which says "the filesystem failed" without pretending to know more.
func filesystemDetails(err error) *wire.FilesystemErrorJSON {
	var (
		pathErr *fs.PathError
		linkErr *os.LinkError
	)
	details := &wire.FilesystemErrorJSON{}
	switch {
	case errors.As(err, &pathErr):
		details.Operation, details.Path = pathErr.Op, pathErr.Path
	case errors.As(err, &linkErr):
		// A rename carries two paths; New is where the write was headed, which is the
		// one a caller acts on.
		details.Operation, details.Path = linkErr.Op, linkErr.New
	default:
		return nil
	}
	details.Class, details.Retryable = classifyFilesystemError(err)
	return details
}

func classifyFilesystemError(err error) (class string, retryable bool) {
	switch {
	case errors.Is(err, fs.ErrPermission):
		return "permission", false
	case errors.Is(err, fs.ErrNotExist):
		return "not-found", false
	case errors.Is(err, syscall.EROFS):
		return "read-only", false
	case errors.Is(err, syscall.ENOSPC), errors.Is(err, syscall.EDQUOT):
		return "no-space", false
	// Genuinely transient: the same call, unchanged, can succeed on a later attempt.
	// Kept narrow — a retry loop on anything above would spin against a condition only
	// an operator can clear.
	case errors.Is(err, syscall.EAGAIN), errors.Is(err, syscall.EINTR), errors.Is(err, syscall.EBUSY):
		return "io", true
	default:
		return "io", false
	}
}
