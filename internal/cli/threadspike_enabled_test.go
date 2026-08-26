//go:build threadspike

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestThreadSpikeTaggedCLIComposeApplyAndInspect(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "epic", "new", "Thread demo", "--description", "Throwaway Thread domain", "--tags", "threads")
	epicID := "01-thread-demo"
	for _, title := range []string{"External gate", "Schema", "Store"} {
		runRoot(t, "-C", root, "task", "new", title, "--epic", epicID, "--tags", "threads")
	}
	taskID := func(slug string) string {
		path := taskPath(t, root, slug)
		return strings.SplitN(filepath.Base(path), "-", 2)[0]
	}
	externalID := taskID("external-gate")
	schemaID := taskID("schema")
	storeID := taskID("store")
	manifestPath := filepath.Join(root, "demo.thread.yaml")
	planPath := filepath.Join(root, "demo.apply.yaml")
	mustWrite(t, manifestPath, fmt.Sprintf(`thread:
  title: Manual spike
  description: Exercise the tagged Thread CLI
  goal: Prove a throwaway workflow from compose through inspection.
  tags: [threads, spike]
nodes:
  - key: external
    task_id: %s
    member: false
  - key: schema
    task_id: %s
  - key: store
    task_id: %s
dependencies:
  - {from: external, to: schema}
  - {from: schema, to: store}
`, externalID, schemaID, storeID))

	compose := runRoot(t, "-C", root, "thread", "compose", "--from", manifestPath, "--out", planPath)
	if !strings.Contains(compose, "composed Thread") {
		t.Fatalf("compose output = %q", compose)
	}
	preview := runRoot(t, "-C", root, "thread", "apply", planPath, "--dry-run")
	if !strings.Contains(preview, "would-update") || !strings.Contains(preview, "would-create") {
		t.Fatalf("preview output = %q", preview)
	}
	if entries, err := os.ReadDir(filepath.Join(root, "threads")); err == nil && len(entries) != 0 {
		t.Fatalf("dry-run created Thread files: %v", entries)
	}
	apply := runRoot(t, "-C", root, "thread", "apply", planPath)
	if !strings.Contains(apply, "created thread") {
		t.Fatalf("apply output = %q", apply)
	}
	list := runRoot(t, "-C", root, "thread", "list")
	if !strings.Contains(list, "manual-spike") || !strings.Contains(list, "frontier:0") {
		t.Fatalf("list output = %q", list)
	}
	show := runRoot(t, "-C", root, "thread", "show", "manual-spike")
	if !strings.Contains(show, "external gates:") || !strings.Contains(show, "external-gate") || !strings.Contains(show, "gate:blocked") {
		t.Fatalf("show output = %q", show)
	}
	plan := runRoot(t, "-C", root, "thread", "plan", "manual-spike")
	if !strings.Contains(plan, "wave 1:") || !strings.Contains(plan, "wave 2:") || !strings.Contains(plan, "external gates:") {
		t.Fatalf("plan output = %q", plan)
	}
	retry := runRoot(t, "-C", root, "thread", "apply", planPath)
	if strings.Count(retry, "already-applied") != 1 {
		t.Fatalf("retry output = %q", retry)
	}
	content, err := os.ReadFile(taskPath(t, root, "store"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "depends_on: ["+schemaID+"]") {
		t.Fatalf("store dependency missing:\n%s", content)
	}
}
