package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// No path capabilities and no finding-like body text: presentation must consume
// loaded evidence, not infer a path or reclassify the body to invent its count.
type incompleteAuditReadStore struct {
	core.Store
	sourceSet core.SourceSetID
	audit     domain.Audit
}

func (s *incompleteAuditReadStore) SourceSetID() core.SourceSetID { return s.sourceSet }
func (s *incompleteAuditReadStore) ReadAudits() (core.AuditRead, error) {
	return core.AuditRead{Records: []core.LoadedRecord[domain.Audit]{{Value: s.audit, Source: core.RecordSource{ID: "source-audit", Location: "urn:opaque:audit"}}}}, nil
}
func (s *incompleteAuditReadStore) ReadAudit(string) (core.LoadedRecord[core.AuditWithBody], error) {
	return core.LoadedRecord[core.AuditWithBody]{Value: core.AuditWithBody{Audit: s.audit, Body: "# Plain body\n"}, Source: core.RecordSource{ID: "source-audit", Location: "urn:opaque:audit"}}, nil
}

func TestPortableAuditViewsQualifyIncompleteEvidence(t *testing.T) {
	for _, tc := range []struct {
		name                           string
		parsed, done, active, unparsed int
	}{
		{"empty", 0, 0, 0, 0}, {"unparsed-only", 0, 0, 0, 2},
		{"mixed", 1, 1, 0, 2}, {"settled", 1, 1, 0, 0},
		{"in-progress", 1, 0, 1, 0}, {"missing-or-invalid-status", 1, 0, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := &incompleteAuditReadStore{sourceSet: core.NewSourceSetID(), audit: domain.Audit{
				ID: "declared-other", Slug: "portable", Bucket: domain.AuditOpen,
				Findings: tc.parsed, DoneFindings: tc.done, ActiveFindings: tc.active, UnparsedFindings: tc.unparsed,
			}}
			svc := core.MustNewService(source)
			msg := loadAuditList(&entityTab{}, svc)().(listLoadedMsg)
			if len(msg.items) != 1 || msg.identityErr != nil {
				t.Fatalf("portable list load: %+v", msg)
			}
			item := msg.items[0].(auditItem)
			if item.ref().key != "source-audit" {
				t.Fatal("read projection lost source identity")
			}
			detail := loadAuditDetail(svc, "source-audit")().(detailMsg)
			content := detail.content.(auditDetail)
			if content.localPath != "" {
				t.Fatal("opaque source location became a local path")
			}
			for _, width := range []int{80, 120, 240} {
				row := renderDelegateRow(t, auditDelegate{st: &testStyles}, item, width)
				meta := ansi.Strip(content.meta(width, &testStyles))
				for _, view := range []string{row, meta} {
					if tc.unparsed > 0 {
						if !strings.Contains(view, "2 unparsed") || strings.Contains(view, "no findings") || strings.Contains(view, "ready to close") {
							t.Fatalf("incomplete projection misrepresented evidence: %s", view)
						}
					} else if strings.Contains(view, "unparsed") || (tc.parsed == 0 && !strings.Contains(view, "no findings")) {
						t.Fatalf("clean projection misrepresented evidence: %s", view)
					}
					if tc.parsed > 0 && !strings.Contains(view, fmt.Sprintf("%d/%d", tc.done, tc.parsed)) {
						t.Fatalf("parsed denominator changed: %s", view)
					}
					if tc.done < tc.parsed && strings.Contains(view, "ready to close") {
						t.Fatalf("unsettled parsed findings advertised readiness: %s", view)
					}
				}
				if tc.unparsed > 0 && (!strings.Contains(meta, "audit lint source-audit") || strings.Contains(meta, "audit lint portable") || strings.Contains(meta, "audit lint declared-other")) {
					t.Fatalf("detail omitted diagnostic route: %s", meta)
				}
			}
		})
	}
}
