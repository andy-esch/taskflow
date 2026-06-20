package cli

import "testing"

// fixtureRepo is the committed, date-stable planning tree the golden snapshots run
// against (testdata/planning/). Edits there are intentional and regenerate via
// -update.
const fixtureRepo = "testdata/planning"

// TestGolden_MachineContract locks the byte-stable machine surfaces against the
// committed fixture: a --json envelope, the csv/name shapes, or the schema contract
// changing shape trips a diff. The in-process runRoot tests prove the *logic*;
// these pin the exact bytes the agent contract promises. Run in-process (output is
// identical to the binary's); main.go wiring is covered by the subprocess smoke.
func TestGolden_MachineContract(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"task_list_json", []string{"-C", fixtureRepo, "task", "list", "--all", "--json"}},
		{"task_list_csv", []string{"-C", fixtureRepo, "task", "list", "--all", "-o", "csv"}},
		{"task_list_name", []string{"-C", fixtureRepo, "task", "list", "--all", "-o", "name"}},
		{"task_show_json", []string{"-C", fixtureRepo, "task", "show", "alpha-task", "--json"}},
		{"epic_list_json", []string{"-C", fixtureRepo, "epic", "list", "--json"}},
		{"epic_show_json", []string{"-C", fixtureRepo, "epic", "show", "01-fixture-epic", "--json"}},
		{"status_json", []string{"-C", fixtureRepo, "status", "--json"}},
		{"lint_json", []string{"-C", fixtureRepo, "lint", "--json"}},
		// Self-description runs anywhere (no planning repo needed) and is fully
		// date-free — ideal to pin byte-for-byte, especially the JSON Schema.
		{"schema_json", []string{"schema", "--json"}},
		{"schema_task_json", []string{"schema", "task", "--json"}},
		{"schema_jsonschema", []string{"schema", "--json-schema"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertGolden(t, tc.name, runRoot(t, tc.args...))
		})
	}
}
