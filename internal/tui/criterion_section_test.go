package tui

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/charmbracelet/x/ansi"
)

func TestCriterionStructureIsVisibleWithoutHidingTaskBody(t *testing.T) {
	root := seedRepo(t)
	svc := core.MustNewService(testutil.Must(store.NewFS(root, core.UnrestrictedMutations())))
	body := "## Acceptance\n- [x] first\n## Acceptance criteria\n- [ ] second\n"
	if _, _, err := svc.ReplaceBody("alpha", body, false); err != nil {
		t.Fatal(err)
	}
	m := loadedAt(t, root, 160, 40)
	m = settleCreate(t, m, m.reloadAll())
	if m.detail.content == nil || m.detail.content.rawBody() != body {
		t.Fatal("criterion diagnostics hid raw body/editor repair context")
	}
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "unavailable:") || !strings.Contains(view, "acceptance-criteria sections") || strings.Contains(view, "0/0") {
		t.Fatalf("task detail concealed criterion structure as zero: %s", view)
	}
}
