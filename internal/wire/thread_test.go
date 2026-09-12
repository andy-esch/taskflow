package wire

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

func TestToThreadJSONPreservesTagOrderAndCanonicalizesMembership(t *testing.T) {
	payload := ToThreadJSON(domain.Thread{
		Tags: []string{"z", "a", "m"}, Tasks: []string{"6g0000000002", "6g0000000001"},
	})
	if !slices.Equal(payload.Tags, []string{"z", "a", "m"}) {
		t.Fatalf("tags = %v", payload.Tags)
	}
	if !slices.Equal(payload.Tasks, []string{"6g0000000001", "6g0000000002"}) {
		t.Fatalf("tasks = %v", payload.Tasks)
	}
}

func TestToThreadsEnvelopeRetainsPathlessIdentityWithoutParsingLocation(t *testing.T) {
	const sourceVersion = "opaque-thread-source-revision"
	payload := ToThreadsEnvelope(core.ThreadListView{}, []core.ThreadReadProblem{
		{ThreadID: "6g0000000002", ThreadSlug: "pathless", Message: "remote decode failed", SourceVersion: sourceVersion},
		{ThreadID: "6g0000000001", ThreadSlug: "explicit", Location: "opaque://6g9999999999-wrong", Message: "bad record"},
	})
	if payload.Unreadable == nil || len(payload.Unreadable) != 2 {
		t.Fatalf("unreadable = %+v", payload.Unreadable)
	}
	if payload.Unreadable[0].ThreadID != "6g0000000002" || payload.Unreadable[0].Location != "" {
		t.Fatalf("pathless problem = %+v", payload.Unreadable[0])
	}
	if payload.Unreadable[1].ThreadID != "6g0000000001" || payload.Unreadable[1].Location != "opaque://6g9999999999-wrong" {
		t.Fatalf("located problem = %+v", payload.Unreadable[1])
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), sourceVersion) || strings.Contains(string(encoded), "source_version") {
		t.Fatalf("opaque Thread source revision leaked through wire JSON: %s", encoded)
	}
}

func TestToThreadGraphProjectionJSONAddsTitleWithoutReplacingLabel(t *testing.T) {
	payload := ToThreadGraphProjectionJSON(core.ThreadGraphProjection{Nodes: []core.ThreadGraphNode{{
		TaskID: "6g0000000001", Label: "canonical-slug", Title: "Canonical TUI Title",
	}}})
	if len(payload.Nodes) != 1 || payload.Nodes[0].Label != "canonical-slug" ||
		payload.Nodes[0].Title != "Canonical TUI Title" {
		t.Fatalf("node = %+v", payload.Nodes)
	}
}

func TestToThreadGraphProjectionJSONKeepsOptionalNeighborhoodScopeExact(t *testing.T) {
	full := ToThreadGraphProjectionJSON(core.ThreadGraphProjection{})
	if full.Scope != nil {
		t.Fatalf("full graph unexpectedly has scope: %+v", full.Scope)
	}

	bounded := ToThreadGraphProjectionJSON(core.ThreadGraphProjection{Scope: &core.ThreadGraphScope{
		Kind: core.ThreadGraphScopeNeighborhood, FocalTaskID: "6g0000000001", Depth: 2,
		TotalNodes: 8, ShownNodes: 4, HiddenNodes: 4, TotalEdges: 10, ShownEdges: 3, HiddenEdges: 7,
		BoundaryEdges: []core.ThreadGraphEdge{{From: "6g0000000004", To: "6g0000000005"}},
	}})
	if bounded.Scope == nil || bounded.Scope.Kind != "neighborhood" || bounded.Scope.FocalTaskID != "6g0000000001" ||
		bounded.Scope.Depth != 2 || bounded.Scope.TotalNodes != 8 || bounded.Scope.ShownNodes != 4 ||
		bounded.Scope.HiddenNodes != 4 || bounded.Scope.TotalEdges != 10 || bounded.Scope.ShownEdges != 3 ||
		bounded.Scope.HiddenEdges != 7 || !slices.Equal(bounded.Scope.BoundaryEdges, []ThreadGraphEdgeJSON{{From: "6g0000000004", To: "6g0000000005"}}) {
		t.Fatalf("scope=%+v", bounded.Scope)
	}
	encoded, err := json.Marshal(full)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"scope"`) {
		t.Fatalf("full graph encoded optional scope: %s", encoded)
	}
}
