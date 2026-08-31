package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

const (
	representativeThreadApplyTasks  = 1000
	representativeThreadApplyWrites = 300
)

func representativeThreadApplyFixture(tb testing.TB) (*FS, core.ThreadApplyPlan) {
	tb.Helper()
	root := tb.TempDir()
	tasksDir := filepath.Join(root, domain.TasksDir)
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		tb.Fatal(err)
	}
	ids := make([]string, representativeThreadApplyTasks)
	for index := range representativeThreadApplyTasks {
		seed := fmt.Sprintf("bulk-benchmark-%04d", index)
		ids[index] = testutil.TaskID(seed)
		content := fmt.Sprintf("---\nschema: 1\nid: %s\nstatus: next-up\nepic: 30\ndescription: benchmark task %d\neffort: 1h\ntier: 2\npriority: medium\nautonomy_level: 2\ntags: [benchmark]\ncreated: \"2026-08-30\"\n---\n# Benchmark %d\n", ids[index], index, index)
		if err := os.WriteFile(filepath.Join(tasksDir, ids[index]+"-"+seed+".md"), []byte(content), 0o644); err != nil {
			tb.Fatal(err)
		}
	}
	dependencies := make([]core.ThreadApplyDependency, 0, representativeThreadApplyWrites)
	members := make([]string, 0, representativeThreadApplyWrites)
	for index := 1; index <= representativeThreadApplyWrites; index++ {
		dependencies = append(dependencies, core.ThreadApplyDependency{From: ids[index-1], To: ids[index]})
		members = append(members, ids[index])
	}
	plan := storeThreadApplyPlan(testutil.TaskID("bulk-benchmark-thread"), members, dependencies...)
	repoID := "planning"
	fs := NewFS(root, WithPlanningIdentityReader(func() (string, string, error) {
		return root, repoID, nil
	}))
	return fs, plan
}

// BenchmarkThreadApplyRepresentativeGuardedDryRun measures the complete
// lock-held read, prefix validation, and materialization path at the release
// gate's representative 1,000-task / 300-write scale.
func BenchmarkThreadApplyRepresentativeGuardedDryRun(b *testing.B) {
	fs, plan := representativeThreadApplyFixture(b)
	b.ResetTimer()
	for range b.N {
		result, err := applyStoredThreadPlan(fs, plan, true)
		if err != nil || !result.Changed || result.Complete {
			b.Fatalf("result=%+v err=%v", result, err)
		}
	}
}

// BenchmarkThreadApplyRepresentativeGuardedApply includes the 300 atomic task
// replacements and final no-clobber Thread create. Fixture construction is not
// timed because production apply begins from an existing planning repository.
func BenchmarkThreadApplyRepresentativeGuardedApply(b *testing.B) {
	for range b.N {
		b.StopTimer()
		fs, plan := representativeThreadApplyFixture(b)
		b.StartTimer()
		result, err := applyStoredThreadPlan(fs, plan, false)
		if err != nil || !result.Complete || !result.Committed {
			b.Fatalf("result=%+v err=%v", result, err)
		}
	}
}
