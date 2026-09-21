package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/andy-esch/taskflow/internal/cli/prompt"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

const (
	exitOK                = 0
	exitError             = 1
	exitNotFound          = 10
	exitValidation        = 11
	exitInvalidTransition = 12
	exitAmbiguous         = 13
	exitConflict          = 14
	exitAborted           = 130
)

// processExitCodes is the complete process-level contract: it owns every stable
// number, machine name, meaning, and whether this binary may emit the code. Keep
// the original four rows first and in order so a consumer comparing that
// historical prefix does not see gratuitous churn. Adding a row to this closed
// vocabulary is still NOT ADDITIVE under ADR-0008.
// Code 12 is deliberately published but reserved: invalid-transition was retired
// when task transitions stopped having a restricted matrix, and the number must
// not acquire a new meaning.
var processExitCodes = []struct {
	code    int
	name    string
	state   string
	meaning string
}{
	{exitNotFound, "not-found", wire.ExitCodeStateActive, "a requested named entity or registered planning space does not exist"},
	{exitValidation, "validation", wire.ExitCodeStateActive, "input or repository state failed validation"},
	{exitAmbiguous, "ambiguous", wire.ExitCodeStateActive, "a selector matched more than one entity"},
	{exitConflict, "conflict", wire.ExitCodeStateActive, "a write collided with existing or concurrently changed state"},
	{exitOK, "ok", wire.ExitCodeStateActive, "the command completed successfully, including an idempotent no-op"},
	{exitError, "error", wire.ExitCodeStateActive, "an unclassified command, filesystem, or flag-parsing failure occurred"},
	{exitAborted, "aborted", wire.ExitCodeStateActive, "an interactive prompt was aborted, normally with Ctrl-C"},
	{exitInvalidTransition, "invalid-transition", wire.ExitCodeStateReserved, "retired and reserved; this binary never emits it"},
}

// classifiedExitCodes is only the adapter mapping from an adapter-neutral domain
// failure class to a process outcome. Success, generic errors, prompt aborts, and
// retired reservations have no domain Class and therefore cannot distort
// domain.Classify or errors.Is behavior.
var classifiedExitCodes = []struct {
	class domain.Class
	code  int
}{
	{domain.ClassNotFound, exitNotFound},
	{domain.ClassValidation, exitValidation},
	{domain.ClassAmbiguous, exitAmbiguous},
	{domain.ClassConflict, exitConflict},
}

// classifiedExitCode selects a candidate process outcome. Keep the active-row
// enforcement in ExitCode: additions here cannot make a reserved or unknown
// number observable at the process boundary.
func classifiedExitCode(err error) int {
	if err == nil {
		return exitOK
	}
	if errors.Is(err, prompt.ErrAborted) {
		return exitAborted // 128 + SIGINT(2): the user interrupted a prompt with ctrl-c
	}
	class := domain.Classify(err)
	for _, e := range classifiedExitCodes {
		if e.class == class {
			return e.code
		}
	}
	return exitError
}

// activeExitCode accepts only published active outcomes. Falling back to the
// generic error is deliberate: an accidentally reused reservation or an
// unpublished number must not become a new process contract by accident.
func activeExitCode(code int) int {
	for _, e := range processExitCodes {
		if e.code == code && e.state == wire.ExitCodeStateActive {
			return code
		}
	}
	return exitError
}

// ExitCode maps an error to a semantic exit code, so agents can route on the
// code without parsing text. 0 also covers idempotent no-ops.
func ExitCode(err error) int {
	return activeExitCode(classifiedExitCode(err))
}

// errorCodeName is the stable machine name for an exit code — the `code` field
// of the --json error envelope. Same vocabulary as the exit codes, as words.
func errorCodeName(code int) string {
	code = activeExitCode(code)
	for _, e := range processExitCodes {
		if e.code == code && e.state == wire.ExitCodeStateActive {
			return e.name
		}
	}
	// ExitCode returns only registered outcomes; this fallback is defensive for a
	// future direct caller and still derives the generic name from the registry.
	for _, e := range processExitCodes {
		if e.code == exitError {
			return e.name
		}
	}
	return ""
}

// schemaExitCodes copies the CLI-owned process taxonomy into its neutral wire
// DTO. Keeping assembly here makes schema.go a consumer rather than a second
// transcription of the numbers, names, or reservation state.
func schemaExitCodes() []wire.SchemaExitCode {
	codes := make([]wire.SchemaExitCode, 0, len(processExitCodes))
	for _, e := range processExitCodes {
		codes = append(codes, wire.SchemaExitCode{
			Code: e.code, Name: e.name, State: e.state, Meaning: e.meaning,
		})
	}
	return codes
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
	// An OS failure otherwise arrives as code "error" plus prose, leaving an agent to
	// tell permission-denied from missing-directory from disk-full by reading English.
	payload.Error.Filesystem = filesystemDetails(err)
	var dependencyErr *dependencyCommandFailure
	if errors.As(err, &dependencyErr) {
		details := wire.ToDependencyMutationJSON(dependencyErr.receipt, dependencyErr.workspace)
		payload.Error.DependencyMutation = &details
	}
	var repairErr *graphRepairCommandFailure
	if errors.As(err, &repairErr) {
		details := wire.ToTaskGraphRepairJSON(repairErr.receipt, repairErr.workspace)
		payload.Error.GraphRepair = &details
	}
	var lifecycleErr *taskLifecycleCommandFailure
	if errors.As(err, &lifecycleErr) {
		details := wire.ToTaskLifecycleRecoveryJSON(lifecycleErr.receipt, lifecycleErr.workspace)
		payload.Error.TaskLifecycle = &details
	}
	var renameErr *taskRenameCommandFailure
	if errors.As(err, &renameErr) {
		details := wire.ToTaskRenameRecoveryJSON(renameErr.receipt, renameErr.workspace)
		payload.Error.TaskRename = &details
	}
	var threadErr *threadCreationCommandFailure
	if errors.As(err, &threadErr) {
		details := wire.ToThreadMutationJSON(threadErr.receipt, threadErr.path, threadErr.workspace)
		payload.Error.ThreadMutation = &details
	}
	var threadUpdateErr *threadMutationCommandFailure
	if errors.As(err, &threadUpdateErr) {
		details := wire.ToThreadUpdateJSON(threadUpdateErr.receipt, threadUpdateErr.path, threadUpdateErr.workspace)
		payload.Error.ThreadUpdate = &details
	}
	var threadPolicy *core.ThreadMutationPolicyError
	if errors.As(err, &threadPolicy) {
		details := wire.ToThreadMutationFailureJSON(threadPolicy)
		payload.Error.ThreadFailure = &details
	}
	var threadApplyErr *threadApplyCommandFailure
	if errors.As(err, &threadApplyErr) {
		details := wire.ToThreadApplyJSON(threadApplyErr.receipt, threadApplyErr.planPath, threadApplyErr.workspace)
		payload.Error.ThreadApply = &details
	}
	// Compact, like every other --json envelope (see wire.EncodeJSON): an agent
	// parsing the failure shouldn't pay for indentation either.
	_ = wire.EncodeJSON(w, payload)
}
