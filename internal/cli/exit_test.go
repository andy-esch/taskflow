package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"testing"

	"github.com/andy-esch/taskflow/internal/cli/prompt"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestProcessExitTaxonomyIsCompleteAndStable(t *testing.T) {
	want := []wire.SchemaExitCode{
		{Code: 10, Name: "not-found", State: wire.ExitCodeStateActive, Meaning: "a requested named entity or registered planning space does not exist"},
		{Code: 11, Name: "validation", State: wire.ExitCodeStateActive, Meaning: "input or repository state failed validation"},
		{Code: 13, Name: "ambiguous", State: wire.ExitCodeStateActive, Meaning: "a selector matched more than one entity"},
		{Code: 14, Name: "conflict", State: wire.ExitCodeStateActive, Meaning: "a write collided with existing or concurrently changed state"},
		{Code: 0, Name: "ok", State: wire.ExitCodeStateActive, Meaning: "the command completed successfully, including an idempotent no-op"},
		{Code: 1, Name: "error", State: wire.ExitCodeStateActive, Meaning: "an unclassified command, filesystem, or flag-parsing failure occurred"},
		{Code: 130, Name: "aborted", State: wire.ExitCodeStateActive, Meaning: "an interactive prompt was aborted, normally with Ctrl-C"},
		{Code: 12, Name: "invalid-transition", State: wire.ExitCodeStateReserved, Meaning: "retired and reserved; this binary never emits it"},
	}
	if got := schemaExitCodes(); !reflect.DeepEqual(got, want) {
		t.Fatalf("published process taxonomy changed:\n got: %+v\nwant: %+v", got, want)
	}

	codes := make(map[int]wire.SchemaExitCode, len(want))
	names := make(map[string]int, len(want))
	for _, row := range schemaExitCodes() {
		if _, exists := codes[row.Code]; exists {
			t.Errorf("duplicate process exit code %d", row.Code)
		}
		if prior, exists := names[row.Name]; exists {
			t.Errorf("duplicate process exit name %q at codes %d and %d", row.Name, prior, row.Code)
		}
		codes[row.Code] = row
		names[row.Name] = row.Code
	}
	for _, mapping := range classifiedExitCodes {
		row, ok := codes[mapping.code]
		if !ok || row.State != wire.ExitCodeStateActive {
			t.Errorf("domain class %v maps to unpublished or inactive code %d", mapping.class, mapping.code)
		}
	}
}

func TestExitCodeAndMachineNameUsePublishedActiveOutcomes(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantCode int
		wantName string
	}{
		{"success", nil, 0, "ok"},
		{"generic", errors.New("boom"), 1, "error"},
		{"not found", fmt.Errorf("task: %w", domain.ErrNotFound), 10, "not-found"},
		{"validation", fmt.Errorf("input: %w", domain.ErrValidation), 11, "validation"},
		{"ambiguous", fmt.Errorf("selector: %w", domain.ErrAmbiguous), 13, "ambiguous"},
		{"conflict", fmt.Errorf("write: %w", domain.ErrConflict), 14, "conflict"},
		{"aborted", prompt.ErrAborted, 130, "aborted"},
		{"filesystem", &fs.PathError{Op: "open", Path: "/blocked", Err: fs.ErrPermission}, 1, "error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code := ExitCode(tc.err)
			if code != tc.wantCode {
				t.Fatalf("ExitCode(%v) = %d, want %d", tc.err, code, tc.wantCode)
			}
			if name := errorCodeName(code); name != tc.wantName {
				t.Fatalf("errorCodeName(%d) = %q, want %q", code, name, tc.wantName)
			}
		})
	}
	for _, candidate := range []int{exitInvalidTransition, 99} {
		if got := activeExitCode(candidate); got != exitError {
			t.Errorf("activeExitCode(%d) = %d, want generic error %d", candidate, got, exitError)
		}
		if got := errorCodeName(candidate); got != "error" {
			t.Errorf("errorCodeName(%d) = %q, want error", candidate, got)
		}
	}
}
