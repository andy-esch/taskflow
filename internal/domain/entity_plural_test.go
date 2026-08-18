package domain

import "testing"

// "research" is a mass noun, so the old kind+"s" produced "researchs" in the
// ambiguous-match error. Every registered kind must declare a plural, and it must not be
// naive suffixing.
func TestPluralKind(t *testing.T) {
	want := map[string]string{
		"task":     "tasks",
		"epic":     "epics",
		"audit":    "audits",
		"research": "research docs",
	}
	for kind, plural := range want {
		if got := PluralKind(kind); got != plural {
			t.Errorf("PluralKind(%q) = %q, want %q", kind, got, plural)
		}
	}
	// Every registered kind declares one, so a new noun can't silently fall back.
	for _, d := range Descriptors() {
		if d.Plural == "" {
			t.Errorf("kind %q has no Plural — declare it so user-facing prose stays correct", d.Kind)
		}
	}
	// An unregistered kind falls back to the regular form.
	if got := PluralKind("project"); got != "projects" {
		t.Errorf("PluralKind fallback = %q, want %q", got, "projects")
	}
}
