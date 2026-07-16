package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// The `info` reads must print the FULL path even on a narrow terminal — the path is
// meant to be copy-pasted, so truncating it (as the browse views do for long values)
// would defeat the command. Guards against a regression to fit=true in fieldPrinter.
func TestTaskInfoHuman_PathNotTruncated(t *testing.T) {
	longPath := "/Users/someone/very/long/path/to/the/planning/tasks/6fxxxxxxxxxx-a-fairly-long-slug.md"
	var b bytes.Buffer
	st := NewStyle(false).WithWidth(40) // a narrow TTY
	TaskInfoHuman(&b, st, domain.Task{Slug: "s", Status: domain.StatusReadyToStart}, domain.ACCount{Checked: 1, Total: 2}, longPath)
	if !strings.Contains(b.String(), longPath) {
		t.Errorf("task info must print the full path on a narrow terminal, got:\n%s", b.String())
	}
}

func TestAuditInfoHuman_PathNotTruncated(t *testing.T) {
	longPath := "/Users/someone/very/long/path/to/the/planning/audits/6fxxxxxxxxxx-2026-01-02-some-area.md"
	var b bytes.Buffer
	st := NewStyle(false).WithWidth(40)
	AuditInfoHuman(&b, st, domain.Audit{Slug: "s", Bucket: domain.AuditOpen, Findings: 2, OpenFindings: 1}, longPath)
	if !strings.Contains(b.String(), longPath) {
		t.Errorf("audit info must print the full path on a narrow terminal, got:\n%s", b.String())
	}
}
