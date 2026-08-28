package store

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestTaskDependencyFieldsRoundTrip(t *testing.T) {
	root := t.TempDir()
	first, second := testutil.TaskID("first"), testutil.TaskID("second")
	content := "---\nid: " + testutil.TaskID("dependent") + "\nstatus: ready-to-start\n" +
		"depends_on: [" + second + ", " + first + "]\n" +
		"blocked_by: [legacy-a]\ndependencies: [legacy-b]\nblocks: [legacy-c]\n---\n# dependent\n"
	writeTask(t, root, "ready-to-start", "dependent.md", content)
	task, _, err := NewFS(root).GetTask("dependent")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(task.DependsOn, []string{second, first}) {
		t.Fatalf("reader must retain raw dependency evidence, got %v", task.DependsOn)
	}
	if !reflect.DeepEqual(task.LegacyBlockedBy, []string{"legacy-a"}) ||
		!reflect.DeepEqual(task.LegacyDependencies, []string{"legacy-b"}) ||
		!reflect.DeepEqual(task.LegacyBlocks, []string{"legacy-c"}) {
		t.Fatalf("legacy fields did not round-trip: %+v", task)
	}
	if !reflect.DeepEqual(task.LegacyDependencyFields, []string{"blocked_by", "dependencies", "blocks"}) {
		t.Fatalf("legacy field presence did not round-trip: %v", task.LegacyDependencyFields)
	}
}

func TestCreateTaskRejectsDependenciesUntilGuardedCreationExists(t *testing.T) {
	root := t.TempDir()
	first, second := testutil.TaskID("first"), testutil.TaskID("second")
	task := domain.Task{
		ID: testutil.TaskID("dependent"), Slug: "dependent", Status: domain.StatusReadyToStart,
		DependsOn: []string{second, first},
	}
	if _, err := NewFS(root).CreateTask(task, "# dependent\n", false); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "graph-owned") {
		t.Fatalf("unguarded create error = %v", err)
	}
	if entries, err := os.ReadDir(filepath.Join(root, "tasks")); err == nil && len(entries) != 0 {
		t.Fatalf("rejected create wrote files: %v", entries)
	}

	// Keep the serializer deterministic for the future guarded create primitive,
	// without exposing it through today's public store write.
	raw, err := buildFile(taskFields(task), "# dependent\n")
	if err != nil {
		t.Fatal(err)
	}
	want := "depends_on: [" + first + ", " + second + "]"
	if !strings.Contains(string(raw), want) {
		t.Fatalf("created dependency order is not stable; want %q\n%s", want, raw)
	}
}

func TestEditTaskRejectsDependencyDeltaButAllowsReordering(t *testing.T) {
	root := t.TempDir()
	first, second := testutil.TaskID("first"), testutil.TaskID("second")
	original := "---\nid: " + testutil.TaskID("dependent") + "\nstatus: ready-to-start\n" +
		"depends_on: [" + second + ", " + first + "]\n---\n# dependent\n"
	writeTask(t, root, "ready-to-start", "dependent.md", original)
	fs := NewFS(root)

	attempts := 0
	_, changed, err := fs.EditTask("dependent", bodyNow, func(current string, prevErr error) (string, error) {
		attempts++
		if attempts == 1 {
			return strings.Replace(current, second+", "+first, first, 1), nil
		}
		if prevErr == nil || !strings.Contains(prevErr.Error(), "guarded dependency") {
			t.Fatalf("reopened without guarded dependency direction: %v", prevErr)
		}
		return current, nil // give up on the rejected edit
	})
	if changed || !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("dependency delta changed=%v err=%v", changed, err)
	}

	reordered := strings.Replace(original, second+", "+first, first+", "+second, 1)
	_, changed, err = fs.EditTask("dependent", bodyNow, func(string, error) (string, error) { return reordered, nil })
	if err != nil || !changed {
		t.Fatalf("order-only edit changed=%v err=%v", changed, err)
	}
}

func TestEditTaskRejectsLegacyDependencyDeltaButAllowsReordering(t *testing.T) {
	root := t.TempDir()
	original := "---\nid: " + testutil.TaskID("dependent") + "\nstatus: ready-to-start\n" +
		"blocked_by: [legacy-b, legacy-a]\ndependencies: [legacy-c]\nblocks: [legacy-d]\n---\n# dependent\n"
	writeTask(t, root, "ready-to-start", "dependent.md", original)
	fs := NewFS(root)

	attempts := 0
	_, changed, err := fs.EditTask("dependent", bodyNow, func(current string, prevErr error) (string, error) {
		attempts++
		if attempts == 1 {
			return strings.Replace(current, "dependencies: [legacy-c]", "dependencies: [legacy-new]", 1), nil
		}
		if prevErr == nil || !strings.Contains(prevErr.Error(), "guarded dependency") {
			t.Fatalf("reopened without guarded dependency direction: %v", prevErr)
		}
		return current, nil
	})
	if changed || !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("legacy dependency delta changed=%v err=%v", changed, err)
	}

	reordered := strings.Replace(original, "legacy-b, legacy-a", "legacy-a, legacy-b", 1)
	_, changed, err = fs.EditTask("dependent", bodyNow, func(string, error) (string, error) { return reordered, nil })
	if err != nil || !changed {
		t.Fatalf("legacy order-only edit changed=%v err=%v", changed, err)
	}
}

func TestEditTaskMalformedDependencyCannotBeDeletedAsRepair(t *testing.T) {
	root := t.TempDir()
	taskID := testutil.TaskID("malformed-dependent")
	original := "---\nid: " + taskID + "\nstatus: ready-to-start\ndepends_on: " + testutil.TaskID("prerequisite") + "\n---\n# dependent\n"
	writeTask(t, root, "ready-to-start", "malformed-dependent.md", original)
	fs := NewFS(root)

	attempts := 0
	_, changed, err := fs.EditTask(taskID, bodyNow, func(current string, prevErr error) (string, error) {
		attempts++
		if attempts == 1 {
			return strings.Replace(current, "depends_on: "+testutil.TaskID("prerequisite")+"\n", "", 1), nil
		}
		if prevErr == nil || !strings.Contains(prevErr.Error(), "cannot verify") {
			t.Fatalf("malformed graph edit was not rejected clearly: %v", prevErr)
		}
		return current, nil
	})
	if changed || !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("malformed dependency deletion changed=%v err=%v", changed, err)
	}
	raw, readErr := os.ReadFile(filepath.Join(root, "tasks", taskID+"-malformed-dependent.md"))
	if readErr != nil || string(raw) != original {
		t.Fatalf("rejected edit changed source: err=%v\n%s", readErr, raw)
	}
}
