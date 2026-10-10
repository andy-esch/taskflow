package core

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

type queryItem struct{ ID, Name, Created, Updated, Text string }

func queryItems(items []queryItem, q ListQuery, diagnostics any) (ListResult[queryItem], error) {
	if err := validateListQuery(q, true); err != nil {
		return ListResult[queryItem]{}, err
	}
	return queryList(items, q, "id", "portable-test", diagnostics, func(i queryItem) listRow {
		return listRow{i.ID, i.Name, i.Created, i.Updated, i.Text}
	})
}

func TestListQueryLegacyAndPageBoundaries(t *testing.T) {
	items := []queryItem{{ID: "c"}, {ID: "a"}, {ID: "b"}}
	before := slices.Clone(items)
	legacy, err := queryItems(items, ListQuery{}, nil)
	if err != nil || legacy.Page != nil || !reflect.DeepEqual(legacy.Items, items) {
		t.Fatalf("legacy=%+v, %v", legacy, err)
	}
	for _, count := range []int{0, 1, 2, 3} {
		for _, limit := range []int{1, 2, 3} {
			q := ListQuery{Limit: limit}
			var ids []string
			for pages := 0; ; pages++ {
				if pages > count {
					t.Fatal("pagination did not terminate")
				}
				result, err := queryItems(items[:count], q, nil)
				if err != nil || result.Page == nil || result.Page.Returned != len(result.Items) ||
					result.Page.Matched != count || len(result.Items) > limit || result.Items == nil {
					t.Fatalf("count=%d limit=%d result=%+v, %v", count, limit, result, err)
				}
				for _, item := range result.Items {
					ids = append(ids, item.ID)
				}
				if result.Page.NextCursor == "" {
					break
				}
				q.Cursor = result.Page.NextCursor
			}
			want := make([]string, 0, count)
			for _, item := range items[:count] {
				want = append(want, item.ID)
			}
			slices.Sort(want)
			if !slices.Equal(ids, want) {
				t.Fatalf("coverage=%v want=%v", ids, want)
			}
		}
	}
	if !reflect.DeepEqual(before, items) {
		t.Fatal("query reordered caller's list")
	}
}

func TestListQueryStableTiesFiltersAndContinuation(t *testing.T) {
	items := []queryItem{
		{ID: "c", Name: "same", Created: "2026-10-09", Text: "BIKE touring"},
		{ID: "a", Name: "same", Created: "2026-10-08", Updated: "2026-10-10", Text: "bike touring"},
		{ID: "b", Name: "same", Created: "2026-10-07", Text: "bike touring"},
	}
	q := ListQuery{Limit: 1, Sort: "-name", Search: "  BiKe  ", UpdatedSince: "2026-10-09"}
	first, err := queryItems(items, q, nil)
	if err != nil || len(first.Items) != 1 || first.Items[0].ID != "a" || first.Page.Matched != 2 || first.Page.Search != "bike" {
		t.Fatalf("first=%+v, %v", first, err)
	}
	q.Cursor = first.Page.NextCursor
	slices.Reverse(items)
	second, err := queryItems(items, q, nil)
	if err != nil || len(second.Items) != 1 || second.Items[0].ID != "c" || second.Page.NextCursor != "" {
		t.Fatalf("second=%+v, %v", second, err)
	}
	items = append(items, queryItem{ID: "nonmatching", Text: "different"})
	if _, err := queryItems(items, q, nil); err != nil {
		t.Fatalf("nonmatching addition invalidated cursor: %v", err)
	}
	for _, sort := range []string{"id", "-id", "name", "-name", "created", "-created", "updated", "-updated"} {
		if _, err := queryItems(items, ListQuery{Sort: sort}, nil); err != nil {
			t.Fatalf("sort %s: %v", sort, err)
		}
	}
}

func TestListQueryRefusesChangedViewsAndForeignQueries(t *testing.T) {
	items := []queryItem{{ID: "a", Text: "match"}, {ID: "b", Text: "match"}}
	first, err := queryItems(items, ListQuery{Limit: 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	q := ListQuery{Limit: 1, Cursor: first.Page.NextCursor}
	for _, mutate := range []func([]queryItem) []queryItem{
		func(i []queryItem) []queryItem { i[0].Text = "edit before offset"; return i },
		func(i []queryItem) []queryItem { i[1].Name = "renamed"; return i },
		func(i []queryItem) []queryItem { return i[:1] },
		func(i []queryItem) []queryItem { return append(i, queryItem{ID: "c"}) },
	} {
		if _, err := queryItems(mutate(slices.Clone(items)), q, nil); !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "restart") {
			t.Fatalf("changed view was continued: %v", err)
		}
	}
	if _, err := queryItems(items, q, "new diagnostic"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("changed diagnostics: %v", err)
	}
	q.Sort = "-id"
	if _, err := queryItems(items, q, nil); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("foreign query: %v", err)
	}
	q.Sort, q.Limit = "", 2
	if _, err := queryItems(items, q, nil); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("changed limit: %v", err)
	}
}

func TestListQueryRejectsInvalidInputsAndCursors(t *testing.T) {
	for _, q := range []ListQuery{
		{Limit: -1}, {Limit: MaxListLimit + 1}, {Cursor: "anything"}, {Sort: "unknown"}, {Sort: "-"},
		{UpdatedSince: "2026-02-30"}, {Search: strings.Repeat("x", MaxListSearchBytes+1)},
		{Limit: 1, Cursor: strings.Repeat("x", maxListCursorBytes+1)},
	} {
		if _, err := queryItems(nil, q, nil); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("query=%+v err=%v", q, err)
		}
	}
	for _, q := range []ListQuery{{UpdatedSince: "2026-10-10"}, {Sort: "created"}, {Sort: "-updated"}} {
		if err := validateListQuery(q, false); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("finding query=%+v err=%v", q, err)
		}
	}
	for _, value := range []string{"!", "{}", `{"v":99,"q":"bad","s":"bad","o":1}`, `{} {}`, `{"unknown":1}`} {
		if value != "!" {
			value = base64.RawURLEncoding.EncodeToString([]byte(value))
		}
		if _, err := queryItems(nil, ListQuery{Limit: 1, Cursor: value}, nil); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("cursor=%q err=%v", value, err)
		}
	}
	items := []queryItem{{ID: "a"}, {ID: "b"}}
	first, _ := queryItems(items, ListQuery{Limit: 1}, nil)
	cursor, err := decodeListCursor(first.Page.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	for _, offset := range []int{-1, 0, 2, 999999} {
		cursor.Offset = offset
		b, _ := json.Marshal(cursor)
		if _, err := queryItems(items, ListQuery{Limit: 1, Cursor: base64.RawURLEncoding.EncodeToString(b)}, nil); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("offset=%d, err=%v", offset, err)
		}
	}
}

func TestListQueryKeepsDuplicateOccurrencesDeterministic(t *testing.T) {
	items := []queryItem{{ID: "same", Name: "same", Text: "second"}, {ID: "same", Name: "same", Text: "first"}, {ID: "same", Name: "same", Text: "first"}}
	first, err := queryItems(items, ListQuery{Limit: 1, Sort: "name"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	slices.Reverse(items)
	again, err := queryItems(items, ListQuery{Limit: 1, Sort: "name"}, nil)
	if err != nil || !reflect.DeepEqual(first, again) || first.Page.Matched != 3 {
		t.Fatalf("permutation changed duplicate coverage/order: %+v %+v, %v", first, again, err)
	}
	q := ListQuery{Limit: 1, Sort: "name"}
	count := 0
	for {
		page, err := queryItems(items, q, nil)
		if err != nil {
			t.Fatal(err)
		}
		count += len(page.Items)
		if page.Page.NextCursor == "" {
			break
		}
		q.Cursor = page.Page.NextCursor
	}
	if count != 3 {
		t.Fatalf("pagination collapsed duplicate occurrences: %d", count)
	}
}

func TestListCursorRequiresCanonicalSHA256Digests(t *testing.T) {
	items := []queryItem{{ID: "a"}, {ID: "b"}}
	first, err := queryItems(items, ListQuery{Limit: 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	original, err := decodeListCursor(first.Page.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"query", "view"} {
		for _, value := range []string{"", strings.Repeat("0", 63), strings.Repeat("0", 65), strings.Repeat("g", 64), strings.Repeat("A", 64)} {
			cursor := original
			if field == "query" {
				cursor.Query = value
			} else {
				cursor.View = value
			}
			b, err := json.Marshal(cursor)
			if err != nil {
				t.Fatal(err)
			}
			encoded := base64.RawURLEncoding.EncodeToString(b)
			if _, err := decodeListCursor(encoded); !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("decoder accepted malformed %s digest %q: %v", field, value, err)
			}
			if _, err := queryItems(items, ListQuery{Limit: 1, Cursor: encoded}, nil); !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("query misclassified malformed %s digest %q: %v", field, value, err)
			}
		}
	}
}

type queryResearchReader struct {
	fakeStore
	records []LoadedRecord[domain.Research]
}

func (r *queryResearchReader) ReadResearch() (ResearchRead, error) {
	return ResearchRead{Records: r.records}, nil
}

func TestListQueryUsesPortableSourceIdentityNotDeclaredIdentity(t *testing.T) {
	reader := &queryResearchReader{records: []LoadedRecord[domain.Research]{
		{Value: domain.Research{ID: "a-declared", Slug: "same", Created: "2026-10-10"}, Source: RecordSource{ID: "z-source", Location: "db://opaque-z"}},
		{Value: domain.Research{ID: "z-declared", Slug: "same", Created: "2026-10-10"}, Source: RecordSource{ID: "a-source", Location: "db://opaque-a"}},
	}}
	svc := MustNewService(reader)
	first, problems, err := svc.QueryResearch("", ListQuery{Limit: 1})
	if err != nil || len(problems) != 0 || first.Items[0].Source.ID != "a-source" || first.Items[0].Value.ID != "z-declared" {
		t.Fatalf("first=%+v problems=%v err=%v", first, problems, err)
	}
	second, _, err := svc.QueryResearch("", ListQuery{Limit: 1, Cursor: first.Page.NextCursor})
	if err != nil || second.Items[0].Source.ID != "z-source" {
		t.Fatalf("second=%+v err=%v", second, err)
	}
}
