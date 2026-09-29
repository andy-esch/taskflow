package tui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

type duplicateEpicSourceStore struct {
	*store.FS
	records []core.LoadedRecord[domain.Epic]
}

func (s *duplicateEpicSourceStore) ReadEpics() (core.EpicRead, error) {
	return core.EpicRead{Records: s.records}, nil
}

func TestEntityViewFiltersCannotHideDuplicateSourceIDs(t *testing.T) {
	root := t.TempDir()
	taskID := testutil.TaskID("cross-view-task-source")
	testutil.Write(t, filepath.Join(root, domain.TasksDir, taskID+"-active.md"),
		"---\nid: "+taskID+"\nstatus: in-progress\ndescription: active occurrence\n---\n# Active\n")
	testutil.Write(t, filepath.Join(root, domain.TasksDir, taskID+"-archived.md"),
		"---\nid: "+testutil.TaskID("other-task-declaration")+"\nstatus: completed\ndescription: hidden occurrence\n---\n# Archived\n")
	auditID := testutil.TaskID("cross-view-audit-source")
	testutil.Write(t, filepath.Join(root, domain.AuditsDir, auditID+"-open.md"),
		"---\nid: "+auditID+"\nbucket: open\narea: review\ndate: 2026-09-28\n---\n# Open\n")
	testutil.Write(t, filepath.Join(root, domain.AuditsDir, auditID+"-closed.md"),
		"---\nid: "+testutil.TaskID("other-audit-declaration")+"\nbucket: closed\narea: review\ndate: 2026-09-28\n---\n# Closed\n")
	fs := store.NewFS(root)
	svc := core.MustNewService(fs)
	tasks := loadTaskList(&entityTab{statusView: "", loadGen: 1}, svc)().(listLoadedMsg)
	if len(tasks.items) != 1 || tasks.identityErr == nil || !strings.Contains(tasks.identityErr.Error(), "shared") {
		t.Fatalf("working task view hid duplicate identity: rows=%d err=%v", len(tasks.items), tasks.identityErr)
	}
	audits := loadAuditList(&entityTab{statusView: "", loadGen: 1}, svc)().(listLoadedMsg)
	if len(audits.items) != 1 || audits.identityErr == nil || !strings.Contains(audits.identityErr.Error(), "shared") {
		t.Fatalf("open audit view hid duplicate identity: rows=%d err=%v", len(audits.items), audits.identityErr)
	}

	const epicID = "01-shared-source"
	epicStore := &duplicateEpicSourceStore{FS: fs, records: []core.LoadedRecord[domain.Epic]{
		{Value: domain.Epic{ID: "declared-active", Status: "active"}, Source: core.RecordSource{ID: epicID}},
		{Value: domain.Epic{ID: "declared-retired", Status: "retired"}, Source: core.RecordSource{ID: epicID}},
	}}
	epics := loadEpicList(&entityTab{statusView: "", loadGen: 1}, core.MustNewService(epicStore))().(listLoadedMsg)
	if len(epics.items) != 1 || epics.identityErr == nil || !strings.Contains(epics.identityErr.Error(), "shared") {
		t.Fatalf("live epic view hid duplicate identity: rows=%d err=%v", len(epics.items), epics.identityErr)
	}

	// A refresh may retain old rows for context, but this full-snapshot failure
	// must quarantine every old action and navigation target.
	m := loaded(t, 120, 40)
	tasks.gen = m.cur().loadGen
	tm, _ := m.Update(tasks)
	m = tm.(Model)
	if !m.cur().identityInvalid || !m.selectedRef().empty() || m.selectedPath() != "" {
		t.Fatalf("hidden duplicate left retained rows actionable: invalid=%v ref=%+v path=%q",
			m.cur().identityInvalid, m.selectedRef(), m.selectedPath())
	}
}
