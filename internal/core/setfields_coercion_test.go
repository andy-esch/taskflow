package core_test

// These tests run the Service against the REAL store (not the fakeStore the
// in-package tests use) because the bug they guard is a YAML serialization
// round-trip: only the real markdown+yaml store can corrupt — and prove it
// doesn't. An external test package (core_test) avoids the store→core import
// cycle that an in-package test would hit.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
)

func setFieldsRepo(t *testing.T) *core.Service {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{filepath.Join("tasks", "ready-to-start"), "epics"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(rel, content string) {
		if err := os.WriteFile(filepath.Join(root, rel), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join("epics", "01-e.md"), "---\nstatus: planning\ndescription: e\n---\n# e\n")
	write(filepath.Join("tasks", "ready-to-start", "t.md"),
		"---\nstatus: ready-to-start\nepic: 01-e\ndescription: t\ntier: 3\n---\n# t\n")
	return core.NewService(store.NewFS(root))
}

// TestSetFields_CoercesTypedStringsThroughRoundTrip is the headline guard: a
// `--set tier=4` / `--set tags=a,b` (strings, from the escape hatch) must write
// native YAML types and reload cleanly — no FileProblem, no dropout from sweeps.
func TestSetFields_CoercesTypedStringsThroughRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name   string
		field  string
		value  string
		verify func(t *testing.T, task domain.Task)
	}{
		{"int tier", "tier", "4", func(t *testing.T, task domain.Task) {
			if task.Tier != 4 {
				t.Errorf("tier = %d, want 4", task.Tier)
			}
		}},
		{"list tags", "tags", "ui, chart", func(t *testing.T, task domain.Task) {
			if len(task.Tags) != 2 || task.Tags[0] != "ui" || task.Tags[1] != "chart" {
				t.Errorf("tags = %v, want [ui chart] (trimmed)", task.Tags)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := setFieldsRepo(t)
			if _, err := svc.SetFields("t", map[string]any{tc.field: tc.value}, false, false); err != nil {
				t.Fatalf("SetFields(%s=%q) rejected: %v", tc.field, tc.value, err)
			}
			// The write must reload cleanly through the strict loader.
			task, _, err := svc.ShowTask("t")
			if err != nil {
				t.Fatalf("task no longer reloads after set %s=%q (corrupted): %v", tc.field, tc.value, err)
			}
			tc.verify(t, task)
		})
	}
}

// TestSetFields_RejectsNonNumericTypedField keeps the failure clean: a typed
// field that can't be coerced is rejected up front (no write), not corrupted.
func TestSetFields_RejectsNonNumericTypedField(t *testing.T) {
	svc := setFieldsRepo(t)
	if _, err := svc.SetFields("t", map[string]any{"tier": "huge"}, false, false); err == nil {
		t.Fatal("want ErrValidation for a non-numeric tier")
	}
	if _, _, err := svc.ShowTask("t"); err != nil {
		t.Errorf("a rejected set must leave the task readable, got: %v", err)
	}
}

// TestSetFields_RejectsUnknownEpic mirrors NewTask: set can't orphan a task onto
// a non-existent epic.
func TestSetFields_RejectsUnknownEpic(t *testing.T) {
	svc := setFieldsRepo(t)
	if _, err := svc.SetFields("t", map[string]any{"epic": "bogus"}, false, false); err == nil {
		t.Fatal("want ErrValidation for an unknown epic")
	}
	// A real epic still passes.
	if _, err := svc.SetFields("t", map[string]any{"epic": "01-e"}, false, false); err != nil {
		t.Errorf("setting an existing epic should succeed, got: %v", err)
	}
}

// TestSetFields_UnsetRemovesKey pins `--unset` (decided 2026-06-12): the key
// is removed through the same validated atomic write path as assignment, and
// unsetting it again is an idempotent no-op.
func TestSetFields_UnsetRemovesKey(t *testing.T) {
	svc := setFieldsRepo(t)
	if _, err := svc.SetFields("t", map[string]any{"tier": domain.UnsetField{}}, false, false); err != nil {
		t.Fatal(err)
	}
	task, _, err := svc.ShowTask("t")
	if err != nil || task.Tier != 0 {
		t.Errorf("tier should be removed: %v tier=%d", err, task.Tier)
	}
	if _, err := svc.SetFields("t", map[string]any{"tier": domain.UnsetField{}}, false, false); err != nil {
		t.Errorf("unsetting an absent key should be a no-op, got %v", err)
	}
	// System fields can't be unset.
	for _, field := range []string{"status", "updated_at"} {
		if _, err := svc.SetFields("t", map[string]any{field: domain.UnsetField{}}, false, false); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("unset %s should be ErrValidation, got %v", field, err)
		}
	}
}

// TestSetFields_UnsetRejectsUnknownField guards the gate that the unset path
// once skipped: `--unset <typo>` without --force must fail like `--set <typo>`,
// not silently no-op. --force still lets a genuine custom field through.
func TestSetFields_UnsetRejectsUnknownField(t *testing.T) {
	svc := setFieldsRepo(t)
	if _, err := svc.SetFields("t", map[string]any{"descriptionn": domain.UnsetField{}}, false, false); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("unset of an unknown field should be ErrValidation, got %v", err)
	}
	if _, err := svc.SetFields("t", map[string]any{"custom": domain.UnsetField{}}, true, false); err != nil {
		t.Errorf("--force should allow unsetting a custom field, got %v", err)
	}
}

// TestSetFields_EpicDetach pins the D5 decision: an empty epic detaches the
// task (removes the key) instead of failing with `unknown epic ""`.
func TestSetFields_EpicDetach(t *testing.T) {
	svc := setFieldsRepo(t)
	if _, err := svc.SetFields("t", map[string]any{"epic": ""}, false, false); err != nil {
		t.Fatal(err)
	}
	task, _, err := svc.ShowTask("t")
	if err != nil || task.Epic != "" {
		t.Errorf("task should be detached from its epic: %v epic=%q", err, task.Epic)
	}
}

// TestSetFields_RejectsUpdatedAt pins the D6 decision: the stamp is a system
// field — an explicit value is rejected, not validated-then-clobbered.
func TestSetFields_RejectsUpdatedAt(t *testing.T) {
	svc := setFieldsRepo(t)
	if _, err := svc.SetFields("t", map[string]any{"updated_at": "2020-01-01"}, false, false); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("explicit updated_at should be ErrValidation, got %v", err)
	}
}

// TestSetFields_ListCoercionAgreesWithFix pins the M5 unification: writing a
// registry list field via `--set` produces the same form `lint --fix` wants,
// so the tool no longer generates drift its own fixer then repairs.
func TestSetFields_ListCoercionAgreesWithFix(t *testing.T) {
	svc := setFieldsRepo(t)
	if _, err := svc.SetFields("t", map[string]any{"related_tasks": "a, b"}, false, false); err != nil {
		t.Fatal(err)
	}
	results, err := svc.LintFix(true) // dry-run: report what WOULD change
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("set and fix disagree on the written form: %+v", results)
	}
}
