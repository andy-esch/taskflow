package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
)

func TestGraphDiagnosticsHumanDeduplicatesRepeatedRepositoryProblems(t *testing.T) {
	problems := []core.GraphProblem{
		{Code: core.ProblemCycle, TaskID: "6g0000000001", Message: "dependency cycle: one -> two -> one"},
		{Code: core.ProblemCycle, TaskID: "6g0000000002", Message: "dependency cycle: one -> two -> one"},
		{Code: core.ProblemMissingDependency, TaskID: "6g0000000003", Message: "missing dependency"},
	}
	var out bytes.Buffer
	graphDiagnosticsHuman(&out, NewStyle(false), problems, nil)
	if got := strings.Count(out.String(), "dependency cycle:"); got != 1 {
		t.Fatalf("cycle rendered %d times:\n%s", got, out.String())
	}
	if !strings.Contains(out.String(), "missing dependency") {
		t.Fatalf("distinct repository problem was lost:\n%s", out.String())
	}
}
