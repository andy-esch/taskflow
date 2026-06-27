package store

import (
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

// TestSetMapNode_PreservesButDoesNotClobberComments guards audit L4: a surgical
// re-set inherits the old value node's comment ONLY when the replacement carries
// none — an intentional comment on the new node must survive.
func TestSetMapNode_PreservesButDoesNotClobberComments(t *testing.T) {
	scalar := func(val, head string) *yaml.Node {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: val, HeadComment: head}
	}
	mapping := func(head string) *yaml.Node {
		return &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "k"}, scalar("old", head),
		}}
	}
	// Inherit: the replacement has no comment → takes the old one.
	m := mapping("# keep me")
	setMapNode(m, "k", scalar("new", ""))
	if got := m.Content[1].HeadComment; got != "# keep me" {
		t.Errorf("comment should be preserved, got %q", got)
	}
	// Don't clobber: the replacement carries its own comment → kept.
	m2 := mapping("# old")
	setMapNode(m2, "k", scalar("new", "# new"))
	if got := m2.Content[1].HeadComment; got != "# new" {
		t.Errorf("the replacement's own comment must not be clobbered, got %q", got)
	}
}
