package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/domain"
)

// ExitCode maps an error to a semantic exit code, so agents can route on the
// code without parsing text. 0 also covers idempotent no-ops.
func ExitCode(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, domain.ErrNotFound):
		return 10
	case errors.Is(err, domain.ErrValidation):
		return 11
	// 12 (invalid-transition) is retired but reserved — see domain/errors.go.
	case errors.Is(err, domain.ErrAmbiguous):
		return 13
	case errors.Is(err, domain.ErrConflict):
		return 14
	default:
		return 1
	}
}

// errorCodeName is the stable machine name for an exit code — the `code` field
// of the --json error envelope. Same vocabulary as the exit codes, as words.
func errorCodeName(code int) string {
	switch code {
	case 10:
		return "not-found"
	case 11:
		return "validation"
	case 13:
		return "ambiguous"
	case 14:
		return "conflict"
	default:
		return "error"
	}
}

// WriteError reports a fatal error on w: prose normally, a versioned JSON
// envelope when the run was --json — an agent driving --json must never have
// to parse prose to learn why a command failed (decided 2026-06-12). It goes
// to stderr either way; stdout stays empty on failure.
func WriteError(w io.Writer, err error, asJSON bool) {
	if !asJSON {
		fmt.Fprintln(w, "error:", err)
		return
	}
	var payload struct {
		SchemaVersion string `json:"schema_version"`
		Error         struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	payload.SchemaVersion = render.SchemaVersion
	payload.Error.Code = errorCodeName(ExitCode(err))
	payload.Error.Message = err.Error()
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(payload)
}
