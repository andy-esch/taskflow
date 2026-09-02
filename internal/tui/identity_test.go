package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

type duplicateEntityFixture struct {
	root        string
	taskIDs     [2]string
	auditIDs    [2]string
	researchIDs [2]string
}

func duplicateEntityRepo(t *testing.T) duplicateEntityFixture {
	t.Helper()
	r := testutil.NewRepo(t)
	f := duplicateEntityFixture{
		root:        r.Root,
		taskIDs:     [2]string{testutil.TaskID("duplicate-task-a"), testutil.TaskID("duplicate-task-b")},
		auditIDs:    [2]string{testutil.TaskID("duplicate-audit-a"), testutil.TaskID("duplicate-audit-b")},
		researchIDs: [2]string{testutil.TaskID("duplicate-research-a"), testutil.TaskID("duplicate-research-b")},
	}
	r.Epic("01-identity.md", "---\nstatus: active\ndescription: identity fixtures\n---\n# Identity fixtures\n")
	for i, id := range f.taskIDs {
		// The first record deliberately drifts; the second omits frontmatter id.
		// Both must remain addressable by the canonical filename identity.
		idLine := ""
		if i == 0 {
			idLine = "id: " + testutil.TaskID("wrong-task-frontmatter") + "\n"
		}
		r.File(filepath.Join(domain.TasksDir, id+"-same-task.md"), fmt.Sprintf(
			"---\n%sstatus: in-progress\nepic: 01-identity\ndescription: duplicate %d\ntags: [identity]\n---\n# Same task\n\nTASK-%d-BODY\n", idLine, i+1, i+1))
	}
	for i, id := range f.auditIDs {
		idLine := ""
		if i == 0 {
			idLine = "id: " + testutil.TaskID("wrong-audit-frontmatter") + "\n"
		}
		r.File(filepath.Join(domain.AuditsDir, id+"-same-audit.md"), fmt.Sprintf(
			"---\n%sbucket: open\narea: identity\ndate: 2026-09-0%d\n---\n# Same audit\n\nAUDIT-%d-BODY\n", idLine, i+1, i+1))
	}
	for i, id := range f.researchIDs {
		idLine := ""
		if i == 0 {
			idLine = "id: " + testutil.TaskID("wrong-research-frontmatter") + "\n"
		}
		r.File(filepath.Join(domain.ResearchDir, id+"-same-research.md"), fmt.Sprintf(
			"---\n%screated: \"2026-09-0%d\"\ndescription: duplicate %d\ntags: [identity]\n---\n# Same research\n\nRESEARCH-%d-BODY\n", idLine, i+1, i+1, i+1))
	}
	return f
}

func duplicateMutationRepo(t *testing.T) duplicateEntityFixture {
	t.Helper()
	r := testutil.NewRepo(t)
	f := duplicateEntityFixture{
		root:     r.Root,
		taskIDs:  [2]string{testutil.TaskID("mutation-task-a"), testutil.TaskID("mutation-task-b")},
		auditIDs: [2]string{testutil.TaskID("mutation-audit-a"), testutil.TaskID("mutation-audit-b")},
	}
	r.Epic("01-identity.md", "---\nstatus: active\ndescription: identity fixtures\n---\n# Identity fixtures\n")
	for i, entityID := range f.taskIDs {
		r.File(filepath.Join(domain.TasksDir, entityID+"-same-task.md"), fmt.Sprintf(
			"---\nid: %s\nstatus: in-progress\nepic: 01-identity\ndescription: duplicate %d\ntags: [identity]\n---\n# Same task\n", entityID, i+1))
	}
	for i, entityID := range f.auditIDs {
		r.File(filepath.Join(domain.AuditsDir, entityID+"-same-audit.md"), fmt.Sprintf(
			"---\nid: %s\nbucket: open\narea: identity\ndate: 2026-09-0%d\n---\n# Same audit\n", entityID, i+1))
	}
	return f
}

func TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState(t *testing.T) {
	f := duplicateEntityRepo(t)
	m := loadedAt(t, f.root, 120, 40)
	items := m.cur().list.Items()
	if len(items) != 2 {
		t.Fatalf("task rows = %d, want both duplicate slugs", len(items))
	}
	first, second := items[0].(taskItem), items[1].(taskItem)
	if first.ref().label != second.ref().label || first.ref().key == second.ref().key {
		t.Fatalf("fixture refs are not duplicate labels with distinct keys: %+v %+v", first.ref(), second.ref())
	}
	if first.displayLabel() == second.displayLabel() || !strings.Contains(first.displayLabel(), "[") || !strings.Contains(second.displayLabel(), "[") {
		t.Fatalf("duplicate task labels were not deterministically disambiguated: %q / %q", first.displayLabel(), second.displayLabel())
	}
	if !strings.Contains(first.FilterValue(), first.ref().key) || !strings.Contains(second.FilterValue(), second.ref().key) {
		t.Fatal("canonical keys must remain filterable even though rows show slugs")
	}
	m.cur().list.SetFilterText(second.ref().key)
	visible := m.cur().list.VisibleItems()
	if len(visible) != 1 || visible[0].(entityItem).ref().key != second.ref().key {
		t.Fatalf("stable-key filter did not isolate the second duplicate: %+v", visible)
	}
	m.cur().list.ResetFilter()

	// Select the second duplicate by key and prove an out-of-order result for the
	// first cannot land over it.
	if !m.cur().selectByKey(second.ref().key) {
		t.Fatal("second duplicate was not independently selectable")
	}
	m.detailGen++
	tm, _ := m.Update(detailMsg{kind: entityTasks, id: first.ref().key, gen: m.detailGen,
		content: taskDetail{t: first.t, body: "WRONG-BODY"}})
	m = tm.(Model)
	if m.detail.hasContent {
		t.Fatal("a stale result for the other duplicate landed")
	}
	tm, _ = m.Update(detailMsg{kind: entityTasks, id: second.ref().key, gen: m.detailGen,
		content: taskDetail{t: second.t, body: "SECOND-BODY"}})
	m = tm.(Model)
	if !m.detail.hasContent || m.detail.loadedKey != second.ref().key {
		t.Fatalf("matching duplicate detail did not land: loaded=%q", m.detail.loadedKey)
	}

	// Manual/watcher reloads share markReload/reload. The selected canonical key
	// must survive even though the visible labels collide.
	restore := m.cur().markReload()
	if restore != second.ref() {
		t.Fatalf("reload captured %+v, want canonical ref %+v", restore, second.ref())
	}
	m = pump(t, m, m.cur().reload(m.svc, restore), 8)
	if m.selectedKey() != second.ref().key {
		t.Fatalf("reload restored %q, want second duplicate %q", m.selectedKey(), second.ref().key)
	}

	// Palette targets, back-stack entries, and cached sessions retain the same key
	// while continuing to display a readable, disambiguated slug.
	var duplicatePalette []paletteItem
	for _, p := range m.paletteIndex() {
		if p.kind == palJump && p.ek == entityTasks && p.ref.label == "same-task" {
			duplicatePalette = append(duplicatePalette, p)
		}
	}
	if len(duplicatePalette) != 2 || duplicatePalette[0].ref.key == duplicatePalette[1].ref.key || duplicatePalette[0].title == duplicatePalette[1].title {
		t.Fatalf("palette collapsed duplicate targets: %+v", duplicatePalette)
	}
	m.pushLoc()
	if len(m.navStack) != 1 || m.navStack[0].ref.key != second.ref().key {
		t.Fatalf("back stack did not retain the selected canonical key: %+v", m.navStack)
	}
	m.workspace = core.Workspace{Checkout: f.root, PlanningRoot: f.root, Planning: m.svc}
	m.saveSession()
	saved := m.sessions[workspaceKey(m.workspace)]
	if len(saved.navStack) != 1 || saved.navStack[0].ref.key != second.ref().key {
		t.Fatalf("workspace session did not retain the canonical navigation key: %+v", saved.navStack)
	}
	if !m.cur().selectByKey(first.ref().key) {
		t.Fatal("first duplicate was not independently selectable")
	}
	tm, cmd := m.navBack()
	m = drain(t, tm.(Model), cmd)
	if m.selectedKey() != second.ref().key {
		t.Fatalf("back navigation restored %q, want second duplicate %q", m.selectedKey(), second.ref().key)
	}

	// Mutation commands consume the same ref contract: editing the selected duplicate
	// must never re-resolve its colliding slug.
	msg := setFieldCmd(m.svc, second.ref(), "description", "changed second duplicate")()
	if _, ok := msg.(editedMsg); !ok {
		t.Fatalf("stable-key edit failed: %T %+v", msg, msg)
	}
	left, _, err := m.svc.ShowTask(first.ref().key)
	if err != nil {
		t.Fatal(err)
	}
	right, _, err := m.svc.ShowTask(second.ref().key)
	if err != nil {
		t.Fatal(err)
	}
	if left.Description == right.Description || right.Description != "changed second duplicate" {
		t.Fatalf("edit crossed duplicate identities: first=%q second=%q", left.Description, right.Description)
	}
}

func TestDuplicateStableIDEntitiesLoadByFilenameIdentity(t *testing.T) {
	f := duplicateEntityRepo(t)
	m := loadedAt(t, f.root, 120, 40)

	tests := []struct {
		name       string
		command    string
		kind       entityKind
		wantKey    string
		wantMarker string
		rows       func(Model) (entityItem, entityItem)
	}{
		{
			name: "tasks", command: "tasks", kind: entityTasks, wantKey: f.taskIDs[1], wantMarker: "TASK-2-BODY",
			rows: func(m Model) (entityItem, entityItem) {
				return m.cur().list.Items()[0].(taskItem), m.cur().list.Items()[1].(taskItem)
			},
		},
		{
			name: "audits", command: "audits", kind: entityAudits, wantKey: f.auditIDs[1], wantMarker: "AUDIT-2-BODY",
			rows: func(m Model) (entityItem, entityItem) {
				return m.cur().list.Items()[0].(auditItem), m.cur().list.Items()[1].(auditItem)
			},
		},
		{
			name: "research", command: "research", kind: entityResearch, wantKey: f.researchIDs[1], wantMarker: "RESEARCH-2-BODY",
			rows: func(m Model) (entityItem, entityItem) {
				return m.cur().list.Items()[0].(researchItem), m.cur().list.Items()[1].(researchItem)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if m.cur().kind != tt.kind {
				m = cmdJump(t, m, tt.command)
			}
			left, right := tt.rows(m)
			if left.ref().label != right.ref().label || left.ref().key == right.ref().key {
				t.Fatalf("rows do not expose duplicate labels with distinct canonical keys: %+v %+v", left.ref(), right.ref())
			}
			if left.displayLabel() == right.displayLabel() {
				t.Fatalf("duplicate visible labels were not disambiguated: %q", left.displayLabel())
			}
			if !m.cur().selectByKey(tt.wantKey) {
				t.Fatalf("canonical filename identity %q was not selectable", tt.wantKey)
			}
			m = drain(t, m, m.refreshDetail())
			if m.detail.loadedKey != tt.wantKey || m.detail.content == nil || !strings.Contains(m.detail.content.rawBody(), tt.wantMarker) {
				t.Fatalf("detail resolved an arbitrary duplicate: loaded=%q body=%q", m.detail.loadedKey, m.detail.content.rawBody())
			}
		})
	}
}

func TestDuplicateIdentityHintsExpandPastACollidingPrefix(t *testing.T) {
	refs := []entityRef{
		{key: "abcdef111111", label: "same"},
		{key: "abcdef222222", label: "same"},
		{key: "zzzzzz333333", label: "unique"},
	}
	hints := duplicateIdentityHints(refs)
	if hints[refs[0].key] != "abcdef1" || hints[refs[1].key] != "abcdef2" {
		t.Fatalf("colliding six-character prefixes were not expanded: %+v", hints)
	}
	if _, ok := hints[refs[2].key]; ok {
		t.Fatal("a unique visible label should not expose an identity hint")
	}

	prefixKeys := []entityRef{{key: "abcdef", label: "prefix"}, {key: "abcdef123", label: "prefix"}}
	prefixHints := duplicateIdentityHints(prefixKeys)
	if prefixHints["abcdef"] != "abcdef" || prefixHints["abcdef123"] != "abcdef1" {
		t.Fatalf("strict-prefix keys need explicit hints on both rows: %+v", prefixHints)
	}
}

func TestDuplicateIdentityHintsLeadRowsSoTruncationCannotHideThem(t *testing.T) {
	const longSlug = "make-tui-entity-navigation-use-stable-identities-with-a-long-shared-name"
	refs := []entityRef{
		{key: "aaaaaa111111", label: longSlug},
		{key: "bbbbbb222222", label: longSlug},
	}
	hints := duplicateIdentityHints(refs)

	tests := []struct {
		name   string
		width  int
		rowFor func(entityRef) string
	}{
		{
			name: "tasks", width: 46,
			rowFor: func(ref entityRef) string {
				item := taskItem{t: domain.Task{ID: ref.key, Slug: ref.label, Status: domain.StatusInProgress}, identityHint: hints[ref.key]}
				return renderDelegateRow(t, taskDelegate{st: &testStyles}, item, 46)
			},
		},
		{
			name: "audits", width: 60,
			rowFor: func(ref entityRef) string {
				item := auditItem{a: domain.Audit{ID: ref.key, Slug: ref.label, Bucket: domain.AuditOpen}, identityHint: hints[ref.key]}
				return renderDelegateRow(t, auditDelegate{st: &testStyles}, item, 60)
			},
		},
		{
			name: "research", width: 46,
			rowFor: func(ref entityRef) string {
				item := researchItem{r: domain.Research{ID: ref.key, Slug: ref.label, Created: "2026-09-02"}, identityHint: hints[ref.key]}
				return renderDelegateRow(t, researchDelegate{st: &testStyles}, item, 46)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left, right := tt.rowFor(refs[0]), tt.rowFor(refs[1])
			if left == right || !strings.Contains(left, "[aaaaaa]") || !strings.Contains(right, "[bbbbbb]") {
				t.Fatalf("width %d hid duplicate identity:\nleft:  %q\nright: %q", tt.width, left, right)
			}
		})
	}
}

func TestDuplicateDetailErrorsKeepTheFriendlyLabel(t *testing.T) {
	f := duplicateEntityRepo(t)
	m := loadedAt(t, f.root, 120, 40)
	second := m.cur().list.Items()[1].(taskItem)
	if !m.cur().selectByKey(second.ref().key) {
		t.Fatal("second duplicate was not selectable")
	}
	if err := os.Remove(second.path()); err != nil {
		t.Fatal(err)
	}
	cmd := m.refreshDetail()
	if cmd == nil {
		t.Fatal("refresh did not issue a detail read")
	}
	tm, _ := m.Update(cmd())
	m = tm.(Model)
	if m.detail.title != second.ref().label || m.detail.title == second.ref().key {
		t.Fatalf("failed detail title = %q, want friendly label %q", m.detail.title, second.ref().label)
	}
}

func TestDuplicateIdentityMutationCommandsUseCanonicalKeys(t *testing.T) {
	f := duplicateMutationRepo(t)
	m := loadedAt(t, f.root, 120, 40)
	first := m.cur().list.Items()[0].(taskItem)
	second := m.cur().list.Items()[1].(taskItem)

	var editor editMenu
	editor.open(second.t)
	if editor.ref != second.ref() {
		t.Fatalf("edit form captured %+v, want %+v", editor.ref, second.ref())
	}

	if _, err := m.svc.SetFields(second.ref().key, map[string]any{"revisit_at": "2099-01-01"}, false, false); err != nil {
		t.Fatal(err)
	}
	msg := unsetFieldCmd(m.svc, second.ref(), "revisit_at")()
	if _, ok := msg.(editedMsg); !ok {
		t.Fatalf("canonical unset failed: %T %+v", msg, msg)
	}
	left, _, err := m.svc.ShowTask(first.ref().key)
	if err != nil {
		t.Fatal(err)
	}
	right, _, err := m.svc.ShowTask(second.ref().key)
	if err != nil {
		t.Fatal(err)
	}
	if left.RevisitAt != "" || right.RevisitAt != "" {
		t.Fatalf("unset crossed duplicate identities: first=%q second=%q", left.RevisitAt, right.RevisitAt)
	}

	moveResult := moveTask(m.svc, second.ref(), transition{to: string(domain.StatusNextUp)})()
	move, ok := moveResult.(movedMsg)
	if !ok || move.ref.key != second.ref().key {
		if failure, failed := moveResult.(actionErrMsg); failed {
			t.Fatalf("canonical lifecycle move failed: %v", failure.err)
		}
		t.Fatalf("canonical lifecycle move failed: %T %+v", moveResult, moveResult)
	}
	right, _, err = m.svc.ShowTask(second.ref().key)
	if err != nil || right.Status != domain.StatusNextUp {
		t.Fatalf("second duplicate status = %q, %v; want next-up", right.Status, err)
	}
	left, _, err = m.svc.ShowTask(first.ref().key)
	if err != nil || left.Status != domain.StatusInProgress {
		t.Fatalf("first duplicate was changed by sibling move: %q, %v", left.Status, err)
	}

	deferResult := deferTaskCmd(m.svc, first.ref(), "")()
	deferred, ok := deferResult.(movedMsg)
	if !ok || deferred.ref.key != first.ref().key {
		t.Fatalf("canonical defer failed: %T %+v", deferResult, deferResult)
	}
	left, _, err = m.svc.ShowTask(first.ref().key)
	if err != nil || left.Status != domain.StatusDeferred {
		t.Fatalf("first duplicate status = %q, %v; want deferred", left.Status, err)
	}

	tm, _ := m.Update(move)
	m = tm.(Model)
	if m.movedAwayKey != second.ref().key {
		t.Fatalf("post-move suppression key = %q, want %q", m.movedAwayKey, second.ref().key)
	}

	auditFirst, _, err := m.svc.ShowAudit(f.auditIDs[0])
	if err != nil {
		t.Fatal(err)
	}
	auditSecond, _, err := m.svc.ShowAudit(f.auditIDs[1])
	if err != nil {
		t.Fatal(err)
	}
	auditResult := moveAudit(m.svc,
		entityRef{key: auditSecond.CanonicalID(), label: auditSecond.Slug},
		transition{to: string(domain.AuditClosed)})()
	if moved, ok := auditResult.(movedMsg); !ok || moved.ref.key != auditSecond.CanonicalID() {
		t.Fatalf("canonical audit move failed: %T %+v", auditResult, auditResult)
	}
	auditFirst, _, err = m.svc.ShowAudit(auditFirst.CanonicalID())
	if err != nil || auditFirst.Bucket != domain.AuditOpen {
		t.Fatalf("first duplicate audit changed: %q, %v", auditFirst.Bucket, err)
	}
	auditSecond, _, err = m.svc.ShowAudit(auditSecond.CanonicalID())
	if err != nil || auditSecond.Bucket != domain.AuditClosed {
		t.Fatalf("second duplicate audit status = %q, %v; want closed", auditSecond.Bucket, err)
	}
}

func TestDuplicateYankCopiesTheUsableCanonicalKey(t *testing.T) {
	f := duplicateEntityRepo(t)
	m := loadedAt(t, f.root, 120, 40)
	second := m.cur().list.Items()[1].(taskItem)
	if !m.cur().selectByKey(second.ref().key) {
		t.Fatal("second duplicate was not selectable")
	}
	tm, cmd := m.Update(press("y"))
	m = tm.(Model)
	if m.flash != "copied id: "+second.ref().key || m.flashErr || cmd == nil {
		t.Fatalf("duplicate yank = %q (err=%v cmd=%v), want canonical id", m.flash, m.flashErr, cmd != nil)
	}
}

func TestEntityRegistryRejectsEmptyOrDuplicateCanonicalKeys(t *testing.T) {
	m := loaded(t, 120, 40)
	tab := m.cur()
	wantRows := len(tab.list.Items())
	tests := []struct {
		name  string
		items []list.Item
		want  string
	}{
		{
			name:  "empty",
			items: []list.Item{taskItem{t: domain.Task{Slug: "missing-identity"}}},
			want:  "has no canonical identity",
		},
		{
			name: "duplicate",
			items: []list.Item{
				taskItem{t: domain.Task{ID: "shared-key", Slug: "first"}},
				taskItem{t: domain.Task{ID: "shared-key", Slug: "second"}},
			},
			want: "is shared by",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tab.loadGen++
			tm, _ := m.Update(listLoadedMsg{kind: entityTasks, gen: tab.loadGen, items: tt.items})
			m = tm.(Model)
			if tab.loadErr == nil || !strings.Contains(tab.loadErr.Error(), tt.want) {
				t.Fatalf("identity violation = %v, want %q", tab.loadErr, tt.want)
			}
			if got := len(tab.list.Items()); got != wantRows {
				t.Fatalf("invalid rows replaced last coherent list: got %d rows, want %d", got, wantRows)
			}
		})
	}
}

func TestPaletteDisambiguatesCrossKindLabelsWithoutDuplicatingKeys(t *testing.T) {
	m := Model{tabs: newEntityTabs(&testStyles)}
	taskKey, auditKey := "aaaaaa111111", "bbbbbb222222"
	m.tabs[indexOfKind(m.tabs, entityTasks)].list.SetItems([]list.Item{
		taskItem{t: domain.Task{ID: taskKey, Slug: "shared"}},
	})
	m.tabs[indexOfKind(m.tabs, entityAudits)].list.SetItems([]list.Item{
		auditItem{a: domain.Audit{ID: auditKey, Slug: "shared"}},
	})

	titles := map[string]string{}
	for _, item := range m.paletteIndex() {
		if item.kind != palJump || item.ref.label != "shared" {
			continue
		}
		titles[item.entity] = item.title
		if strings.Count(item.filter, item.ref.key) != 1 {
			t.Errorf("%s palette key appears more than once in %q", item.entity, item.filter)
		}
	}
	if titles["tasks"] != "shared · tasks" || titles["audits"] != "shared · audits" {
		t.Fatalf("cross-kind palette titles = %+v", titles)
	}
}

func TestEpicTaskRosterDisambiguatesDuplicateSlugs(t *testing.T) {
	tasks := []domain.Task{
		{ID: "aaaaaa111111", Slug: "same-task", Status: domain.StatusInProgress},
		{ID: "bbbbbb222222", Slug: "same-task", Status: domain.StatusNextUp},
	}
	meta := renderEpicMeta(core.EpicSummary{Epic: domain.Epic{ID: "01-identity"}}, tasks, 80, &testStyles)
	if !strings.Contains(meta, "[aaaaaa]") || !strings.Contains(meta, "[bbbbbb]") {
		t.Fatalf("epic roster did not distinguish duplicate tasks:\n%s", meta)
	}
}

func TestDashboardInProgressRowsCarryCanonicalDuplicateTargets(t *testing.T) {
	tasks := []domain.Task{
		{ID: "aaaaaa111111", Slug: "same-task", Status: domain.StatusInProgress},
		{ID: "bbbbbb222222", Slug: "same-task", Status: domain.StatusInProgress},
	}
	var d dashboard
	d.setSummary(core.Summary{InProgress: tasks}, &testStyles, false)
	seen := map[string]bool{}
	for _, row := range d.rows {
		if row.target == nil || row.target.kind != entityTasks || row.target.ref.label != "same-task" {
			continue
		}
		seen[row.target.ref.key] = true
		if !strings.Contains(row.text, "["+row.target.ref.key[:6]+"]") {
			t.Errorf("dashboard row lost its visible identity hint: %q", row.text)
		}
	}
	if !seen[tasks[0].CanonicalID()] || !seen[tasks[1].CanonicalID()] {
		t.Fatalf("dashboard duplicate targets = %+v", seen)
	}
}
