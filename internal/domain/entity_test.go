package domain

import "testing"

// TestEntityRegistry_CoversEverySchemaKind pins the M1 entity descriptor: every
// kind `schema` describes has a registry entry with a directory, authoring fields,
// and conventions, and the public lookups read the registry (so a new entity is a
// registry entry, not a per-layer switch edit).
func TestEntityRegistry_CoversEverySchemaKind(t *testing.T) {
	for _, kind := range SchemaKinds() {
		d, ok := descriptorFor(kind)
		if !ok {
			t.Fatalf("no descriptor for schema kind %q", kind)
		}
		if d.Dir == "" {
			t.Errorf("%s: descriptor has an empty Dir", kind)
		}
		if len(d.AuthoringFields) == 0 {
			t.Errorf("%s: descriptor has no authoring fields", kind)
		}
		if len(d.Conventions) == 0 {
			t.Errorf("%s: descriptor has no conventions", kind)
		}
		if d.BodyTemplate == "" {
			t.Errorf("%s: descriptor has no body template", kind)
		}
		if BodyTemplate(kind) != d.BodyTemplate {
			t.Errorf("%s: BodyTemplate disagrees with the descriptor", kind)
		}
		af, err := AuthoringFields(kind)
		if err != nil {
			t.Errorf("%s: AuthoringFields returned error: %v", kind, err)
		}
		if len(af) != len(d.AuthoringFields) {
			t.Errorf("%s: AuthoringFields (%d) disagrees with the descriptor (%d)", kind, len(af), len(d.AuthoringFields))
		}
		if len(Conventions(kind)) != len(d.Conventions) {
			t.Errorf("%s: Conventions disagrees with the descriptor", kind)
		}
	}
	// The directories are the canonical layout constants, tying kind ↔ dir.
	if d, _ := descriptorFor("task"); d.Dir != TasksDir {
		t.Errorf("task dir = %q, want %q", d.Dir, TasksDir)
	}
	if d, _ := descriptorFor("epic"); d.Dir != EpicsDir {
		t.Errorf("epic dir = %q, want %q", d.Dir, EpicsDir)
	}
	if d, _ := descriptorFor("audit"); d.Dir != AuditsDir {
		t.Errorf("audit dir = %q, want %q", d.Dir, AuditsDir)
	}
	// Descriptors() exposes the same kinds, in the same order, as a read-only copy.
	ds := Descriptors()
	if len(ds) != len(SchemaKinds()) {
		t.Errorf("Descriptors() has %d entries, SchemaKinds() %d", len(ds), len(SchemaKinds()))
	}
	for i, k := range SchemaKinds() {
		if ds[i].Kind != k {
			t.Errorf("Descriptors()[%d].Kind = %q, want %q", i, ds[i].Kind, k)
		}
	}
	// An unknown kind is a clean error / nil, not a panic.
	if _, err := AuthoringFields("bogus"); err == nil {
		t.Error("AuthoringFields(bogus) should error")
	}
	if Conventions("bogus") != nil {
		t.Error("Conventions(bogus) should be nil for an unknown kind")
	}
}
