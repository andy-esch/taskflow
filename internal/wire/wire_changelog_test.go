package wire

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

var schemaChangelogEntry = regexp.MustCompile(`(?m)^// (\d+)\.(\d+):`)

type schemaRelease struct{ major, minor int }

func parseSchemaRelease(s string) (schemaRelease, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return schemaRelease{}, fmt.Errorf("schema version %q is not major.minor", s)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return schemaRelease{}, fmt.Errorf("schema version %q major: %w", s, err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return schemaRelease{}, fmt.Errorf("schema version %q minor: %w", s, err)
	}
	return schemaRelease{major, minor}, nil
}

func validateSchemaChangelog(releases []schemaRelease, current schemaRelease) error {
	if len(releases) == 0 {
		return fmt.Errorf("SchemaVersion changelog has no entries")
	}
	if releases[0] != (schemaRelease{1, 1}) {
		return fmt.Errorf("SchemaVersion changelog starts at %d.%d, want 1.1", releases[0].major, releases[0].minor)
	}
	for i := 1; i < len(releases); i++ {
		previous, next := releases[i-1], releases[i]
		if next.major == previous.major && next.minor == previous.minor+1 {
			continue
		}
		if next.major == previous.major+1 && next.minor == 0 {
			continue
		}
		return fmt.Errorf("SchemaVersion changelog is not contiguous: %d.%d follows %d.%d",
			next.major, next.minor, previous.major, previous.minor)
	}
	if last := releases[len(releases)-1]; last != current {
		return fmt.Errorf("SchemaVersion = %d.%d, last changelog entry = %d.%d",
			current.major, current.minor, last.major, last.minor)
	}
	return nil
}

// TestSchemaVersionChangelogIsAscending makes the compatibility record usable
// as a changelog: concurrent feature branches must append rather than inserting
// a newer version above older entries.
func TestSchemaVersionChangelogIsAscending(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	source, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), "wire.go"))
	if err != nil {
		t.Fatal(err)
	}
	start := bytes.Index(source, []byte("// SchemaVersion is"))
	if start < 0 {
		t.Fatal("find SchemaVersion changelog start")
	}
	end := bytes.Index(source[start:], []byte("const SchemaVersion"))
	if end < 0 {
		t.Fatal("find SchemaVersion declaration after changelog")
	}
	matches := schemaChangelogEntry.FindAllSubmatch(source[start:start+end], -1)
	releases := make([]schemaRelease, 0, len(matches))
	for _, match := range matches {
		major, err := strconv.Atoi(string(match[1]))
		if err != nil {
			t.Fatal(err)
		}
		minor, err := strconv.Atoi(string(match[2]))
		if err != nil {
			t.Fatal(err)
		}
		releases = append(releases, schemaRelease{major, minor})
	}
	current, err := parseSchemaRelease(SchemaVersion)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateSchemaChangelog(releases, current); err != nil {
		t.Fatal(err)
	}
}

func TestValidateSchemaChangelogAllowsMajorTransition(t *testing.T) {
	releases := []schemaRelease{{1, 1}, {1, 2}, {2, 0}, {2, 1}}
	if err := validateSchemaChangelog(releases, schemaRelease{2, 1}); err != nil {
		t.Fatalf("valid major transition: %v", err)
	}
}

func TestValidateSchemaChangelogRejectsMissingReleases(t *testing.T) {
	for _, tc := range []struct {
		name     string
		releases []schemaRelease
		current  schemaRelease
	}{
		{"missing minor", []schemaRelease{{1, 1}, {1, 3}}, schemaRelease{1, 3}},
		{"missing major", []schemaRelease{{1, 1}, {3, 0}}, schemaRelease{3, 0}},
		{"wrong current", []schemaRelease{{1, 1}, {1, 2}}, schemaRelease{1, 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateSchemaChangelog(tc.releases, tc.current); err == nil {
				t.Fatal("invalid changelog accepted")
			}
		})
	}
}
