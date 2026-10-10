package main

import (
	"encoding/base64"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/wire"
)

func TestSmoke_ListCursorMalformedDigestsAreValidationNotConflicts(t *testing.T) {
	root := t.TempDir()
	for _, args := range [][]string{
		{"init", "--no-register"},
		{"epic", "new", "Cursor smoke", "--description", "Process cursor validation"},
		{"task", "new", "Alpha", "--epic", "01-cursor-smoke", "--tags", "test", "--description", "Alpha"},
		{"task", "new", "Beta", "--epic", "01-cursor-smoke", "--tags", "test", "--description", "Beta"},
	} {
		if out, code := run(t, root, args...); code != 0 {
			t.Fatalf("fixture %v: exit %d\n%s", args, code, out)
		}
	}
	args := []string{"task", "list", "--json", "--limit", "1", "--sort", "name"}
	stdout, stderr, code := runStreams(t, root, args...)
	var page wire.TasksEnvelope
	if err := json.Unmarshal([]byte(stdout), &page); err != nil || code != 0 || stderr != "" || page.Page == nil || page.Page.NextCursor == "" {
		t.Fatalf("first page: exit %d err=%v\nstdout=%s\nstderr=%s", code, err, stdout, stderr)
	}
	payload, err := base64.RawURLEncoding.DecodeString(page.Page.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"q", "s"} {
		for _, value := range []string{strings.Repeat("g", 64), strings.Repeat("0", 63), strings.Repeat("A", 64)} {
			var cursor map[string]any
			if err := json.Unmarshal(payload, &cursor); err != nil {
				t.Fatal(err)
			}
			cursor[field] = value
			encoded, err := json.Marshal(cursor)
			if err != nil {
				t.Fatal(err)
			}
			stdout, stderr, code := runStreams(t, root, append(slices.Clone(args), "--cursor", base64.RawURLEncoding.EncodeToString(encoded))...)
			var failure wire.ErrorEnvelope
			if err := json.Unmarshal([]byte(stderr), &failure); err != nil || stdout != "" || code != 11 || failure.Error.Code != "validation" || !strings.Contains(failure.Error.Message, "malformed list cursor") {
				t.Fatalf("corrupt %s digest %q: exit %d err=%v\nstdout=%s\nstderr=%s", field, value, code, err, stdout, stderr)
			}
		}
	}
}
