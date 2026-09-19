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

var (
	schemaChangelogEntry                   = regexp.MustCompile(`(?m)^// (\d+)\.(\d+):`)
	schemaChangelogClassification          = regexp.MustCompile(`(?m)^// (\d+)\.(\d+): ([A-Z][A-Z ]*) —`)
	schemaChangelogClassificationCandidate = regexp.MustCompile(`(?m)^// (\d+)\.(\d+): ([A-Z][A-Z ]*)(?:\s+.*)?$`)
)

type schemaRelease struct{ major, minor int }

func (r schemaRelease) atOrAfter(other schemaRelease) bool {
	return r.major > other.major || (r.major == other.major && r.minor >= other.minor)
}

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

func validateSchemaClassifications(
	releases []schemaRelease,
	classifications map[schemaRelease]string,
	since schemaRelease,
	current schemaRelease,
	currentCompatibility string,
) error {
	valid := map[string]bool{"ADDITIVE": true, "NOT ADDITIVE": true}
	for _, release := range releases {
		if !release.atOrAfter(since) {
			continue
		}
		classification, ok := classifications[release]
		if !ok {
			return fmt.Errorf("SchemaVersion changelog %d.%d has no compatibility classification", release.major, release.minor)
		}
		if !valid[classification] {
			return fmt.Errorf("SchemaVersion changelog %d.%d has unknown compatibility classification %q", release.major, release.minor, classification)
		}
	}

	wantCurrent := strings.ToUpper(strings.ReplaceAll(currentCompatibility, "-", " "))
	if got := classifications[current]; got != wantCurrent {
		return fmt.Errorf("SchemaRevisionCompatibility = %q, current changelog classification = %q", currentCompatibility, got)
	}
	return nil
}

func parseSchemaClassifications(changelog []byte) (map[schemaRelease]string, error) {
	classifications := make(map[schemaRelease]string)
	for _, match := range schemaChangelogClassification.FindAllSubmatch(changelog, -1) {
		major, err := strconv.Atoi(string(match[1]))
		if err != nil {
			return nil, err
		}
		minor, err := strconv.Atoi(string(match[2]))
		if err != nil {
			return nil, err
		}
		classifications[schemaRelease{major, minor}] = strings.TrimSpace(string(match[3]))
	}
	for _, match := range schemaChangelogClassificationCandidate.FindAllSubmatch(changelog, -1) {
		major, err := strconv.Atoi(string(match[1]))
		if err != nil {
			return nil, err
		}
		minor, err := strconv.Atoi(string(match[2]))
		if err != nil {
			return nil, err
		}
		release := schemaRelease{major, minor}
		if _, ok := classifications[release]; !ok {
			return nil, fmt.Errorf(
				"SchemaVersion changelog %d.%d classification must use `ADDITIVE — description` or `NOT ADDITIVE — description` (with an em dash)",
				major, minor,
			)
		}
	}
	return classifications, nil
}

// TestSchemaVersionChangelogIsAscending makes ADR-0008's compatibility record
// executable: concurrent feature branches must append rather than insert, and
// every revision from the policy boundary declares its compatibility.
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
	changelog := source[start : start+end]
	matches := schemaChangelogEntry.FindAllSubmatch(changelog, -1)
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

	classifications, err := parseSchemaClassifications(changelog)
	if err != nil {
		t.Fatal(err)
	}
	since, err := parseSchemaRelease(SchemaRevisionClassificationSince)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateSchemaClassifications(releases, classifications, since, current, SchemaRevisionCompatibility); err != nil {
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

func TestValidateSchemaClassifications(t *testing.T) {
	r167 := schemaRelease{1, 67}
	r168 := schemaRelease{1, 68}
	r169 := schemaRelease{1, 69}
	releases := []schemaRelease{r167, r168, r169}

	if err := validateSchemaClassifications(releases, map[schemaRelease]string{
		r168: "ADDITIVE",
		r169: "NOT ADDITIVE",
	}, r168, r169, "not-additive"); err != nil {
		t.Fatalf("valid classifications: %v", err)
	}

	for _, tc := range []struct {
		name            string
		classifications map[schemaRelease]string
		current         string
	}{
		{"missing", map[schemaRelease]string{r168: "ADDITIVE"}, "not-additive"},
		{"unknown", map[schemaRelease]string{r168: "ADDITIVE", r169: "BREAKING"}, "not-additive"},
		{"current mismatch", map[schemaRelease]string{r168: "ADDITIVE", r169: "NOT ADDITIVE"}, "additive"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateSchemaClassifications(releases, tc.classifications, r168, r169, tc.current); err == nil {
				t.Fatal("invalid classifications accepted")
			}
		})
	}
}

func TestParseSchemaClassificationsExplainsDelimiter(t *testing.T) {
	_, err := parseSchemaClassifications([]byte("// 1.68: ADDITIVE - adds policy metadata\n"))
	if err == nil || !strings.Contains(err.Error(), "with an em dash") {
		t.Fatalf("ASCII delimiter should get an actionable diagnostic, got %v", err)
	}
}

func TestSchemaRevisionPolicyBoundary(t *testing.T) {
	if SchemaRevisionScheme != "monotonic-revision" {
		t.Fatalf("ADR-0008 revision scheme changed to %q", SchemaRevisionScheme)
	}
	if SchemaRevisionClassificationSince != "1.68" {
		t.Fatalf("ADR-0008 classification boundary changed to %q", SchemaRevisionClassificationSince)
	}
	if SchemaRevisionScope != "all-json-output" || JSONSchemaScope != "typed-envelopes" {
		t.Fatalf("ADR-0008 scopes changed: revision=%q generated-schema=%q", SchemaRevisionScope, JSONSchemaScope)
	}
	if SchemaRevisionCompatibilityDefault != "additive" ||
		SchemaRevisionReaderExpectation != "ignore-unknown-object-fields" ||
		JSONSchemaValidationMode != "exact-revision" {
		t.Fatalf("ADR-0008 reader policy changed: default=%q reader=%q validation=%q",
			SchemaRevisionCompatibilityDefault, SchemaRevisionReaderExpectation, JSONSchemaValidationMode)
	}
	current, err := parseSchemaRelease(SchemaVersion)
	if err != nil {
		t.Fatal(err)
	}
	if current.major != 1 {
		t.Fatalf("ADR-0008 reserves major revision movement for a superseding ADR; got %s", SchemaVersion)
	}
}
