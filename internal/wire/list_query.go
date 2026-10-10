package wire

import "github.com/andy-esch/taskflow/internal/core"

// ListPageJSON accompanies opt-in list queries in typed and projected JSON.
// Repeat the same filters/order/limit with next_cursor via --cursor. A changed
// matching view or diagnostic set requires restarting; this is not a lock,
// authentication capability, or promise that the adapter avoided a full scan.
type ListPageJSON struct {
	Order        []string `json:"order"`
	Limit        int      `json:"limit" jsonschema:"minimum=0,maximum=1000,description=zero means unbounded"`
	Returned     int      `json:"returned" jsonschema:"minimum=0"`
	Matched      int      `json:"matched" jsonschema:"minimum=0"`
	Search       string   `json:"search,omitempty"`
	UpdatedSince string   `json:"updated_since,omitempty"`
	NextCursor   string   `json:"next_cursor,omitempty"`
}

func ToListPageJSON(page *core.ListPage) *ListPageJSON {
	if page == nil {
		return nil
	}
	return &ListPageJSON{Order: page.Order, Limit: page.Limit, Returned: page.Returned,
		Matched: page.Matched, Search: page.Search, UpdatedSince: page.UpdatedSince, NextCursor: page.NextCursor}
}
