package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/andy-esch/taskflow/internal/wire"
)

// updateGolden regenerates the committed snapshots: `go test ./internal/cli -update`.
var updateGolden = flag.Bool("update", false, "regenerate golden files under testdata/golden/")

const goldenRevisionPath = "testdata/golden/machine_contract_revision.txt"

var (
	machineGoldenSeenMu sync.Mutex
	machineGoldenSeen   = make(map[string]bool)
)

// assertGolden compares got against testdata/golden/<name>.golden, rewriting it
// under -update. A tiny dep-free helper (the repo's "no library for a 15-line job"
// ethos): the byte-stable machine contract is exactly what these snapshots lock,
// so any unintended drift in a --json envelope / csv / schema trips a diff.
func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name+".golden")
	if *updateGolden {
		_, machine, err := machineGoldenRevision(name, []byte(got))
		if err != nil {
			t.Fatal(err)
		}
		if machine {
			machineGoldenSeenMu.Lock()
			machineGoldenSeen[name] = true
			machineGoldenSeenMu.Unlock()
		}
		want, err := os.ReadFile(path)
		if err == nil && bytes.Equal(want, []byte(got)) {
			return
		}
		if err != nil && !os.IsNotExist(err) {
			t.Fatalf("read golden %s before update: %v", path, err)
		}
		baseline, err := readGoldenRevision()
		if err != nil {
			t.Fatal(err)
		}
		if err := validateMachineGoldenUpdate(name, want, []byte(got), baseline); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v\n(regenerate with: go test ./internal/cli -update)", path, err)
	}
	if got != string(want) {
		t.Errorf("output drift vs %s — regenerate with -update if intended.\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}

func validateMachineGoldenUpdate(name string, previous, data []byte, baseline string) error {
	revision, machine, err := machineGoldenRevision(name, data)
	if err != nil {
		return err
	}
	if machine && revision != wire.SchemaVersion {
		return fmt.Errorf("machine-contract golden %s declares revision %s, running contract is %s",
			name, revision, wire.SchemaVersion)
	}
	if machine && revision == baseline && !goldenBodyOnlyChange(name, previous, data) {
		return fmt.Errorf("refusing to rewrite machine-contract golden %s at unchanged revision %s; "+
			"advance SchemaVersion and append its classified changelog entry first", name, revision)
	}
	return nil
}

// goldenBodyOnlyChange recognizes the one byte-golden field whose semantics ADR-0008
// explicitly excludes from SchemaVersion. The allowlist identifies a committed snapshot,
// and comparing the decoded payloads with `body` removed proves that no adjacent template
// metadata or envelope field is smuggled through the exception.
func goldenBodyOnlyChange(name string, previous, candidate []byte) bool {
	if name != "template_show_security_json" || len(previous) == 0 {
		return false
	}
	var before, after map[string]json.RawMessage
	if json.Unmarshal(previous, &before) != nil || json.Unmarshal(candidate, &after) != nil {
		return false
	}
	var beforeBody, afterBody string
	if raw, ok := before["body"]; !ok || json.Unmarshal(raw, &beforeBody) != nil {
		return false
	}
	if raw, ok := after["body"]; !ok || json.Unmarshal(raw, &afterBody) != nil {
		return false
	}
	delete(before, "body")
	delete(after, "body")
	beforeContract, err := json.Marshal(before)
	if err != nil {
		return false
	}
	afterContract, err := json.Marshal(after)
	return err == nil && bytes.Equal(beforeContract, afterContract) && beforeBody != afterBody
}

func machineGoldenRevision(name string, data []byte) (string, bool, error) {
	if !strings.HasSuffix(name, "_json") && name != "schema_jsonschema" {
		return "", false, nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return "", true, fmt.Errorf("machine-contract golden %s is not a JSON object: %w", name, err)
	}
	encoded, ok := object["schema_version"]
	if !ok {
		encoded, ok = object["x-taskflow-schema-version"]
	}
	if !ok {
		return "", true, fmt.Errorf("machine-contract golden %s has no revision identity", name)
	}
	var revision string
	if err := json.Unmarshal(encoded, &revision); err != nil || revision == "" {
		return "", true, fmt.Errorf("machine-contract golden %s has an invalid revision identity", name)
	}
	return revision, true, nil
}

func readGoldenRevision() (string, error) {
	b, err := os.ReadFile(goldenRevisionPath)
	if err != nil {
		return "", fmt.Errorf("read machine-contract golden revision: %w", err)
	}
	revision := strings.TrimSpace(string(b))
	if revision == "" {
		return "", fmt.Errorf("machine-contract golden revision is empty")
	}
	return revision, nil
}

func finalizeGoldenRevision() error {
	current, err := readGoldenRevision()
	if err != nil {
		return err
	}
	if current == wire.SchemaVersion {
		return nil
	}
	entries, err := os.ReadDir(filepath.Dir(goldenRevisionPath))
	if err != nil {
		return err
	}
	machineGoldenSeenMu.Lock()
	defer machineGoldenSeenMu.Unlock()
	missing := make([]string, 0)
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), ".golden")
		if entry.IsDir() || (!strings.HasSuffix(name, "_json") && name != "schema_jsonschema") {
			continue
		}
		if !machineGoldenSeen[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("refusing to advance machine-contract golden revision after a partial update; "+
			"the full suite did not exercise: %s", strings.Join(missing, ", "))
	}

	dir := filepath.Dir(goldenRevisionPath)
	tmp, err := os.CreateTemp(dir, ".machine-contract-revision-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := fmt.Fprintln(tmp, wire.SchemaVersion); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, goldenRevisionPath)
}

func TestMachineGoldenUpdateRequiresRevisionAdvance(t *testing.T) {
	tests := []struct {
		name     string
		golden   string
		wantRev  string
		machine  bool
		wantFail bool
	}{
		{"task_list_json", `{"schema_version":"1.68"}`, "1.68", true, false},
		{"schema_jsonschema", `{"x-taskflow-schema-version":"1.68"}`, "1.68", true, false},
		{"task_list_csv", "slug,status\n", "", false, false},
		{"task_list_json", `{}`, "", true, true},
	}
	for _, tc := range tests {
		t.Run(tc.name+tc.wantRev, func(t *testing.T) {
			got, machine, err := machineGoldenRevision(tc.name, []byte(tc.golden))
			if (err != nil) != tc.wantFail {
				t.Fatalf("machineGoldenRevision error = %v, wantFail %v", err, tc.wantFail)
			}
			if !tc.wantFail && (got != tc.wantRev || machine != tc.machine) {
				t.Fatalf("machineGoldenRevision = %q, %v; want %q, %v", got, machine, tc.wantRev, tc.machine)
			}
		})
	}
}

func TestValidateMachineGoldenUpdate(t *testing.T) {
	changed := []byte(`{"schema_version":"1.73","new_contract_field":true}`)
	if err := validateMachineGoldenUpdate("schema_json", nil, changed, "1.73"); err == nil ||
		!strings.Contains(err.Error(), "unchanged revision 1.73") {
		t.Fatalf("same-revision rewrite should be refused, got %v", err)
	}
	if err := validateMachineGoldenUpdate("schema_json", nil, changed, "1.71"); err != nil {
		t.Fatalf("advanced-revision rewrite should be allowed: %v", err)
	}
	wrongRevision := []byte(`{"schema_version":"1.71","new_contract_field":true}`)
	if err := validateMachineGoldenUpdate("schema_json", nil, wrongRevision, "1.70"); err == nil ||
		!strings.Contains(err.Error(), "running contract is 1.73") {
		t.Fatalf("output from another revision should be refused, got %v", err)
	}
	if err := validateMachineGoldenUpdate("task_list_csv", nil, []byte("new header\n"), "1.73"); err != nil {
		t.Fatalf("non-JSON golden should not require a wire revision: %v", err)
	}

	before := []byte(`{"schema_version":"1.73","template":{"kind":"audit","name":"security","description":"same"},"body":"old"}`)
	bodyChanged := []byte(`{"schema_version":"1.73","template":{"kind":"audit","name":"security","description":"same"},"body":"new"}`)
	if err := validateMachineGoldenUpdate("template_show_security_json", before, bodyChanged, "1.73"); err != nil {
		t.Fatalf("ADR-0008 excludes a body-only Markdown template change from the JSON revision: %v", err)
	}
	metadataChanged := []byte(`{"schema_version":"1.73","template":{"kind":"audit","name":"security","description":"changed"},"body":"new"}`)
	metadataRemoved := []byte(`{"schema_version":"1.73","template":{"name":"security","description":"same"},"body":"new"}`)
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{name: "template_show_security_json", data: metadataChanged},
		{name: "template_show_security_json", data: metadataRemoved},
		{name: "template_show_but_actually_task_list_json", data: bodyChanged},
	} {
		if err := validateMachineGoldenUpdate(tc.name, before, tc.data, "1.73"); err == nil ||
			!strings.Contains(err.Error(), "unchanged revision 1.73") {
			t.Errorf("non-body contract change %s should be refused, got %v", tc.name, err)
		}
	}
}

func TestMachineGoldenRevisionMatchesSchemaVersion(t *testing.T) {
	if *updateGolden {
		return // TestMain advances the marker only after the complete update suite succeeds.
	}
	got, err := readGoldenRevision()
	if err != nil {
		t.Fatal(err)
	}
	if got != wire.SchemaVersion {
		t.Fatalf("machine-contract goldens describe revision %s, running contract is %s; regenerate with go test ./internal/cli -update", got, wire.SchemaVersion)
	}
}
