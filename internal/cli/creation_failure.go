package cli

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/wire"
)

// createdCommandFailure adds CLI workspace context to a kind-specific core
// receipt whose write already committed. The existing created-item projection
// remains the machine vocabulary; callers must inspect before retrying.
type createdCommandFailure struct {
	cause     error
	created   wire.CreatedItem
	localPath string
	workspace wire.WorkspaceJSON
}

func committedCreateFailure(app *App, cause error, kind, id, slug, status, localPath string) error {
	return &createdCommandFailure{
		cause: cause,
		created: wire.CreatedItem{
			Kind: kind, ID: id, Slug: slug, Status: status, Path: app.rel(localPath),
		},
		localPath: localPath, workspace: app.workspace(),
	}
}

func (e *createdCommandFailure) Error() string {
	message := fmt.Sprintf("%s %q was created", e.created.Kind, e.created.ID)
	if e.localPath != "" {
		message += fmt.Sprintf(" at %q", e.localPath)
	}
	return fmt.Sprintf("%s, but finalization failed: %v; inspect the committed document before retrying", message, e.cause)
}

func (e *createdCommandFailure) Unwrap() error { return e.cause }
