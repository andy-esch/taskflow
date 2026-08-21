package tomledit

import (
	"strings"
	"testing"
)

func TestSetTableKeyPreservesUnrelatedText(t *testing.T) {
	input := "# header\nunknown = 1\n\n[pager]\n  enabled = true # keep note\ncommand = \"a#b\"\n\n[other]\nx = 1\n"
	value := "false"
	got, changed, err := SetTableKey(input, "pager", "enabled", &value)
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	for _, want := range []string{"# header", "unknown = 1", "  enabled = false # keep note", `command = "a#b"`, "[other]", "x = 1"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
}

func TestSetTableKeyAddsAndRemoves(t *testing.T) {
	input := "theme = \"top-level\"\n"
	value := `"neon"`
	got, changed, err := SetTableKey(input, "theme", "name", &value)
	if err != nil || !changed || !strings.Contains(got, "[theme]\nname = \"neon\"") {
		t.Fatalf("add: changed=%v err=%v\n%s", changed, err, got)
	}
	removed, changed, err := SetTableKey(got, "theme", "name", nil)
	if err != nil || !changed || strings.Contains(removed, "name =") {
		t.Fatalf("remove: changed=%v err=%v\n%s", changed, err, removed)
	}
	if again, changed, err := SetTableKey(removed, "theme", "name", nil); err != nil || changed || again != removed {
		t.Fatalf("second remove should be no-op: changed=%v err=%v", changed, err)
	}
}
