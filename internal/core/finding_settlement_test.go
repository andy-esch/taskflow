package core

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// This adapter has no filesystem/path capability and hands the callback the only
// authoritative body/bucket. The service must not perform a separate read.
func TestEditFindingPortableBucketPolicy(t *testing.T) {
	for _, bucket := range domain.AllAuditBuckets() {
		for _, status := range []string{"open", "IN-PROGRESS (working)", "fixed 2026-10-09", "tracked by task", "deferred", "superseded", "wontfix"} {
			for _, dry := range []bool{true, false} {
				t.Run(fmt.Sprintf("%s/%s/dry=%t", bucket, status, dry), func(t *testing.T) {
					body := "#### H1. Current · **Status:** fixed\n"
					source := &findingCreationStore{body: body, bucket: bucket}
					svc := MustNewService(source)
					_, _, err := svc.SetFindingStatus("opaque-selector", "h1", status, dry)
					token := strings.ToLower(strings.Fields(status)[0])
					// Independent oracle: a shared production-mapping defect must not
					// change both the result and this test's expected outcome together.
					refuse := bucket != domain.AuditOpen && (token == "open" || token == "in-progress")
					if refuse != errors.Is(err, domain.ErrValidation) || (!refuse && err != nil) {
						t.Fatalf("portable status policy: %v", err)
					}
					if source.transformCalls != 1 || source.lastTransformDry != dry {
						t.Fatalf("wrong guarded operation: %+v", source)
					}
					if (refuse || dry) && source.body != body {
						t.Fatal("preview/refusal persisted body")
					}
				})
			}
		}
	}
}
