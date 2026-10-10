package core

import (
	"cmp"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

const (
	MaxListLimit       = 1000
	MaxListSearchBytes = 256
	maxListCursorBytes = 2048
	// Bump when changing continuation encoding or ordering/digest semantics.
	// Never substitute a live source-set witness for the complete view digest.
	listCursorVersion = 1
)

// ListQuery is an optional application-level view over an existing list read.
// Zero preserves legacy ordering and filtering. A positive Limit bounds returned
// rows, not adapter I/O; local readers still scan the corpus. Cursor is an opaque,
// versioned continuation, not a persisted source-set witness or write authority.
type ListQuery struct {
	Sort         string
	Limit        int
	Cursor       string
	Search       string
	UpdatedSince string
}

func (q ListQuery) Active() bool {
	return q.Sort != "" || q.Limit != 0 || q.Cursor != "" || q.Search != "" || q.UpdatedSince != ""
}

// ListPage describes the applied query and how to resume. NextCursor is absent
// on the final page. Any change to matching returned values or full diagnostics
// invalidates continuation; this is detection, not a multi-read snapshot lock.
type ListPage struct {
	Order        []string
	Limit        int
	Returned     int
	Matched      int
	Search       string
	UpdatedSince string
	NextCursor   string
}

type ListResult[T any] struct {
	Items []T
	Page  *ListPage
}

type listRow struct{ id, name, created, updated, text string }

// A cursor binds both the normalized query and the complete matching view. It
// deliberately contains only digests and an offset, never source paths/bodies.
type listCursor struct {
	Version int    `json:"v"`
	Query   string `json:"q"`
	View    string `json:"s"`
	Offset  int    `json:"o"`
}

func validateListQuery(q ListQuery, dated bool) error {
	if q.Limit < 0 || q.Limit > MaxListLimit {
		return fmt.Errorf("%w: limit must be between 1 and %d, or omitted", domain.ErrValidation, MaxListLimit)
	}
	if q.Cursor != "" && q.Limit == 0 {
		return fmt.Errorf("%w: cursor requires a positive limit", domain.ErrValidation)
	}
	if len(q.Search) > MaxListSearchBytes {
		return fmt.Errorf("%w: search exceeds %d UTF-8 bytes", domain.ErrValidation, MaxListSearchBytes)
	}
	field := strings.TrimPrefix(q.Sort, "-")
	if q.Sort == "-" || (field != "" && field != "id" && field != "name" && field != "created" && field != "updated") {
		return fmt.Errorf("%w: unsupported sort %q (id|name|created|updated, optionally prefixed with -)", domain.ErrValidation, q.Sort)
	}
	if !dated && (field == "created" || field == "updated" || q.UpdatedSince != "") {
		return fmt.Errorf("%w: findings have no creation/update date; use id or name sorting without updated-since", domain.ErrValidation)
	}
	if q.UpdatedSince != "" {
		if err := domain.ValidateDate(q.UpdatedSince); err != nil {
			return fmt.Errorf("updated-since: %w", err)
		}
	}
	if len(q.Cursor) > maxListCursorBytes {
		return fmt.Errorf("%w: cursor is too large", domain.ErrValidation)
	}
	return nil
}

func queryList[T any](items []T, q ListQuery, defaultSort string, scope, diagnostics any, row func(T) listRow) (ListResult[T], error) {
	if !q.Active() {
		return ListResult[T]{Items: items}, nil
	}
	q.Search = strings.ToLower(strings.TrimSpace(q.Search))
	if q.Sort == "" {
		q.Sort = defaultSort
	}
	// Continuation itself is excluded, while changing the page size is a query change.
	query, err := listDigest(struct {
		Scope         any
		Sort          string
		Limit         int
		Search, Since string
	}{scope, q.Sort, q.Limit, q.Search, q.UpdatedSince})
	if err != nil {
		return ListResult[T]{}, err
	}
	type entry struct {
		value  T
		row    listRow
		digest string
	}
	entries := make([]entry, 0, len(items))
	for _, item := range items {
		r := row(item)
		if q.Search != "" && !strings.Contains(strings.ToLower(r.text), q.Search) {
			continue
		}
		if r.updated == "" {
			r.updated = r.created
		}
		if q.UpdatedSince != "" && (domain.ValidateDate(r.updated) != nil || r.updated < q.UpdatedSince) {
			continue
		}
		digest, err := listDigest(item)
		if err != nil {
			return ListResult[T]{}, err
		}
		entries = append(entries, entry{item, r, digest})
	}
	key := func(r listRow) string {
		switch strings.TrimPrefix(q.Sort, "-") {
		case "name":
			return r.name
		case "created":
			return r.created
		case "updated":
			return r.updated
		default:
			return r.id
		}
	}
	slices.SortFunc(entries, func(a, b entry) int {
		order := cmp.Compare(key(a.row), key(b.row))
		if strings.HasPrefix(q.Sort, "-") {
			order = -order
		}
		if order != 0 {
			return order
		}
		if order = cmp.Compare(a.row.id, b.row.id); order != 0 {
			return order
		}
		return cmp.Compare(a.digest, b.digest)
	})
	digests := make([]string, len(entries))
	for i := range entries {
		digests[i] = entries[i].digest
	}
	view, err := listDigest(struct {
		Rows        []string
		Diagnostics any
	}{digests, diagnostics})
	if err != nil {
		return ListResult[T]{}, err
	}
	offset := 0
	if q.Cursor != "" {
		cursor, err := decodeListCursor(q.Cursor)
		if err != nil {
			return ListResult[T]{}, err
		}
		if cursor.Query != query {
			return ListResult[T]{}, fmt.Errorf("%w: cursor belongs to a different query; repeat its filters, sort, and limit", domain.ErrValidation)
		}
		if cursor.View != view {
			return ListResult[T]{}, fmt.Errorf("%w: matching list results or diagnostics changed; restart the query without a cursor", domain.ErrConflict)
		}
		if cursor.Offset <= 0 || cursor.Offset >= len(entries) {
			return ListResult[T]{}, fmt.Errorf("%w: cursor offset is outside the matching list", domain.ErrValidation)
		}
		offset = cursor.Offset
	}
	end := len(entries)
	if q.Limit > 0 && end-offset > q.Limit {
		end = offset + q.Limit
	}
	out := make([]T, 0, end-offset)
	for _, item := range entries[offset:end] {
		out = append(out, item.value)
	}
	page := &ListPage{Order: []string{q.Sort, "source_id", "value_digest"}, Limit: q.Limit,
		Returned: len(out), Matched: len(entries), Search: q.Search, UpdatedSince: q.UpdatedSince}
	if end < len(entries) {
		b, err := json.Marshal(listCursor{Version: listCursorVersion, Query: query, View: view, Offset: end})
		if err != nil {
			return ListResult[T]{}, err
		}
		page.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return ListResult[T]{Items: out, Page: page}, nil
}

func listDigest(value any) (string, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("%w: cannot describe list query: %v", domain.ErrValidation, err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func decodeListCursor(value string) (listCursor, error) {
	b, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil {
		return listCursor{}, fmt.Errorf("%w: invalid list cursor encoding", domain.ErrValidation)
	}
	var cursor listCursor
	decoder := json.NewDecoder(strings.NewReader(string(b)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cursor); err != nil {
		return listCursor{}, fmt.Errorf("%w: invalid list cursor payload", domain.ErrValidation)
	}
	if decoder.Decode(new(any)) != io.EOF || cursor.Version != listCursorVersion || len(cursor.Query) != 64 || len(cursor.View) != 64 {
		return listCursor{}, fmt.Errorf("%w: unsupported or malformed list cursor; restart the query", domain.ErrValidation)
	}
	return cursor, nil
}
