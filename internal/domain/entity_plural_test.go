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

// RECOGNIZED and SETTABLE are different sets, and conflating them is what made the
// unknown-field error advertise four keys the write path refuses. A caller offering
// choices — an error listing alternatives, a TUI edit menu — must use the settable set.
func TestResearchFields_RecognizedVsSettable(t *testing.T) {
	known := KnownResearchFieldNames()
	settable := SettableResearchFields()

	// Recognized: the registry's authoring fields plus the one stamp. Matches
	// KnownTaskFieldNames' rule — updated_at in, id/schema out.
	wantKnown := []string{"created", "description", "tags", "updated_at"}
	if len(known) != len(wantKnown) {
		t.Fatalf("recognized = %v, want %v", known, wantKnown)
	}
	for i, w := range wantKnown {
		if known[i] != w {
			t.Errorf("recognized[%d] = %q, want %q (sorted)", i, known[i], w)
		}
	}
	for _, absent := range []string{"id", "schema"} {
		if KnownResearchField(absent) {
			t.Errorf("%q is storage machinery and must not be a recognized authoring field", absent)
		}
	}

	// Settable: recognized MINUS protected. created and updated_at are tool-managed.
	wantSettable := []string{"description", "tags"}
	if len(settable) != len(wantSettable) {
		t.Fatalf("settable = %v, want %v", settable, wantSettable)
	}
	for i, w := range wantSettable {
		if settable[i] != w {
			t.Errorf("settable[%d] = %q, want %q", i, settable[i], w)
		}
	}
	// Every settable field must actually survive the protected gate — the invariant a
	// TUI edit menu depends on.
	for _, f := range settable {
		if reason, protected := ProtectedResearchField(f); protected {
			t.Errorf("settable field %q is protected (%s) — the two sets disagree", f, reason)
		}
	}
}
