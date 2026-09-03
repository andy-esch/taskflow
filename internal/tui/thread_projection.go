package tui

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
)

// threadListDiagnostics is the only Thread-specific runtime state that does not
// belong to a row. The entity registry owns it alongside the ordinary loaded/error
// state; keeping the repository graph evidence here avoids repeating it on every
// Thread while eliminating the former parallel list/detail state machine.
type threadListDiagnostics struct {
	graphHealth   core.GraphHealth
	graphProblems []core.GraphProblem
	readProblems  []core.ThreadReadProblem
}

// loadThreadList adapts core's shared Thread projection to ordinary registry
// rows. It never traverses the graph or derives readiness: each row retains the
// complete core.ThreadView and the renderer only presents its supplied fields.
func loadThreadList(t *entityTab, svc *core.Service) tea.Cmd {
	gen := t.loadGen
	return func() tea.Msg {
		view, problems, err := svc.ListThreadViews()
		if err != nil {
			return errMsg{kind: entityThreads, gen: gen, err: err}
		}
		refs := make([]entityRef, 0, len(view.Threads))
		for _, thread := range view.Threads {
			refs = append(refs, entityRef{key: thread.Thread.CanonicalID(), label: thread.Thread.Slug})
		}
		hints := duplicateIdentityHints(refs)
		countsW := countsWidth(view.Threads, func(v core.ThreadView) (int, int) {
			return v.Rollup.Done, v.Rollup.Total
		})
		items := make([]list.Item, 0, len(view.Threads))
		for _, thread := range view.Threads {
			items = append(items, threadItem{
				view: thread, countsW: countsW,
				identityHint: hints[thread.Thread.CanonicalID()],
			})
		}
		return listLoadedMsg{
			kind: entityThreads, gen: gen, items: items,
			threadDiagnostics: &threadListDiagnostics{
				graphHealth: view.GraphHealth, graphProblems: view.GraphProblems,
				readProblems: problems,
			},
		}
	}
}

// loadThreadDetail joins one coherent graph-projection/body read with the
// independently optional local-path capability. A pathless adapter still gets
// both summary and topology views. Only the semantic read participates in
// contention retry; optional path failure is explanatory and cannot blank
// otherwise coherent content.
func loadThreadDetail(svc *core.Service, id string) tea.Cmd {
	return func() tea.Msg {
		// Path lookup is optional and non-authoritative. Resolve it independently so
		// a local repair path survives a semantic projection failure, but never let a
		// missing/path-only failure blank otherwise coherent Thread content.
		path, pathErr := svc.ThreadPath(id)
		projection, body, err := svc.ShowThreadGraphDetail(id)
		if err != nil {
			return detailErrMsg{kind: entityThreads, id: id, err: err, localPath: path}
		}
		issue := ""
		if pathErr != nil {
			issue = pathErr.Error()
		}
		return detailMsg{kind: entityThreads, id: id, content: threadDetail{
			projection: projection, body: body, path: path, pathIssue: issue,
		}}
	}
}
