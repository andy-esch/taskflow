package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/andy-esch/taskflow/internal/cli/prompt"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

// errCodes is the CLI's error policy: it ties each domain error Class to its exit
// code and the stable machine name for the --json envelope. The *classification*
// (which Class an error is) now lives in domain.Classify, shared with the TUI and a
// future web adapter (audit H4); this table is only the CLI-specific Class → code +
// name mapping. ExitCode and errorCodeName both read it, so the code and its name
// can't drift apart, and schema.go iterates it for the `exit_codes` contract — so
// the order and names are part of the wire golden and must not change.
// 12 (invalid-transition) is retired but reserved — see domain/errors.go.
var errCodes = []struct {
	class domain.Class
	code  int
	name  string
}{
	{domain.ClassNotFound, 10, "not-found"},
	{domain.ClassValidation, 11, "validation"},
	{domain.ClassAmbiguous, 13, "ambiguous"},
	{domain.ClassConflict, 14, "conflict"},
}

// ExitCode maps an error to a semantic exit code, so agents can route on the
// code without parsing text. 0 also covers idempotent no-ops.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, prompt.ErrAborted) {
		return 130 // 128 + SIGINT(2): the user interrupted a prompt with ctrl-c
	}
	class := domain.Classify(err)
	for _, e := range errCodes {
		if e.class == class {
			return e.code
		}
	}
	return 1
}

// errorCodeName is the stable machine name for an exit code — the `code` field
// of the --json error envelope. Same vocabulary as the exit codes, as words.
func errorCodeName(code int) string {
	for _, e := range errCodes {
		if e.code == code {
			return e.name
		}
	}
	return "error"
}

// WriteError reports a fatal error on w: prose normally, a versioned JSON
// envelope when the run was --json — an agent driving --json must never have
// to parse prose to learn why a command failed (decided 2026-06-12). It goes
// to stderr either way; stdout stays empty on failure.
func WriteError(w io.Writer, err error, asJSON bool) {
	// A prompt abort only reaches here on a TTY (the gate is closed under --json),
	// so it's the human path: a quiet line, never a scary "error:" or a JSON
	// envelope. Pairs with exit code 130.
	if errors.Is(err, prompt.ErrAborted) {
		fmt.Fprintln(w, "aborted")
		return
	}
	if !asJSON {
		fmt.Fprintln(w, "error:", err)
		return
	}
	payload := wire.ErrorEnvelope{SchemaVersion: wire.SchemaVersion}
	payload.Error.Code = errorCodeName(ExitCode(err))
	payload.Error.Message = err.Error()
	// Compact, like every other --json envelope (see wire.EncodeJSON): an agent
	// parsing the failure shouldn't pay for indentation either.
	_ = wire.EncodeJSON(w, payload)
}
