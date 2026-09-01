package wire

import (
	"slices"
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
	payload := ToThreadsEnvelope(core.ThreadListView{}, []core.ThreadReadProblem{
		{ThreadID: "6g0000000002", ThreadSlug: "pathless", Message: "remote decode failed"},
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
}
