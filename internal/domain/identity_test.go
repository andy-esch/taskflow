package domain

import (
	"reflect"
	"testing"
)

func TestSemanticEntitiesDoNotCarrySourceEvidence(t *testing.T) {
	for _, entity := range []any{Task{}, Thread{}, Epic{}, Audit{}, Research{}} {
		typ := reflect.TypeOf(entity)
		t.Run(typ.Name(), func(t *testing.T) {
			for _, field := range []string{"Path", "FilenameID", "SourceVersion"} {
				if _, exists := typ.FieldByName(field); exists {
					t.Errorf("semantic %s contains source-only field %s", typ.Name(), field)
				}
			}
			if _, exists := typ.MethodByName("CanonicalID"); exists {
				t.Errorf("semantic %s supplies source-resolution identity", typ.Name())
			}
		})
	}
}
