package wire

import (
	"slices"
	"testing"

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
