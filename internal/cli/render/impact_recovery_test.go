package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestDependencyHumanUsesCoreGuidanceWithoutInventingRecovery(t *testing.T) {
	for _, remedy := range []string{"", "owner-supplied recovery intent"} {
		receipt := core.DependencyMutationReceipt{Changed: true, Remedy: remedy,
			Impacts: []core.TaskGraphStateImpact{{TaskID: "affected-task",
				Before: core.TaskGraphState{Gate: core.GateClear}, After: core.TaskGraphState{Gate: core.GateBlocked},
			}},
		}
		var out bytes.Buffer
		if err := DependencyMutationHuman(&out, NewStyle(false), receipt); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "⚠ affected-task") ||
			strings.Contains(out.String(), "remedy:") != (remedy != "") ||
			(remedy != "" && !strings.Contains(out.String(), remedy)) || strings.Contains(out.String(), "remove the edge") {
			t.Fatalf("renderer invented or dropped guidance:\n%s", out.String())
		}
	}
}

func TestMovesHumanUsesProjectedImpactFlagsNotWireStatePolicy(t *testing.T) {
	for _, flag := range []bool{false, true} {
		// Deliberately disagree with the state fields. Rendering must trust the
		// copied core decisions rather than implement another policy over strings.
		afterGate := string(core.GateBlocked)
		if flag {
			afterGate = string(core.GateClear)
		}
		row := MoveResult{Slug: "moved", To: "next-up", Lifecycle: &wire.TaskLifecycleJSON{
			Impacts: []wire.TaskGraphStateImpactJSON{{TaskID: "affected-task", NewlyUnsafe: flag,
				Before: wire.TaskGraphStateJSON{Gate: string(core.GateClear)}, After: wire.TaskGraphStateJSON{Gate: afterGate},
			}},
			ThreadImpacts: []wire.ThreadProjectionImpactJSON{{ThreadID: "thread-id", Slug: "affected-thread",
				NewlyInconsistent: flag, After: wire.ThreadViewJSON{Inconsistent: !flag},
			}},
		}}
		var out, errOut bytes.Buffer
		MovesHuman(&out, &errOut, NewStyle(false), []MoveResult{row}, false)
		if errOut.Len() != 0 || strings.Contains(out.String(), "⚠ affected-task") != flag || strings.Contains(out.String(), "⚠ affected-thread") != flag {
			t.Fatalf("renderer re-derived policy instead of copying flags=%t:\n%s\n%s", flag, out.String(), errOut.String())
		}
	}
}
