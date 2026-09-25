package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestProjectedFindingsJSONPreservesOpaqueDiagnosticLocation(t *testing.T) {
	var out bytes.Buffer
	app := &App{Out: &out, ErrOut: &bytes.Buffer{}}
	findings := []core.AuditFinding{{
		Finding: domain.Finding{Code: "M1", Title: "portable read", Status: "open"},
		Audit:   "2026-09-24-portable", AuditID: "6g0000000001", Bucket: "open",
	}}
	problems := []core.LintLoadProblem{{
		EntityKind: core.LintEntityAudit, EntityID: "6g0000000002", EntitySlug: "broken-audit",
		Location: "db://audits/6g0000000002", LocationIsPath: false, Message: "remote decode failed",
	}}

	err := renderListWithProblems(app, modeJSON, []string{"audit", "code"}, findings, problems,
		"findings", render.FindingColumns(), wire.ToLintLoadProblemsJSON,
		render.FindingsJSON, render.FindingsHuman, render.LintProblemsHuman)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		SchemaVersion string                     `json:"schema_version"`
		Findings      []map[string]string        `json:"findings"`
		Unreadable    []wire.LintLoadProblemJSON `json:"unreadable"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("decode projected findings: %v\n%s", err, out.String())
	}
	if len(got.Findings) != 1 || got.Findings[0]["audit"] != "2026-09-24-portable" || got.Findings[0]["code"] != "M1" {
		t.Fatalf("projected findings = %+v", got.Findings)
	}
	if len(got.Unreadable) != 1 || got.Unreadable[0].EntityID != "6g0000000002" ||
		got.Unreadable[0].EntitySlug != "broken-audit" || got.Unreadable[0].Location != "db://audits/6g0000000002" ||
		got.Unreadable[0].Path != "" || got.Unreadable[0].Message != "remote decode failed" {
		t.Fatalf("projected unreadable = %+v\n%s", got.Unreadable, out.String())
	}
}
