package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

// fixtureRepo is the committed, date-stable planning tree the golden snapshots run
// against (testdata/planning/). Edits there are intentional and regenerate via
// -update.
const fixtureRepo = "testdata/planning"

// redactFixtureRoot replaces the machine-specific absolute fixture path with a
// stable placeholder, so byte-golden cases for path-bearing output (task/audit info,
// task/epic/audit path — all emit filepath.Abs of the resolved store path) are
// portable across checkouts, CI, AND platforms instead of embedding one machine's
// repo root. It also folds Windows separators to '/': json.Marshal escapes native
// backslashes as `\\`, so the redactor normalizes those to '/' before matching, and
// resolves the root to its forward-slash form — the golden is one form everywhere.
func redactFixtureRoot(t *testing.T) func(string) string {
	t.Helper()
	root, err := filepath.Abs(fixtureRepo)
	if err != nil {
		t.Fatal(err)
	}
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved // discovery EvalSymlinks the root (macOS /var → /private/var)
	}
	slashRoot := filepath.ToSlash(root)
	return func(s string) string {
		s = strings.ReplaceAll(s, `\\`, "/") // JSON-escaped Windows separators → '/'
		return strings.ReplaceAll(s, slashRoot, "<ROOT>")
	}
}

// TestGolden_MachineContract locks the byte-stable machine surfaces against the
// committed fixture: a --json envelope, the csv/name shapes, or the schema contract
// changing shape trips a diff. The in-process runRoot tests prove the *logic*;
// these pin the exact bytes the agent contract promises. Run in-process (output is
// identical to the binary's); main.go wiring is covered by the subprocess smoke.
func TestGolden_MachineContract(t *testing.T) {
	redact := redactFixtureRoot(t)
	cases := []struct {
		name   string
		args   []string
		redact func(string) string // optional: normalize machine-specific output
	}{
		{"task_list_json", []string{"-C", fixtureRepo, "task", "list", "--all", "--json"}, nil},
		{"task_list_csv", []string{"-C", fixtureRepo, "task", "list", "--all", "-o", "csv"}, nil},
		{"task_list_name", []string{"-C", fixtureRepo, "task", "list", "--all", "-o", "name"}, nil},
		{"task_show_json", []string{"-C", fixtureRepo, "task", "show", "alpha-task", "--json"}, nil},
		{"task_acceptance_json", []string{"-C", fixtureRepo, "task", "ac", "alpha-task", "--json"}, nil},
		// task info / task path emit an absolute file path → redact the fixture root
		// so the committed golden is portable (pins schema_version + shape + tally).
		{"task_info_json", []string{"-C", fixtureRepo, "task", "info", "alpha-task", "--json"}, redact},
		{"task_path_json", []string{"-C", fixtureRepo, "task", "path", "alpha-task", "--json"}, redact},
		{"epic_path_json", []string{"-C", fixtureRepo, "epic", "path", "01-fixture-epic", "--json"}, redact},
		{"epic_list_json", []string{"-C", fixtureRepo, "epic", "list", "--json"}, nil},
		{"epic_show_json", []string{"-C", fixtureRepo, "epic", "show", "01-fixture-epic", "--json"}, nil},
		{"status_json", []string{"-C", fixtureRepo, "status", "--json"}, nil},
		{"board_json", []string{"-C", fixtureRepo, "board", "--json"}, nil},
		// audit info emits an absolute path → redact the fixture root (pins
		// schema_version + bucket + finding tally shape).
		{"audit_info_json", []string{"-C", fixtureRepo, "audit", "info", "2026-01-02-fixture-area", "--json"}, redact},
		{"audit_path_json", []string{"-C", fixtureRepo, "audit", "path", "2026-01-02-fixture-area", "--json"}, redact},
		{"audit_findings_json", []string{"-C", fixtureRepo, "audit", "findings", "--json"}, nil},
		{"audit_findings_open_json", []string{"-C", fixtureRepo, "audit", "findings", "--status", "open", "--json"}, nil},
		{"lint_json", []string{"-C", fixtureRepo, "lint", "--json"}, nil},
		// Self-description runs anywhere (no planning repo needed) and is fully
		// date-free — ideal to pin byte-for-byte, especially the JSON Schema.
		{"schema_json", []string{"schema", "--json"}, nil},
		{"schema_task_json", []string{"schema", "task", "--json"}, nil},
		{"schema_jsonschema", []string{"schema", "--json-schema"}, nil},
		// Templates are built-in + date-free, so their --json is byte-pinnable too.
		{"template_list_json", []string{"template", "list", "--json"}, nil},
		{"template_show_security_json", []string{"template", "show", "audit", "security", "--json"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := runRoot(t, tc.args...)
			if tc.redact != nil {
				out = tc.redact(out)
			}
			assertGolden(t, tc.name, out)
		})
	}
}
