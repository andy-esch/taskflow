package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestDependencyMutationPublishesCoreRecoveryIntent(t *testing.T) {
	for _, dryRun := range []bool{false, true} {
		t.Run(map[bool]string{false: "committed", true: "preview"}[dryRun], func(t *testing.T) {
			root := dependencyCLIRepo(t,
				dependencyCLITask{slug: "prerequisite", status: domain.StatusNextUp},
				dependencyCLITask{slug: "dependent", status: domain.StatusReadyToStart},
			)
			args := []string{"task", "depend", "add", "dependent", "--on", "prerequisite"}
			if dryRun {
				args = append(args, "--dry-run")
			}
			out, stderr, err := runIn(t, root, append(args, "--json")...)
			if err != nil || stderr != "" {
				t.Fatalf("dependency mutation = %v, stderr=%s", err, stderr)
			}
			var payload wire.DependencyMutationEnvelope
			if err := json.Unmarshal([]byte(out), &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Remedy == "" || len(payload.Impacts) != 1 || !payload.Impacts[0].NewlyUnsafe || payload.Impacts[0].After.Gate != "blocked" {
				t.Fatalf("JSON lost core-owned recovery intent or impact classification:\n%s", out)
			}
			// Use a fresh corpus for the human route so both observe the same mutation.
			humanRoot := dependencyCLIRepo(t,
				dependencyCLITask{slug: "prerequisite", status: domain.StatusNextUp},
				dependencyCLITask{slug: "dependent", status: domain.StatusReadyToStart},
			)
			human, stderr, err := runIn(t, humanRoot, append(args, "--color=never")...)
			if err != nil || stderr != "" || !strings.Contains(human, payload.Remedy) {
				t.Fatalf("human and machine recovery intent differ: %v, %s\n%s", err, stderr, human)
			}
		})
	}
}
