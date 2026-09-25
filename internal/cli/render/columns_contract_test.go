package render

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

func TestColumnRegistriesRejectContradictoryDeclarations(t *testing.T) {
	for name, check := range map[string]func() error{
		"task":     func() error { return validateColumnRegistry(TaskColumns()) },
		"epic":     func() error { return validateColumnRegistry(EpicColumns()) },
		"audit":    func() error { return validateColumnRegistry(AuditColumns()) },
		"research": func() error { return validateColumnRegistry(ResearchColumns()) },
		"finding":  func() error { return validateColumnRegistry(FindingColumns()) },
		"thread":   func() error { return validateColumnRegistry(ThreadColumns()) },
	} {
		t.Run(name, func(t *testing.T) {
			if err := check(); err != nil {
				t.Fatalf("official registry is invalid: %v", err)
			}
		})
	}

	extract := func(s string) string { return s }
	invalid := []struct {
		name string
		cols []Column[string]
		want string
	}{
		{
			name: "duplicate display selector",
			cols: []Column[string]{column("same", "first", extract), column("same", "second", extract)},
			want: `selector "same"`,
		},
		{
			name: "canonical selector shadows another display name",
			cols: []Column[string]{
				contractColumn("legacy", "canonical", "first", extract, extract),
				column("canonical", "second", extract),
			},
			want: `selector "canonical"`,
		},
		{
			name: "missing display extractor",
			cols: []Column[string]{{Name: "broken"}},
			want: "no display extractor",
		},
		{
			name: "missing canonical selector",
			cols: []Column[string]{{
				Name: "legacy", Extract: extract,
				projection: &columnProjection[string]{extract: extract},
			}},
			want: "empty canonical selector",
		},
		{
			name: "missing raw projector",
			cols: []Column[string]{{
				Name: "legacy", Extract: extract,
				projection: &columnProjection[string]{name: "canonical"},
			}},
			want: "no raw projector",
		},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateColumnRegistry(tc.cols); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("validateColumnRegistry() error = %v, want text %q", err, tc.want)
			}
			for _, names := range [][]string{
				nil,
				{"anything"},
				{tc.cols[len(tc.cols)-1].selectorName()},
			} {
				if _, err := SelectColumns(tc.cols, names); err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("SelectColumns(%q) did not fail closed on the invalid registry: %v", names, err)
				}
			}
			assertPanicsWith(t, tc.want, func() { Specs(tc.cols) })
		})
	}
}

func TestContractColumnRequiresCanonicalSelectorAndRawProjector(t *testing.T) {
	extract := func(s string) string { return s }
	assertPanicsWith(t, "canonical selector and raw projector", func() {
		contractColumn("legacy", "", "missing name", extract, extract)
	})
	assertPanicsWith(t, "canonical selector and raw projector", func() {
		contractColumn("legacy", "canonical", "missing projector", extract, nil)
	})
}

func TestContractColumnSupportsSelfNamedCanonicalProjection(t *testing.T) {
	cols := columnRegistry(contractColumn(
		"value", "value", "self-named canonical field",
		func(string) string { return "display fallback" },
		func(string) string { return "" },
	))
	if got := cols[0].Extract(""); got != "display fallback" {
		t.Fatalf("default table should retain its display value, got %q", got)
	}
	var table bytes.Buffer
	WriteTablePlain(&table, cols, []string{""})
	if got := table.String(); got != "value\ndisplay fallback\n" {
		t.Fatalf("default table should retain its display rendering, got %q", got)
	}
	selected, err := SelectColumns(cols, []string{"value"})
	if err != nil {
		t.Fatal(err)
	}
	if got := selected[0].Extract(""); got != "" {
		t.Fatalf("explicit canonical selection should use the raw projector, got %q", got)
	}
	table.Reset()
	WriteTablePlain(&table, selected, []string{""})
	if got := table.String(); got != "value\n\n" {
		t.Fatalf("explicit canonical table should use the raw rendering, got %q", got)
	}
	var projected bytes.Buffer
	if err := ProjectedListJSON(&projected, "items", selected, []string{""}, nil); err != nil {
		t.Fatal(err)
	}
	if got := projected.String(); !strings.Contains(got, `"value":""`) {
		t.Fatalf("self-named canonical JSON projection should use the raw value: %s", got)
	}
}

func assertPanicsWith(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		got := recover()
		if got == nil || !strings.Contains(fmt.Sprint(got), want) {
			t.Fatalf("panic = %v, want text %q", got, want)
		}
	}()
	fn()
}

type registryFixture[T any] struct {
	name string
	item T
}

type projectionException struct {
	reason string
	want   func(map[string]json.RawMessage) (string, error)
}

func TestColumnRegistriesMatchFullWireValues(t *testing.T) {
	assertRegistryMatchesFullWire(t, "task", "tasks", TaskColumns(), TasksJSON,
		[]registryFixture[domain.Task]{
			{name: "present", item: domain.Task{
				ID: "6ga000000001", Slug: "task-present", Status: domain.StatusInProgress,
				Tier: 2, Priority: "high", Epic: "20-cli", Updated: "2026-09-14",
				Description: "present task", RevisitAt: "2026-10-01",
			}},
			{name: "absent optional strings", item: domain.Task{
				ID: "6ga000000002", Slug: "task-absent", Status: domain.StatusNextUp, Tier: 3,
			}},
			{name: "created but never edited", item: domain.Task{
				ID: "6ga000000007", Slug: "task-never-edited", Status: domain.StatusNextUp,
				Tier: 3, Created: "2026-09-01",
			}},
		}, nil)

	assertRegistryMatchesFullWire(t, "epic", "epics", EpicColumns(), EpicsJSON,
		[]registryFixture[core.EpicSummary]{
			{name: "present", item: core.EpicSummary{
				Epic: domain.Epic{ID: "20-cli", Status: "active", Priority: "high", Description: "present epic"},
				Done: 2, Total: 5, Deprecated: 1,
			}},
			{name: "zero rollup and absent optional strings", item: core.EpicSummary{
				Epic: domain.Epic{ID: "21-core"},
			}},
		}, nil)

	assertRegistryMatchesFullWire(t, "audit", "audits", AuditColumns(), AuditsJSON,
		[]registryFixture[domain.Audit]{
			{name: "present", item: domain.Audit{
				ID: "6ga000000003", Slug: "2026-09-14-present", Bucket: domain.AuditOpen,
				Area: "cli", Date: "2026-09-14", Findings: 4, OpenFindings: 2,
			}},
			{name: "zero counts and absent optional strings", item: domain.Audit{
				ID: "6ga000000004", Slug: "2026-09-14-zero", Bucket: domain.AuditClosed,
			}},
		}, nil)

	assertRegistryMatchesFullWire(t, "research", "research", ResearchColumns(), ResearchJSON,
		[]registryFixture[domain.Research]{
			{name: "present", item: domain.Research{
				ID: "6ga000000005", Slug: "research-present", Created: "2026-09-01",
				Description: "present research", Tags: []string{"cli", "contract"}, Updated: "2026-09-14",
			}},
			{name: "absent optional strings and empty tags", item: domain.Research{
				ID: "6ga000000006", Slug: "research-absent",
			}},
			{name: "created but never edited", item: domain.Research{
				ID: "6ga000000008", Slug: "research-never-edited", Created: "2026-09-01",
			}},
		}, map[string]projectionException{
			"tags": {
				reason: "projected list values are strings, so the documented tags view is comma-joined",
				want: func(row map[string]json.RawMessage) (string, error) {
					var tags []string
					if raw := row["tags"]; len(raw) > 0 {
						if err := json.Unmarshal(raw, &tags); err != nil {
							return "", err
						}
					}
					return strings.Join(tags, ","), nil
				},
			},
		})

	assertRegistryMatchesFullWire(t, "finding", "findings", FindingColumns(), FindingsJSON,
		[]registryFixture[core.AuditFinding]{
			{name: "present", item: core.AuditFinding{
				Audit: "2026-09-14-contract", Bucket: "open",
				Finding: domain.Finding{Code: "M1", Status: "open", Effort: "S", Urgency: "soon",
					Component: "cli", File: "internal/cli/render/columns.go:1", Title: "present finding"},
			}},
			{name: "absent optional strings", item: core.AuditFinding{
				Audit: "2026-09-14-contract", Bucket: "closed",
				Finding: domain.Finding{Code: "L1", Status: "fixed", Title: "minimal finding"},
			}},
		}, map[string]projectionException{
			"ref": {
				reason: "ref is the documented synthetic addressable audit:code identity",
				want: func(row map[string]json.RawMessage) (string, error) {
					audit, err := wireScalarString(row["audit"])
					if err != nil {
						return "", err
					}
					code, err := wireScalarString(row["code"])
					if err != nil {
						return "", err
					}
					return audit + ":" + code, nil
				},
			},
		})
}

func assertRegistryMatchesFullWire[T, P any](
	t *testing.T,
	registry, listKey string,
	cols []Column[T],
	fullJSON func(io.Writer, []T, []P) error,
	fixtures []registryFixture[T],
	exceptions map[string]projectionException,
) {
	t.Helper()
	for canonical, exception := range exceptions {
		if exception.reason == "" || exception.want == nil {
			t.Fatalf("column %q has an incomplete wire exception", canonical)
		}
		found := false
		for _, declared := range cols {
			if declared.selectorName() == canonical {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("wire exception %q does not name a registry column", canonical)
		}
	}
	wireCovered := make(map[string]bool, len(cols))
	for _, fixture := range fixtures {
		t.Run(registry+"/"+fixture.name, func(t *testing.T) {
			var full bytes.Buffer
			if err := fullJSON(&full, []T{fixture.item}, nil); err != nil {
				t.Fatal(err)
			}
			fullRow := decodeProjectedRow(t, full.Bytes(), listKey)

			for _, declared := range cols {
				canonical := declared.selectorName()
				raw, present := fullRow[canonical]
				if present && !bytes.Equal(raw, []byte("null")) {
					wireCovered[canonical] = true
				}
				want, err := wireScalarString(raw)
				if exception, ok := exceptions[canonical]; ok {
					want, err = exception.want(fullRow)
				}
				if err != nil {
					t.Fatalf("column %q needs an explicit string-view exception: %v", canonical, err)
				}

				selected, err := SelectColumns(cols, []string{canonical})
				if err != nil {
					t.Fatal(err)
				}
				if got := selected[0].Name; got != canonical {
					t.Errorf("canonical table/CSV header = %q, want %q", got, canonical)
				}
				if got := selected[0].Extract(fixture.item); got != want {
					t.Errorf("canonical table value for %q = %q, full wire value = %q", canonical, got, want)
				}

				var table bytes.Buffer
				WriteTablePlain(&table, selected, []T{fixture.item})
				if got := table.String(); got != canonical+"\n"+want+"\n" {
					t.Errorf("canonical table output for %q = %q, want header/value %q", canonical, got, want)
				}

				var csvOutput bytes.Buffer
				if err := WriteCSV(&csvOutput, selected, []T{fixture.item}); err != nil {
					t.Fatal(err)
				}
				var expectedCSV bytes.Buffer
				expectedWriter := csv.NewWriter(&expectedCSV)
				if err := expectedWriter.WriteAll([][]string{{canonical}, {csvInjectionSafe(want)}}); err != nil {
					t.Fatal(err)
				}
				if got := csvOutput.String(); got != expectedCSV.String() {
					t.Errorf("canonical CSV output for %q = %q, want header/value %q", canonical, got, want)
				}

				var projected bytes.Buffer
				if err := ProjectedListJSON(&projected, listKey, selected, []T{fixture.item}, nil); err != nil {
					t.Fatal(err)
				}
				projectedRow := decodeProjectedRow(t, projected.Bytes(), listKey)
				if len(projectedRow) != 1 || projectedRow[canonical] == nil {
					t.Fatalf("canonical JSON projection for %q has wrong keys: %#v", canonical, projectedRow)
				}
				var got string
				if err := json.Unmarshal(projectedRow[canonical], &got); err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Errorf("canonical JSON value for %q = %q, full wire value = %q", canonical, got, want)
				}
			}
		})
	}
	for _, declared := range cols {
		canonical := declared.selectorName()
		if !wireCovered[canonical] && exceptions[canonical].reason == "" {
			t.Errorf("%s column %q never reaches the full wire in any fixture", registry, canonical)
		}
	}
}

func decodeProjectedRow(t *testing.T, data []byte, listKey string) map[string]json.RawMessage {
	t.Helper()
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatal(err)
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(envelope[listKey], &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("%s envelope has %d rows, want 1", listKey, len(rows))
	}
	return rows[0]
}

func wireScalarString(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return "", nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return "", err
	}
	switch value := value.(type) {
	case string:
		return value, nil
	case json.Number:
		return value.String(), nil
	default:
		return "", fmt.Errorf("wire value %s is %T, not a scalar", raw, value)
	}
}

func TestTaskTierZeroProjectionIsLintInvalidCompatibility(t *testing.T) {
	task := domain.Task{
		ID: "6ga000000009", FilenameID: "6ga000000009", Slug: "missing-tier",
		Status: domain.StatusReadyToStart, Epic: "20-cli", Priority: "medium", Effort: "S",
		Created: "2026-09-14", Description: "lint-invalid tier fixture", Tags: []string{"contract"},
	}
	issues := domain.LintTask(task, func(string) bool { return true })
	if len(issues) != 1 || issues[0].Field != "tier" || issues[0].Message != "missing" {
		t.Fatalf("zero tier must be explicitly lint-invalid, got %#v", issues)
	}

	var full bytes.Buffer
	if err := TasksJSON(&full, []domain.Task{task}, nil); err != nil {
		t.Fatal(err)
	}
	if row := decodeProjectedRow(t, full.Bytes(), "tasks"); row["tier"] != nil {
		t.Fatalf("full JSON should omit the lint-invalid zero tier: %#v", row)
	}

	selected, err := SelectColumns(TaskColumns(), []string{"tier"})
	if err != nil {
		t.Fatal(err)
	}
	var projected bytes.Buffer
	if err := ProjectedListJSON(&projected, "tasks", selected, []domain.Task{task}, nil); err != nil {
		t.Fatal(err)
	}
	var tier string
	if err := json.Unmarshal(decodeProjectedRow(t, projected.Bytes(), "tasks")["tier"], &tier); err != nil {
		t.Fatal(err)
	}
	if tier != "0" {
		t.Fatalf("projected compatibility view should retain the established zero string, got %q", tier)
	}
}
