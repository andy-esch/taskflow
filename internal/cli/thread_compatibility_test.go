package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/wire"
)

const threadCompatibilityFixtureRoot = "testdata/thread_compatibility"

var threadPreviewReleases = []struct {
	name           string
	wireVersion    string
	threadFile     string
	listGolden     string
	showGolden     string
	graphGolden    string
	frontierGolden string
	pathGolden     string
	planGolden     string
}{
	{
		name: "v0.18.0", wireVersion: "1.58",
		threadFile: "v0.18.0/6fjangd7kvh4-fixture-thread.md",
		listGolden: "v0.18.0/thread-list.json", showGolden: "v0.18.0/thread-show.json",
		graphGolden: "v0.18.0/thread-graph.json", frontierGolden: "v0.18.0/thread-frontier.json",
		pathGolden: "v0.18.0/thread-path.json", planGolden: "v0.18.0/thread-plan.json",
	},
	{
		name: "v0.19.0", wireVersion: "1.60",
		threadFile: "v0.19.0/6fjangd7kvh4-fixture-thread.md",
		listGolden: "v0.19.0/thread-list.json", showGolden: "v0.19.0/thread-show.json",
		graphGolden: "v0.19.0/thread-graph.json", frontierGolden: "v0.19.0/thread-frontier.json",
		pathGolden: "v0.19.0/thread-path.json", planGolden: "v0.19.0/thread-plan.json",
	},
}

func compatibilityRepo(t *testing.T, releaseFile string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "planning")
	if err := os.CopyFS(root, os.DirFS(filepath.Join(threadCompatibilityFixtureRoot, "planning"))); err != nil {
		t.Fatalf("copy compatibility planning fixture: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, domain.ThreadsDir), 0o755); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(threadCompatibilityFixtureRoot, releaseFile))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, domain.ThreadsDir, filepath.Base(releaseFile)), content, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// assertHistoricalJSONSubset pins every field a preview release emitted without
// turning later additive fields into breaking changes. Lists retain exact order
// and cardinality because both are meaningful parts of the graph projection.
func assertHistoricalJSONSubset(t *testing.T, release, historicalVersion, artifact string, historical, current []byte) {
	t.Helper()
	var want, got map[string]any
	if err := json.Unmarshal(historical, &want); err != nil {
		t.Fatalf("decode %s %s fixture: %v", release, artifact, err)
	}
	if err := json.Unmarshal(current, &got); err != nil {
		t.Fatalf("decode current %s output: %v\n%s", artifact, err, current)
	}
	if want["schema_version"] != historicalVersion || got["schema_version"] != wire.SchemaVersion {
		t.Fatalf("%s %s schema versions: historical=%v current=%v", release, artifact, want["schema_version"], got["schema_version"])
	}
	delete(want, "schema_version")
	delete(got, "schema_version")
	assertJSONSubset(t, release+" "+artifact, want, got)
}

func assertJSONSubset(t *testing.T, at string, want, got any) {
	t.Helper()
	switch expected := want.(type) {
	case map[string]any:
		actual, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("%s: current value has type %T, want object", at, got)
		}
		for key, value := range expected {
			current, exists := actual[key]
			if !exists {
				t.Fatalf("%s.%s: historical field is missing", at, key)
			}
			assertJSONSubset(t, at+"."+key, value, current)
		}
	case []any:
		actual, ok := got.([]any)
		if !ok || len(actual) != len(expected) {
			t.Fatalf("%s: current list = %#v, want %d historical item(s)", at, got, len(expected))
		}
		for index := range expected {
			assertJSONSubset(t, fmt.Sprintf("%s[%d]", at, index), expected[index], actual[index])
		}
	default:
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("%s: current value = %#v, historical value = %#v", at, got, want)
		}
	}
}

func TestThreadPreviewReleaseWireSemanticsRemainCompatible(t *testing.T) {
	for _, release := range threadPreviewReleases {
		t.Run(release.name, func(t *testing.T) {
			root := compatibilityRepo(t, release.threadFile)

			listed, stderr, err := runIn(t, root, "thread", "list", "--json")
			if err != nil || stderr != "" {
				t.Fatalf("current list: %v\nstdout=%s\nstderr=%s", err, listed, stderr)
			}
			historicalList, err := os.ReadFile(filepath.Join(threadCompatibilityFixtureRoot, release.listGolden))
			if err != nil {
				t.Fatal(err)
			}
			assertHistoricalJSONSubset(t, release.name, release.wireVersion, "thread list", historicalList, []byte(listed))

			show, stderr, err := runIn(t, root, "thread", "show", "fixture-thread", "--json")
			if err != nil || stderr != "" {
				t.Fatalf("current show: %v\nstdout=%s\nstderr=%s", err, show, stderr)
			}
			historicalShow, err := os.ReadFile(filepath.Join(threadCompatibilityFixtureRoot, release.showGolden))
			if err != nil {
				t.Fatal(err)
			}
			assertHistoricalJSONSubset(t, release.name, release.wireVersion, "thread show", historicalShow, []byte(show))

			graph, stderr, err := runIn(t, root, "thread", "graph", "fixture-thread", "--json")
			if err != nil || stderr != "" {
				t.Fatalf("current graph: %v\nstdout=%s\nstderr=%s", err, graph, stderr)
			}
			historicalGraph, err := os.ReadFile(filepath.Join(threadCompatibilityFixtureRoot, release.graphGolden))
			if err != nil {
				t.Fatal(err)
			}
			assertHistoricalJSONSubset(t, release.name, release.wireVersion, "thread graph", historicalGraph, []byte(graph))

			frontier, stderr, err := runIn(t, root, "thread", "frontier", "fixture-thread", "--json")
			if err != nil || stderr != "" {
				t.Fatalf("current frontier: %v\nstdout=%s\nstderr=%s", err, frontier, stderr)
			}
			historicalFrontier, err := os.ReadFile(filepath.Join(threadCompatibilityFixtureRoot, release.frontierGolden))
			if err != nil {
				t.Fatal(err)
			}
			assertHistoricalJSONSubset(t, release.name, release.wireVersion, "thread frontier", historicalFrontier, []byte(frontier))

			threadPath, stderr, err := runIn(t, root, "thread", "path", "fixture-thread", "--json")
			if err != nil || stderr != "" {
				t.Fatalf("current path: %v\nstdout=%s\nstderr=%s", err, threadPath, stderr)
			}
			historicalPath, err := os.ReadFile(filepath.Join(threadCompatibilityFixtureRoot, release.pathGolden))
			if err != nil {
				t.Fatal(err)
			}
			resolvedRoot := root
			if resolved, err := filepath.EvalSymlinks(root); err == nil {
				resolvedRoot = resolved
			}
			normalizedPath := strings.ReplaceAll(threadPath, `\\`, "/")
			normalizedPath = strings.ReplaceAll(normalizedPath, filepath.ToSlash(resolvedRoot), "<ROOT>")
			assertHistoricalJSONSubset(t, release.name, release.wireVersion, "thread path", historicalPath, []byte(normalizedPath))

			plan, stderr, err := runIn(t, root, "thread", "plan", "fixture-thread", "--json")
			if err != nil || stderr != "" {
				t.Fatalf("current plan: %v\nstdout=%s\nstderr=%s", err, plan, stderr)
			}
			historicalPlan, err := os.ReadFile(filepath.Join(threadCompatibilityFixtureRoot, release.planGolden))
			if err != nil {
				t.Fatal(err)
			}
			assertHistoricalJSONSubset(t, release.name, release.wireVersion, "thread plan", historicalPlan, []byte(plan))
		})
	}
}

func TestThreadPreviewReleaseDocumentsRemainSurgicallyMutable(t *testing.T) {
	for _, release := range threadPreviewReleases {
		t.Run(release.name, func(t *testing.T) {
			root := compatibilityRepo(t, release.threadFile)
			path := filepath.Join(root, domain.ThreadsDir, filepath.Base(release.threadFile))
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			thread, body, err := store.NewFS(root).GetThread("fixture-thread")
			if err != nil || thread.ID != "6fjangd7kvh4" || thread.Status != domain.ThreadStatusUnstarted {
				t.Fatalf("read historical Thread: %+v body=%q err=%v", thread, body, err)
			}

			withUnknown := strings.Replace(string(before),
				"tags: [fixture, graph]\ntasks: [6fjangd7kvh0, 6fjangd7kvh1]",
				"future_scalar: keep\nfuture_map:\n  nested: keep\nfuture_list: [one, two]\ntags: [fixture, graph]\ntasks: [6fjangd7kvh0, 6fjangd7kvh1] # retained task comment", 1)
			if withUnknown == string(before) {
				t.Fatal("fixture insertion anchor was not found")
			}
			if err := os.WriteFile(path, []byte(withUnknown), 0o644); err != nil {
				t.Fatal(err)
			}

			addedText, stderr, err := runIn(t, root, "thread", "add", "fixture-thread", "gamma-task", "--json")
			if err != nil || stderr != "" {
				t.Fatalf("add old Thread member: %v\nstdout=%s\nstderr=%s", err, addedText, stderr)
			}
			var added wire.ThreadUpdateEnvelope
			if err := json.Unmarshal([]byte(addedText), &added); err != nil {
				t.Fatal(err)
			}
			if added.SchemaVersion != wire.SchemaVersion || added.Operation != string(core.ThreadMutationAddMembers) ||
				added.ThreadID != "6fjangd7kvh4" || !added.Changed || !added.Committed || len(added.MemberOutcomes) != 1 ||
				added.MemberOutcomes[0].Outcome != "added" || len(added.Before.Members) != 2 || len(added.After.Members) != 3 ||
				added.Workspace.RepoID != "6fjangd7kvhz" || !strings.HasPrefix(added.Path, "threads/") {
				t.Fatalf("membership receipt = %+v", added)
			}

			startedText, stderr, err := runIn(t, root, "thread", "start", "fixture-thread", "--json")
			if err != nil || stderr != "" {
				t.Fatalf("start old Thread: %v\nstdout=%s\nstderr=%s", err, startedText, stderr)
			}
			var started wire.ThreadUpdateEnvelope
			if err := json.Unmarshal([]byte(startedText), &started); err != nil {
				t.Fatal(err)
			}
			if started.SchemaVersion != wire.SchemaVersion || started.Operation != string(core.ThreadMutationStart) ||
				started.Before.Thread.Status != string(domain.ThreadStatusUnstarted) ||
				started.After.Thread.Status != string(domain.ThreadStatusInProgress) || started.After.Thread.StartedAt == "" {
				t.Fatalf("lifecycle receipt = %+v", started)
			}

			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			text := string(after)
			for _, retained := range []string{
				"schema: 1", "id: 6fjangd7kvh4", "future_scalar: keep", "future_map:\n  nested: keep",
				"future_list: [one, two]", "] # retained task comment", body,
			} {
				if !strings.Contains(text, retained) {
					t.Fatalf("surgical update lost %q:\n%s", retained, text)
				}
			}
			ordered := []string{"future_scalar:", "future_map:", "future_list:", "tags:", "tasks:"}
			prior := -1
			for _, key := range ordered {
				index := strings.Index(text, key)
				if index <= prior {
					t.Fatalf("historical/additive key order changed around %q:\n%s", key, text)
				}
				prior = index
			}
			mutated, mutatedBody, err := store.NewFS(root).GetThread("6fjangd7kvh4")
			if err != nil || mutated.ID != thread.ID || mutatedBody != body ||
				!slices.Equal(mutated.Tasks, []string{"6fjangd7kvh0", "6fjangd7kvh1", "6fjangd7kvh2"}) {
				t.Fatalf("mutated Thread = %+v body=%q err=%v", mutated, mutatedBody, err)
			}

			advisoryRoot := compatibilityRepo(t, release.threadFile)
			advisoryPath := filepath.Join(advisoryRoot, domain.ThreadsDir, filepath.Base(release.threadFile))
			advisory, err := os.ReadFile(advisoryPath)
			if err != nil {
				t.Fatal(err)
			}
			advisory = []byte(strings.Replace(string(advisory), "schema: 1", "schema: 2", 1))
			if err := os.WriteFile(advisoryPath, advisory, 0o644); err != nil {
				t.Fatal(err)
			}
			if output, stderr, err := runIn(t, advisoryRoot, "thread", "show", "fixture-thread", "--json"); err != nil || stderr != "" {
				t.Fatalf("read advisory-schema Thread: %v\nstdout=%s\nstderr=%s", err, output, stderr)
			}
			if output, stderr, err := runIn(t, advisoryRoot, "thread", "add", "fixture-thread", "gamma-task", "--json"); err != nil || stderr != "" {
				t.Fatalf("mutate advisory-schema Thread: %v\nstdout=%s\nstderr=%s", err, output, stderr)
			}
			advisory, err = os.ReadFile(advisoryPath)
			if err != nil || !strings.Contains(string(advisory), "schema: 2") {
				t.Fatalf("advisory schema marker was not preserved: %v\n%s", err, advisory)
			}
		})
	}
}

func TestThreadPreviewManifestSchemasRemainCompatible(t *testing.T) {
	manifests := []string{"manifest-schema-omitted.yml", "manifest-schema-zero.yml", "manifest-schema-one.yml"}
	for _, release := range threadPreviewReleases {
		for _, manifest := range manifests {
			t.Run(release.name+"/"+manifest, func(t *testing.T) {
				root := compatibilityRepo(t, release.threadFile)
				manifestPath := filepath.Join(threadCompatibilityFixtureRoot, "artifacts", manifest)
				planPath := filepath.Join(root, strings.TrimSuffix(manifest, ".yml")+"-plan.yml")
				output, stderr, err := runIn(t, root, "thread", "compose", "--from", manifestPath, "--out", planPath, "--json")
				if err != nil || stderr != "" {
					t.Fatalf("compose retained manifest: %v\nstdout=%s\nstderr=%s", err, output, stderr)
				}
				var envelope wire.ThreadApplyComposeEnvelope
				if err := json.Unmarshal([]byte(output), &envelope); err != nil {
					t.Fatal(err)
				}
				if envelope.SchemaVersion != wire.SchemaVersion || !envelope.Written || envelope.DryRun ||
					envelope.PlanPath != planPath || envelope.Plan.Schema != core.ThreadApplyPlanSchema ||
					envelope.Plan.PlanningRepoID != "6fjangd7kvhz" || envelope.Plan.Thread.Slug != "retained-preview-delivery" ||
					!slices.Equal(envelope.Plan.Thread.Tasks, []string{"6fjangd7kvh0", "6fjangd7kvh1"}) ||
					len(envelope.Plan.Dependencies) != 1 || envelope.Plan.Dependencies[0].From != "6fjangd7kvh2" ||
					envelope.Plan.Dependencies[0].To != "6fjangd7kvh1" || envelope.Workspace.RepoID != "6fjangd7kvhz" {
					t.Fatalf("compose envelope = %+v", envelope)
				}
			})
		}
	}

	t.Run("unsupported authoring schema writes nothing", func(t *testing.T) {
		root := compatibilityRepo(t, threadPreviewReleases[0].threadFile)
		content, err := os.ReadFile(filepath.Join(threadCompatibilityFixtureRoot, "artifacts", "manifest-schema-one.yml"))
		if err != nil {
			t.Fatal(err)
		}
		manifestPath := filepath.Join(root, "unsupported-manifest.yml")
		planPath := filepath.Join(root, "must-not-exist.yml")
		if err := os.WriteFile(manifestPath, []byte(strings.Replace(string(content), "schema: 1", "schema: 2", 1)), 0o644); err != nil {
			t.Fatal(err)
		}
		_, _, err = runIn(t, root, "thread", "compose", "--from", manifestPath, "--out", planPath)
		if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "unsupported Thread authoring manifest schema 2") {
			t.Fatalf("unsupported manifest error = %v", err)
		}
		if _, statErr := os.Stat(planPath); !os.IsNotExist(statErr) {
			t.Fatalf("unsupported manifest wrote plan: %v", statErr)
		}
	})
}

func TestThreadPreviewApplyPlanRemainsStrictAndRetryable(t *testing.T) {
	planFixture := filepath.Join(threadCompatibilityFixtureRoot, "artifacts", "plan-schema-one.yml")
	planContent, err := os.ReadFile(planFixture)
	if err != nil {
		t.Fatal(err)
	}

	for _, schema := range []string{"0", "2"} {
		t.Run("reject schema "+schema+" before mutation", func(t *testing.T) {
			root := compatibilityRepo(t, threadPreviewReleases[0].threadFile)
			betaPath := filepath.Join(root, domain.TasksDir, "6fjangd7kvh1-beta-task.md")
			before, err := os.ReadFile(betaPath)
			if err != nil {
				t.Fatal(err)
			}
			planPath := filepath.Join(root, "unsupported-plan.yml")
			if err := os.WriteFile(planPath, []byte(strings.Replace(string(planContent), "schema: 1", "schema: "+schema, 1)), 0o600); err != nil {
				t.Fatal(err)
			}
			_, _, err = runIn(t, root, "thread", "apply", planPath)
			if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "unsupported Thread apply-plan schema "+schema) {
				t.Fatalf("unsupported plan error = %v", err)
			}
			after, _ := os.ReadFile(betaPath)
			if !slices.Equal(before, after) {
				t.Fatal("unsupported plan mutated its dependency owner")
			}
			if _, statErr := os.Stat(filepath.Join(root, domain.ThreadsDir, "6fjangd7kvh5-retained-preview-delivery.md")); !os.IsNotExist(statErr) {
				t.Fatalf("unsupported plan created its Thread: %v", statErr)
			}
		})
	}

	t.Run("legacy repository identity names migration before mutation", func(t *testing.T) {
		root := compatibilityRepo(t, threadPreviewReleases[0].threadFile)
		configPath := filepath.Join(root, ".tskflwctl.toml")
		configBytes, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}
		legacy := strings.Replace(string(configBytes), "id = \"6fjangd7kvhz\"\n", "", 1)
		if err := os.WriteFile(configPath, []byte(legacy), 0o644); err != nil {
			t.Fatal(err)
		}
		betaPath := filepath.Join(root, domain.TasksDir, "6fjangd7kvh1-beta-task.md")
		before, _ := os.ReadFile(betaPath)
		_, _, err = runIn(t, root, "thread", "apply", planFixture)
		if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "config migrate") {
			t.Fatalf("legacy repository error = %v", err)
		}
		after, _ := os.ReadFile(betaPath)
		if !slices.Equal(before, after) {
			t.Fatal("identity refusal mutated its dependency owner")
		}
		if _, statErr := os.Stat(filepath.Join(root, domain.ThreadsDir, "6fjangd7kvh5-retained-preview-delivery.md")); !os.IsNotExist(statErr) {
			t.Fatalf("identity refusal created its Thread: %v", statErr)
		}
	})

	t.Run("interrupted durable prefix converges from the same plan", func(t *testing.T) {
		root := compatibilityRepo(t, threadPreviewReleases[1].threadFile)
		betaPath := filepath.Join(root, domain.TasksDir, "6fjangd7kvh1-beta-task.md")
		beta, err := os.ReadFile(betaPath)
		if err != nil {
			t.Fatal(err)
		}
		partial := strings.Replace(string(beta), "started_at: \"2026-01-03\"\n", "started_at: \"2026-01-03\"\ndepends_on: [6fjangd7kvh2]\n", 1)
		if partial == string(beta) {
			t.Fatal("partial apply insertion anchor was not found")
		}
		if err := os.WriteFile(betaPath, []byte(partial), 0o644); err != nil {
			t.Fatal(err)
		}

		output, stderr, err := runIn(t, root, "thread", "apply", planFixture, "--json")
		if err != nil || stderr != "" {
			t.Fatalf("retry retained plan: %v\nstdout=%s\nstderr=%s", err, output, stderr)
		}
		var applied wire.ThreadApplyEnvelope
		if err := json.Unmarshal([]byte(output), &applied); err != nil {
			t.Fatal(err)
		}
		if applied.SchemaVersion != wire.SchemaVersion || applied.ThreadID != "6fjangd7kvh5" ||
			applied.ThreadSlug != "retained-preview-delivery" || !applied.Changed || !applied.Complete || !applied.Committed ||
			len(applied.Operations) != 2 || applied.Operations[0].Kind != "dependency" ||
			applied.Operations[0].State != "skipped" || applied.Operations[1].Kind != "thread" ||
			applied.Operations[1].State != "applied" || applied.PlanPath != planFixture ||
			applied.Workspace.RepoID != "6fjangd7kvhz" {
			t.Fatalf("recovery apply = %+v", applied)
		}

		convergedText, stderr, err := runIn(t, root, "thread", "apply", planFixture, "--json")
		if err != nil || stderr != "" {
			t.Fatalf("converged retained plan: %v\nstdout=%s\nstderr=%s", err, convergedText, stderr)
		}
		var converged wire.ThreadApplyEnvelope
		if err := json.Unmarshal([]byte(convergedText), &converged); err != nil {
			t.Fatal(err)
		}
		if converged.Changed || !converged.Complete || converged.Committed ||
			len(converged.Operations) != 2 || converged.Operations[0].State != "skipped" || converged.Operations[1].State != "skipped" {
			t.Fatalf("converged apply = %+v", converged)
		}
	})
}
