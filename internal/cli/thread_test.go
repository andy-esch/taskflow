package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

func threadCLIRepo(t *testing.T) string {
	t.Helper()
	repo := testutil.NewRepo(t)
	externalID := testutil.TaskID("external")
	for _, task := range []struct {
		slug      string
		status    domain.Status
		dependsOn string
	}{
		{"external", domain.StatusCompleted, ""},
		{"alpha", domain.StatusReadyToStart, externalID},
		{"beta", domain.StatusReadyToStart, testutil.TaskID("alpha")},
	} {
		dependsOn := ""
		if task.dependsOn != "" {
			dependsOn = "depends_on: [" + task.dependsOn + "]\n"
		}
		repo.Task(string(task.status), task.slug+".md", "---\n"+
			"id: "+testutil.TaskID(task.slug)+"\n"+
			"status: "+string(task.status)+"\n"+
			"description: "+task.slug+"\n"+
			"tags: [threads]\n"+dependsOn+
			"---\n# "+task.slug+"\n")
	}
	return repo.Root
}

func threadPath(t *testing.T, root, slug string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(root, domain.ThreadsDir, "*-"+slug+".md"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("Thread path for %s: matches=%v err=%v", slug, matches, err)
	}
	return matches[0]
}

func TestThreadNewListShowPathAndFrontier(t *testing.T) {
	root := threadCLIRepo(t)
	out, errOut, err := runIn(t, root, "thread", "new", "Delivery", "--description", "Ship Threads", "--goal", "Dogfood Threads", "--task", "beta", "--task", "alpha", "--json")
	if err != nil || errOut != "" {
		t.Fatalf("thread new: %v\nstdout=%s\nstderr=%s", err, out, errOut)
	}
	var created wire.ThreadMutationEnvelope
	if err := json.Unmarshal([]byte(out), &created); err != nil {
		t.Fatalf("decode creation: %v\n%s", err, out)
	}
	wantMembers := []string{testutil.TaskID("alpha"), testutil.TaskID("beta")}
	if created.Thread.Status != "unstarted" || !created.Changed || !created.Committed || !slices.Equal(created.Thread.Tasks, wantMembers) {
		t.Fatalf("creation = %+v", created)
	}
	content, err := os.ReadFile(threadPath(t, root, "delivery"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "tasks: ["+strings.Join(wantMembers, ", ")+"]") {
		t.Fatalf("persisted membership is not canonical:\n%s", content)
	}

	out, errOut, err = runIn(t, root, "thread", "list", "--json")
	if err != nil || errOut != "" {
		t.Fatalf("thread list: %v\nstdout=%s\nstderr=%s", err, out, errOut)
	}
	var listed wire.ThreadsEnvelope
	if err := json.Unmarshal([]byte(out), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Threads) != 1 {
		t.Fatalf("list = %+v", listed)
	}
	view := listed.Threads[0]
	if view.Rollup.Total != 2 || view.Rollup.Done != 0 || view.Rollup.Drained != 0 ||
		view.GraphHealth != "healthy" || view.ProjectionHealth != "healthy" {
		t.Fatalf("rollup = %+v graph=%s projection=%s", view.Rollup, view.GraphHealth, view.ProjectionHealth)
	}
	if len(view.ExternalGates) != 1 || view.ExternalGates[0].Task.TaskID != testutil.TaskID("external") || view.ExternalGates[0].Outstanding {
		t.Fatalf("external gates = %+v", view.ExternalGates)
	}
	if len(view.Frontier) != 1 || view.Frontier[0].Task.TaskID != testutil.TaskID("alpha") {
		t.Fatalf("frontier = %+v", view.Frontier)
	}

	show := runRoot(t, "-C", root, "thread", "show", "delivery", "--raw")
	if !strings.Contains(show, "Members") || !strings.Contains(show, "External gates") || !strings.Contains(show, "# Thread: Delivery") {
		t.Fatalf("show output:\n%s", show)
	}
	frontier := runRoot(t, "-C", root, "thread", "frontier", "delivery")
	if !strings.Contains(frontier, "alpha") || strings.Contains(frontier, "beta  ") {
		t.Fatalf("frontier output:\n%s", frontier)
	}
	gotPath := strings.TrimSpace(runRoot(t, "-C", root, "thread", "path", "delivery"))
	if filepath.Base(gotPath) != filepath.Base(threadPath(t, root, "delivery")) {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestThreadNewDryRunAndInvalidMember(t *testing.T) {
	root := threadCLIRepo(t)
	out, _, err := runIn(t, root, "thread", "new", "Preview", "--description", "Preview Thread", "--goal", "Prove preview", "--task", "alpha", "--dry-run", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var preview wire.ThreadMutationEnvelope
	if err := json.Unmarshal([]byte(out), &preview); err != nil {
		t.Fatal(err)
	}
	if !preview.DryRun || !preview.Changed || preview.Committed {
		t.Fatalf("preview = %+v", preview)
	}
	if matches, _ := filepath.Glob(filepath.Join(root, domain.ThreadsDir, "*-preview.md")); len(matches) != 0 {
		t.Fatalf("dry-run wrote %v", matches)
	}
	if _, _, err := runIn(t, root, "thread", "new", "Bad", "--description", "Bad Thread", "--goal", "Fail", "--task", "missing"); err == nil {
		t.Fatal("missing initial member must fail creation")
	}
}

func TestLintReportsMalformedThreadMembership(t *testing.T) {
	root := threadCLIRepo(t)
	threadID := testutil.TaskID("thread")
	missingID := testutil.TaskID("missing")
	mustWrite(t, filepath.Join(root, domain.ThreadsDir, threadID+"-bad-thread.md"), "---\n"+
		"id: "+threadID+"\nstatus: unstarted\ndescription: bad\ngoal: repair it\ncreated: \"2026-08-29\"\n"+
		"tasks: ["+missingID+"]\n---\n# Bad Thread\n")
	out, err := runRootRC(t, "-C", root, "lint", "--json")
	if err == nil {
		t.Fatal("missing Thread member must make ordinary lint fail")
	}
	var lint wire.LintEnvelope
	if decodeErr := json.Unmarshal([]byte(out), &lint); decodeErr != nil {
		t.Fatalf("decode lint: %v\n%s", decodeErr, out)
	}
	found := false
	for _, item := range lint.Issues {
		if item.Slug == "bad-thread" {
			for _, issue := range item.Issues {
				found = found || issue.Field == "tasks" && strings.Contains(issue.Message, "unknown member")
			}
		}
	}
	if !found {
		t.Fatalf("Thread membership issue missing: %+v", lint)
	}
}

func TestInitScaffoldsThreadsWithoutProjects(t *testing.T) {
	root := freshRepo(t)
	if info, err := os.Stat(filepath.Join(root, domain.ThreadsDir)); err != nil || !info.IsDir() {
		t.Fatalf("threads scaffold: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, domain.ProjectsDir)); !os.IsNotExist(err) {
		t.Fatalf("fresh init unexpectedly scaffolded projects/: %v", err)
	}
}

func TestInitReportsPreservedLegacyProjectsContent(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, domain.ProjectsDir, "legacy.md")
	mustWrite(t, legacy, "# Legacy Project\n")
	out := runRoot(t, "init", "--path", root, "--no-register", "--json")
	var envelope wire.InitEnvelope
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("decode init: %v\n%s", err, out)
	}
	if envelope.LegacyProjects != filepath.Dir(legacy) || envelope.LegacyProjectsRemedy == "" || len(envelope.Removed) != 0 {
		t.Fatalf("init legacy report = %+v", envelope)
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Fatalf("legacy content was not preserved: %v", err)
	}
}

func TestBareInitReportsAvailableThreadScaffoldRepair(t *testing.T) {
	root := freshRepo(t)
	if err := os.RemoveAll(filepath.Join(root, domain.ThreadsDir)); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(root, domain.ProjectsDir, ".gitkeep"), "")

	out := runRoot(t, "init", "--path", root, "--no-register")
	if !strings.Contains(out, "scaffold repair available") ||
		!strings.Contains(out, `tskflwctl init --taskflow-root "."`) {
		t.Fatalf("bare init omitted repair guidance:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(root, domain.ThreadsDir)); !os.IsNotExist(err) {
		t.Fatalf("bare init mutated the scaffold: %v", err)
	}

	jsonOut := runRoot(t, "init", "--path", root, "--no-register", "--json")
	var envelope wire.InitEnvelope
	if err := json.Unmarshal([]byte(jsonOut), &envelope); err != nil {
		t.Fatal(err)
	}
	if !envelope.ScaffoldRepairAvailable || envelope.ScaffoldRepairCommand == "" {
		t.Fatalf("repair receipt = %+v", envelope)
	}
}

func TestInitLegacyProjectsOnlyUsesNeutralHumanReceipt(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, domain.ProjectsDir, "legacy.md"), "# Legacy\n")

	out := runRoot(t, "init", "--path", root, "--taskflow-root", ".", "--no-register")
	if strings.Contains(out, "✔ updated") || !strings.Contains(out, "already initialized") ||
		!strings.Contains(out, "preserved legacy Projects content") {
		t.Fatalf("legacy-only init receipt:\n%s", out)
	}
}

func TestThreadCreationCommittedFailureHasStructuredRecovery(t *testing.T) {
	receipt := core.ThreadCreationReceipt{
		Thread: domain.Thread{ID: "6g3q4rtmv4ak", Slug: "delivery", Status: domain.ThreadStatusUnstarted,
			Description: "delivery", Goal: "ship", Created: "2026-08-29"},
		Changed: true, Committed: true,
	}
	cause := &core.ThreadCreationMutationFailure{Cause: domain.ErrConflict, Receipt: receipt}
	err := &threadCreationCommandFailure{
		cause: cause, receipt: receipt, path: "threads/6g3q4rtmv4ak-delivery.md",
		workspace: wire.WorkspaceJSON{PlanningRoot: "/repo/planning", Source: wire.WorkspaceSourceConfig},
	}
	var out bytes.Buffer
	WriteError(&out, err, true)
	var envelope wire.ErrorEnvelope
	if decodeErr := json.Unmarshal(out.Bytes(), &envelope); decodeErr != nil {
		t.Fatalf("decode recovery: %v\n%s", decodeErr, out.String())
	}
	if envelope.Error.Code != "conflict" || envelope.Error.ThreadMutation == nil ||
		!envelope.Error.ThreadMutation.Committed || envelope.Error.ThreadMutation.Thread.ID != receipt.Thread.ID ||
		envelope.Error.ThreadMutation.Path != "threads/6g3q4rtmv4ak-delivery.md" {
		t.Fatalf("recovery = %+v", envelope)
	}
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatal("command recovery wrapper lost error classification")
	}
}
