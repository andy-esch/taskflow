package listfilter

import "testing"

func TestSubstring(t *testing.T) {
	targets := []string{
		"04-multiuser-10-frontend-backfill",
		"04-multiuser-13-internal-admin",
		"move-slack-webhook-url-secret-from-github-to-infisical", // fuzzy-matches, substring must not
	}
	if got := Substring("multiuser", targets); len(got) != 2 ||
		got[0].Index != 0 || got[1].Index != 1 {
		t.Errorf("substring 'multiuser' should match exactly the 2 multiuser items in order, got %+v", got)
	}
	if len(Substring("MULTI", targets)) != 2 {
		t.Error("filter should be case-insensitive")
	}
	if len(Substring("", targets)) != 3 {
		t.Error("empty term should match all (in order)")
	}
	if len(Substring("nope", targets)) != 0 {
		t.Error("no substring match should yield nothing")
	}
}

// TestSubstring_Unicode covers the rune-offset path and the lengthening ToLower
// case that previously panicked.
func TestSubstring_Unicode(t *testing.T) {
	got := Substring("multi", []string{"café-multiuser", "plain"})
	if len(got) != 1 || got[0].Index != 0 {
		t.Fatalf("should match only the café label, got %+v", got)
	}
	if len(got[0].MatchedIndexes) == 0 || got[0].MatchedIndexes[0] != 5 {
		t.Errorf("match should start at rune 5 (after 'café-'), got %v", got[0].MatchedIndexes)
	}
	// Ⱥ→ⱥ grows in ToLower; this must not panic (it did before the fix).
	_ = Substring("x", []string{"ȺȺȺx"})
}
