package tui

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

var documentCreateKinds = map[string]entityKind{"audit": entityAudits, "research": entityResearch}

func fillDocumentCreate(m Model) Model {
	m = pasteCreation(m, "Review / qhjkn 世界")
	if m.creation.kind == entityResearch {
		m, _ = createKey(m, "enter")
		m = pasteCreation(m, "Compare approaches α")
		m, _ = createKey(m, "enter")
		m = pasteCreation(m, "planning, tui")
	}
	return m
}

func submitDocumentCreate(m Model) (Model, tea.Cmd) {
	controls := creationContractControls(&m)
	controls.focus(controls.rows[len(controls.rows)-1])
	return createKey(m, "enter")
}

func TestAuditCreateDurableDefaultsAndDates(t *testing.T) {
	for _, date := range []string{"", "2025-01-03"} {
		t.Run("date="+date, func(t *testing.T) {
			root := testutil.NewRepo(t).Root // creation must work with no existing audits
			fs := testutil.Must(store.NewFS(root, core.UnrestrictedMutations()))
			now := time.Date(2026, 10, 20, 23, 59, 0, 0, time.UTC)
			m := loadedAt(t, root, 120, 40)
			m.svc = core.MustNewService(fs, core.WithClock(func() time.Time { return now }))
			m = openCreationKind(t, m, entityAudits)
			if m.creation.audit.dirty() {
				t.Fatal("untouched date default made the form dirty")
			}
			m = fillDocumentCreate(m)
			m, _ = createKey(m, "enter") // area -> date
			m = pasteCreation(m, date)
			before := testutil.SnapshotTree(t, root)
			now = now.Add(24 * time.Hour) // default is core's submission date, not opening date
			m, write := submitDocumentCreate(m)
			m, duplicate := createKey(m, "enter")
			if write == nil || duplicate != nil || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("submission wrote early or issued twice")
			}
			m = settleCreate(t, m, write)
			ref := m.selectedRef()
			record, err := m.svc.ShowAudit(ref.key)
			if err != nil {
				t.Fatal(err)
			}
			wantDate := date
			if date == "" {
				wantDate = "2026-10-21"
			}
			a := record.Value.Audit
			if m.activeCreation() != nil || record.Source.ID != ref.key || a.ID != ref.key || a.Bucket != domain.AuditOpen ||
				a.Area != "Review / qhjkn 世界" || a.Date != wantDate || a.Slug != wantDate+"-review-qhjkn-世界" ||
				!strings.Contains(record.Value.Body, "# Audit: Review / qhjkn 世界") || !strings.Contains(record.Value.Body, "## Findings") || a.Findings != 0 {
				t.Fatalf("wrong durable audit/scaffold: %+v\n%s", a, record.Value.Body)
			}
			after := testutil.SnapshotTree(t, root)
			delete(after, filepath.Join(domain.AuditsDir, ref.key+"-"+a.Slug+".md"))
			if _, exists := before[domain.AuditsDir]; !exists {
				delete(after, domain.AuditsDir)
			}
			if !maps.Equal(before, after) {
				t.Fatal("creation changed unrelated records or modes")
			}
		})
	}
}

func TestResearchCreateDurableDateAndOptionalMetadata(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(fmt.Sprint(explicit), func(t *testing.T) {
			root := testutil.NewRepo(t).Root
			fs := testutil.Must(store.NewFS(root, core.UnrestrictedMutations()))
			now := time.Date(2026, 10, 20, 23, 59, 0, 0, time.UTC)
			m := loadedAt(t, root, 120, 40)
			m.svc = core.MustNewService(fs, core.WithClock(func() time.Time { return now }))
			m = openCreationKind(t, m, entityResearch)
			if m.creation.research.dirty() {
				t.Fatal("untouched research was dirty")
			}
			if explicit {
				m = fillDocumentCreate(m)
				m, _ = createKey(m, "enter") // tags -> created date
				m = pasteCreation(m, "2025-01-03")
			} else {
				m = pasteCreation(m, "Review / qhjkn 世界") // description/tags/date are optional
			}
			before := testutil.SnapshotTree(t, root)
			now = now.Add(24 * time.Hour)
			m, write := submitDocumentCreate(m)
			m, duplicate := createKey(m, "enter")
			if duplicate != nil || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("write before command or duplicate submission")
			}
			m = settleCreate(t, m, write)
			ref := m.selectedRef()
			record, err := m.svc.ShowResearch(ref.key)
			if err != nil {
				t.Fatal(err)
			}
			r := record.Value.Research
			wantDate, wantDescription := "2026-10-21", ""
			var wantTags []string
			if explicit {
				wantDate, wantDescription, wantTags = "2025-01-03", "Compare approaches α", []string{"planning", "tui"}
			}
			if m.activeCreation() != nil || r.ID != ref.key || record.Source.ID != ref.key || r.Slug != "review-qhjkn-世界" ||
				r.Created != wantDate || id.Time(r.ID).UTC().Format("2006-01-02") != wantDate || r.Description != wantDescription ||
				!slices.Equal(r.Tags, wantTags) || !strings.Contains(record.Value.Body, "# Review / qhjkn 世界") || !strings.Contains(record.Value.Body, "## Question") {
				t.Fatalf("wrong durable research/scaffold: %+v\n%s", r, record.Value.Body)
			}
			after := testutil.SnapshotTree(t, root)
			delete(after, filepath.Join(domain.ResearchDir, ref.key+"-"+r.Slug+".md"))
			if _, exists := before[domain.ResearchDir]; !exists {
				delete(after, domain.ResearchDir)
			}
			if !maps.Equal(before, after) {
				t.Fatal("research creation changed unrelated records or modes")
			}
		})
	}
}

func TestDocumentCreateValidationAndDeliberateRetry(t *testing.T) {
	for name, kind := range documentCreateKinds {
		for _, bad := range []string{"missing title", "invalid date", "unmintable date"} {
			if kind == entityAudits && bad == "unmintable date" {
				continue
			}
			t.Run(name+"/"+bad, func(t *testing.T) {
				root := seedRepo(t)
				m := openCreationKind(t, loadedAt(t, root, 120, 40), kind)
				if bad != "missing title" {
					m = fillDocumentCreate(m)
				}
				dateRow := auditCreateDate
				if kind == entityResearch {
					dateRow = researchCreateDate
				}
				if bad != "missing title" {
					creationContractControls(&m).focus(dateRow)
					date := "2026-02-30"
					if bad == "unmintable date" {
						date = "1026-01-01"
					}
					m = pasteCreation(m, date)
				}
				before := testutil.SnapshotTree(t, root)
				m, cmd := submitDocumentCreate(m)
				m = settleCreate(t, m, cmd)
				if f := m.activeCreation(); f == nil || f.locked || f.saving || f.err == "" || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatal("core refusal lost input, wrote, or disabled deliberate retry")
				}
				creationContractControls(&m).focus(0)
				creationContractInputs(&m)[0].SetValue("")
				m = pasteCreation(m, "Corrected intent")
				creationContractControls(&m).focus(dateRow)
				creationContractInputs(&m)[dateRow].SetValue("")
				m = pasteCreation(m, "2026-10-01")
				m, cmd = submitDocumentCreate(m)
				m = settleCreate(t, m, cmd)
				if m.activeCreation() != nil || m.selectedRef().empty() {
					t.Fatal("explicit repair/retry did not land")
				}
			})
		}
	}
}

// No local-path interfaces: reads can forward to a real Store, but typed create
// outcomes and opaque identity are independently controlled at the port boundary.
type documentCreatePort struct {
	core.Store
	setID     core.SourceSetID
	calls     int
	audit     domain.Audit
	research  domain.Research
	body      string
	outcome   string
	readCalls int
}

func (p *documentCreatePort) SourceSetID() core.SourceSetID { return p.setID }

func (p *documentCreatePort) createError() error {
	switch p.outcome {
	case "read-only":
		return core.ErrReadOnlyPersistence
	case "conflict", "collision once", "committed refusal":
		return domain.ErrConflict
	case "path evidence":
		return domain.ErrValidation
	case "unknown":
		return errors.New("transport outcome unknown")
	case "finalization", "committed cleanup":
		return &core.CreateFinalizationError{Cause: domain.ErrConflict}
	}
	return nil
}

func (p *documentCreatePort) CreateAudit(a domain.Audit, body string, dry bool) (core.AuditCreationReceipt, error) {
	p.calls++
	p.audit, p.body = a, body
	if p.outcome == "pathless" {
		a.ID = "urn:audit:created"
		p.audit = a
		return core.AuditCreationReceipt{Audit: a, Committed: true}, nil
	}
	if p.outcome == "committed cleanup" || p.outcome == "success" {
		receipt, err := p.Store.CreateAudit(a, body, dry)
		if err != nil {
			return receipt, err
		}
		return receipt, p.createError()
	}
	receipt := core.AuditCreationReceipt{}
	if p.outcome == "committed refusal" {
		receipt = core.AuditCreationReceipt{Audit: a, Committed: true}
	}
	if p.outcome == "preview" {
		receipt = core.AuditCreationReceipt{Audit: a, Committed: true, DryRun: true}
	}
	if p.outcome == "path evidence" {
		receipt.Local.CommittedPath = "/opaque/not-authority"
	}
	return receipt, p.createError()
}

func (p *documentCreatePort) CreateResearch(r domain.Research, body string, dry bool) (core.ResearchCreationReceipt, error) {
	p.calls++
	p.research, p.body = r, body
	if p.outcome == "pathless" {
		r.ID = "urn:research:created"
		p.research = r
		return core.ResearchCreationReceipt{Research: r, Committed: true}, nil
	}
	if p.outcome == "success" || p.outcome == "committed cleanup" || (p.outcome == "collision once" && p.calls > 1) {
		receipt, err := p.Store.CreateResearch(r, body, dry)
		if err != nil {
			return receipt, err
		}
		if p.outcome == "collision once" {
			return receipt, nil
		}
		return receipt, p.createError()
	}
	receipt := core.ResearchCreationReceipt{}
	if p.outcome == "committed refusal" {
		receipt = core.ResearchCreationReceipt{Research: r, Committed: true}
	}
	if p.outcome == "preview" {
		receipt = core.ResearchCreationReceipt{Research: r, Committed: true, DryRun: true}
	}
	if p.outcome == "path evidence" {
		receipt.Local.CommittedPath = "/opaque/not-authority"
	}
	return receipt, p.createError()
}

func (p *documentCreatePort) ReadAudits() (core.AuditRead, error) {
	p.readCalls++
	if p.outcome == "pathless" && p.audit.ID != "" {
		return core.AuditRead{Records: []core.LoadedRecord[domain.Audit]{{Value: p.audit, Source: core.RecordSource{ID: p.audit.ID}}}}, nil
	}
	return p.Store.ReadAudits()
}

func (p *documentCreatePort) ReadAudit(ref string) (core.LoadedRecord[core.AuditWithBody], error) {
	p.readCalls++
	if p.outcome == "pathless" && ref == p.audit.ID {
		return core.LoadedRecord[core.AuditWithBody]{Value: core.AuditWithBody{Audit: p.audit, Body: p.body}, Source: core.RecordSource{ID: p.audit.ID}}, nil
	}
	return p.Store.ReadAudit(ref)
}

func (p *documentCreatePort) ReadResearch() (core.ResearchRead, error) {
	p.readCalls++
	if p.outcome == "pathless" && p.research.ID != "" {
		return core.ResearchRead{Records: []core.LoadedRecord[domain.Research]{{Value: p.research, Source: core.RecordSource{ID: p.research.ID}}}}, nil
	}
	return p.Store.ReadResearch()
}

func (p *documentCreatePort) ReadResearchDocument(ref string) (core.LoadedRecord[core.ResearchWithBody], error) {
	p.readCalls++
	if p.outcome == "pathless" && ref == p.research.ID {
		return core.LoadedRecord[core.ResearchWithBody]{Value: core.ResearchWithBody{Research: p.research, Body: p.body}, Source: core.RecordSource{ID: p.research.ID}}, nil
	}
	return p.Store.ReadResearchDocument(ref)
}

func TestDocumentCreatePortableReceiptsAndNoReplay(t *testing.T) {
	for name, kind := range documentCreateKinds {
		for _, outcome := range []string{"pathless", "read-only", "conflict", "unknown", "unconfirmed", "preview", "finalization", "committed cleanup", "committed refusal", "path evidence"} {
			t.Run(name+"/"+outcome, func(t *testing.T) {
				root := seedRepo(t)
				fs := testutil.Must(store.NewFS(root, core.UnrestrictedMutations()))
				port := &documentCreatePort{Store: fs, setID: fs.SourceSetID(), outcome: outcome}
				m := loadedAt(t, root, 120, 40)
				m.svc = core.MustNewService(port)
				m = fillDocumentCreate(openCreationKind(t, m, kind))
				before := testutil.SnapshotTree(t, root)
				m, write := submitDocumentCreate(m)
				m = settleCreate(t, m, write)
				wantCalls := 1
				if kind == entityResearch && outcome == "conflict" {
					wantCalls = 8
				} // bounded core collision loop, not a UI retry
				if port.calls != wantCalls {
					t.Fatalf("calls=%d, want %d", port.calls, wantCalls)
				}
				if kind == entityAudits {
					if port.audit.Area != "Review / qhjkn 世界" || port.audit.Bucket != domain.AuditOpen || !strings.Contains(port.body, "# Audit: Review / qhjkn 世界") {
						t.Fatal("audit policy lost at portable port")
					}
				} else if port.research.Description != "Compare approaches α" || !slices.Equal(port.research.Tags, []string{"planning", "tui"}) || !strings.Contains(port.body, "# Review / qhjkn 世界") {
					t.Fatal("research policy lost at portable port")
				}
				if outcome == "pathless" {
					if m.activeCreation() != nil || m.selectedRef().key != "urn:"+name+":created" || strings.Contains(m.flash, "editor") {
						t.Fatal("pathless landing lost opaque identity or promised local editing")
					}
				} else {
					known := outcome == "read-only" || outcome == "conflict"
					if f := m.activeCreation(); f == nil || f.saving || f.locked == known || creationTitle(m) != "Review / qhjkn 世界" {
						t.Fatal("receipt/refusal chose wrong recovery policy")
					}
					if known {
						return
					}
					m, follow := createKey(m, "enter") // inspect/close, never another create
					m = settleCreate(t, m, follow)
					if port.calls != wantCalls {
						t.Fatal("recovery replayed a create")
					}
					if outcome == "committed cleanup" {
						want := port.audit.ID
						if kind == entityResearch {
							want = port.research.ID
						}
						if m.selectedRef().key != want {
							t.Fatal("cleanup inspection missed exact committed identity")
						}
					}
				}
				if outcome != "committed cleanup" && !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatal("portable receipt/refusal changed local bytes/modes")
				}
			})
		}
	}
}

func TestResearchCreateLeavesCollisionRetryToCore(t *testing.T) {
	root := seedRepo(t)
	fs := testutil.Must(store.NewFS(root, core.UnrestrictedMutations()))
	port := &documentCreatePort{Store: fs, setID: fs.SourceSetID(), outcome: "collision once"}
	m := loadedAt(t, root, 120, 40)
	m.svc = core.MustNewService(port)
	m = fillDocumentCreate(openCreationKind(t, m, entityResearch))
	m, write := submitDocumentCreate(m)
	m = settleCreate(t, m, write)
	if port.calls != 2 || m.activeCreation() != nil || m.selectedRef().key != port.research.ID {
		t.Fatal("single UI action did not preserve core collision handling")
	}
}

func TestDocumentCreateReadOnlyAndSubmitOriginGuards(t *testing.T) {
	for name, kind := range documentCreateKinds {
		for _, refusal := range []string{"read-only", "service changed", "session changed"} {
			t.Run(name+"/"+refusal, func(t *testing.T) {
				root := seedRepo(t)
				m := loadedAt(t, root, 120, 40)
				if refusal == "read-only" {
					m.svc = core.MustNewService(testutil.Must(store.NewFS(root, core.ReadOnlyMutations())))
				}
				m = fillDocumentCreate(openCreationKind(t, m, kind))
				if refusal == "service changed" {
					m.svc = core.MustNewService(testutil.Must(store.NewFS(root, core.UnrestrictedMutations())))
				}
				if refusal == "session changed" {
					m.sessionGen++
				}
				before := testutil.SnapshotTree(t, root)
				m, cmd := submitDocumentCreate(m)
				if refusal != "read-only" && cmd != nil {
					t.Fatal("stale service/session issued a write")
				}
				m = settleCreate(t, m, cmd)
				f := m.activeCreation()
				if f == nil || f.saving || f.err == "" || f.locked != (refusal != "read-only") || creationTitle(m) != "Review / qhjkn 世界" || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatal("refusal lost input, offered stale submission, or changed durable bytes/modes")
				}
			})
		}
	}
}

func TestDocumentCreateActualResultGuards(t *testing.T) {
	for name, kind := range documentCreateKinds {
		t.Run(name, func(t *testing.T) {
			root := seedRepo(t)
			m := fillDocumentCreate(openCreationKind(t, loadedAt(t, root, 120, 40), kind))
			m, cmd := submitDocumentCreate(m)
			result := cmd() // actual typed receipt, after one durable create
			var stale tea.Msg
			switch msg := result.(type) {
			case auditCreatedMsg:
				msg.gen++
				stale = msg
			case researchCreatedMsg:
				msg.gen++
				stale = msg
			default:
				t.Fatalf("wrong result type: %T", result)
			}
			before := testutil.SnapshotTree(t, root)
			next, follow := m.Update(stale)
			m = settleCreate(t, next.(Model), follow)
			if !m.activeCreation().saving || follow != nil {
				t.Fatal("stale typed result altered a pending attempt")
			}
			next, follow = m.Update(result)
			m = settleCreate(t, next.(Model), follow)
			ref := m.selectedRef()
			if m.activeCreation() != nil || ref.empty() {
				t.Fatal("matching actual result failed to land")
			}
			next, follow = m.Update(result)
			m = settleCreate(t, next.(Model), follow)
			if follow != nil || m.selectedRef() != ref || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatal("duplicate result issued work or changed identity/bytes")
			}
		})
	}
}

func TestDocumentCreatePaletteWatcherAndOrigin(t *testing.T) {
	for name, kind := range documentCreateKinds {
		t.Run(name+"/palette and watcher", func(t *testing.T) {
			m := loaded(t, 120, 40)
			m = settleCreate(t, m, m.switchTab(indexOfKind(m.tabs, kind)))
			m.setFocus(focusDetail)
			m, cmd := createKey(m, "ctrl+p")
			m = settleCreate(t, m, cmd)
			found := false
			for i, raw := range m.palette.list.Items() {
				if it := raw.(paletteItem); it.kind == palAction && it.word == "new-"+name {
					m.palette.list.Select(i)
					found = true
					break
				}
			}
			if !found {
				t.Fatal("named palette action missing")
			}
			m, cmd = createKey(m, "enter")
			m = fillDocumentCreate(settleCreate(t, m, cmd))
			m.cur().list.SetFilterText("will not match")
			if kind == entityAudits {
				m.cur().statusView = "closed"
			}
			m, write := submitDocumentCreate(m)
			next, landing := m.Update(write())
			m = next.(Model)
			pending := m.cur().restore
			m = settleCreate(t, m, m.reloadAll())
			m = settleCreate(t, m, landing)
			if pending.empty() || m.selectedRef() != pending {
				t.Fatal("watcher/filter/bucket stole created-record selection")
			}
			m.atlas.opening = true
			m, _ = createKey(m, "c")
			if m.activeCreation() != nil || !strings.Contains(m.flash, "pending workspace") {
				t.Fatal("opened under a pending workspace activation")
			}
		})
		for _, issued := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/origin/issued=%t", name, issued), func(t *testing.T) {
				alpha, beta := seedRepo(t), seedRepo(t)
				m := fillDocumentCreate(openCreationKind(t, loadedAt(t, alpha, 120, 40), kind))
				origin := m.svc
				m.sessionScope = true
				var write tea.Cmd
				if issued {
					m, write = submitDocumentCreate(m)
				}
				incomingFS := testutil.Must(store.NewFS(beta, core.UnrestrictedMutations()))
				incoming := &documentCreatePort{Store: incomingFS, setID: incomingFS.SourceSetID(), outcome: "success"}
				m.activateWorkspace(core.Workspace{Checkout: beta, PlanningRoot: beta, Planning: core.MustNewService(incoming)}, nil, nil)
				before := testutil.SnapshotTree(t, beta)
				reads := incoming.readCalls
				prior := m
				if issued {
					result := write() // execute the actual origin command only after switching
					payload := result
					if scoped, ok := payload.(sessionMsg); ok {
						payload = scoped.msg
					}
					switch msg := payload.(type) {
					case auditCreatedMsg:
						// Same-ID incoming data must not become a landing/reload target.
						a := msg.receipt.Audit
						a.Area = "Incoming impostor"
						if _, err := incomingFS.CreateAudit(a, "# Incoming impostor\n", false); err != nil {
							t.Fatal(err)
						}
						record, err := origin.ShowAudit(a.ID)
						if err != nil || record.Value.Audit.Area != "Review / qhjkn 世界" {
							t.Fatalf("origin audit was not committed: %v", err)
						}
					case researchCreatedMsg:
						r := msg.receipt.Research
						r.Description = "Incoming impostor"
						if _, err := incomingFS.CreateResearch(r, "# Incoming impostor\n", false); err != nil {
							t.Fatal(err)
						}
						record, err := origin.ShowResearch(r.ID)
						if err != nil || record.Value.Research.Description != "Compare approaches α" {
							t.Fatalf("origin research was not committed: %v", err)
						}
					default:
						t.Fatalf("wrong origin payload: %T", payload)
					}
					before = testutil.SnapshotTree(t, beta)
					next, follow := m.Update(result)
					m = settleCreate(t, next.(Model), follow)
					// Duplicate actual delivery and recovery actions must also execute without replay.
					next, follow = m.Update(result)
					m = settleCreate(t, next.(Model), follow)
					committed := m.creation.audit.receipt.Committed
					if kind == entityResearch {
						committed = m.creation.research.receipt.Committed
					}
					if f := m.activeCreation(); !committed || f == nil || f.saving {
						t.Fatal("departed result lost origin receipt")
					}
				}
				if f := m.activeCreation(); f == nil || !f.locked || !f.displaced || creationTitle(m) != "Review / qhjkn 世界" {
					t.Fatal("workspace switch lost/rebound the draft")
				}
				m, follow := createKey(m, "enter")
				m = settleCreate(t, m, follow)
				if incoming.calls != 0 || incoming.readCalls != reads || !maps.Equal(before, testutil.SnapshotTree(t, beta)) || !prior.activeCreation().active {
					t.Fatal("late outcome reloaded/wrote incoming or mutated retained recovery")
				}
			})
		}
	}
}

func TestDocumentCreateNarrowModes(t *testing.T) {
	for name, kind := range documentCreateKinds {
		for _, mode := range []string{"fields", "discard", "pending", "recovery"} {
			t.Run(name+"/"+mode, func(t *testing.T) {
				m := fillDocumentCreate(openCreationKind(t, loaded(t, 120, 40), kind))
				switch mode {
				case "discard":
					m, _ = createKey(m, "esc")
				case "pending":
					m, _ = submitDocumentCreate(m)
				case "recovery":
					m.creation.displace()
				}
				for _, size := range [][2]int{{120, 40}, {46, 18}, {45, 18}, {46, 17}, {12, 5}, {120, 40}} {
					next, _ := m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
					m = next.(Model)
					view := m.View().Content
					if lipgloss.Width(view) > size[0] || lipgloss.Height(view) > size[1] {
						t.Fatal("modal overflow")
					}
					if mode == "fields" && size[0] >= 46 && size[1] >= 18 && !strings.Contains(ansi.Strip(view), "[Create]") {
						t.Fatal("supported size lost submit action")
					}
				}
			})
		}
	}
}

func TestDocumentCreateInRealProgram(t *testing.T) {
	for name, kind := range documentCreateKinds {
		t.Run(name, func(t *testing.T) {
			m := loaded(t, 120, 40)
			m = settleCreate(t, m, m.switchTab(indexOfKind(m.tabs, kind)))
			tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 40))
			wait := func(texts ...string) {
				t.Helper()
				teatest.WaitFor(t, tm.Output(), func(output []byte) bool {
					for _, text := range texts {
						if !bytes.Contains(output, []byte(text)) {
							return false
						}
					}
					return true
				}, teatest.WithDuration(5*time.Second))
			}
			tm.Send(press("c"))
			wait("New " + name)
			tm.Send(tea.PasteMsg{Content: "Runtime document 世界"})
			tm.Send(press("enter"))
			if kind == entityResearch {
				tm.Send(tea.PasteMsg{Content: "A useful comparison"})
				tm.Send(press("enter"))
				tm.Send(tea.PasteMsg{Content: "tui, runtime"})
				tm.Send(press("enter"))
			}
			tm.Send(tea.PasteMsg{Content: "2026-10-01"})
			tm.Send(press("enter")) // date -> Cancel
			tm.Send(press("tab"))   // Cancel -> Create
			tm.Send(press("enter"))
			// Wait for loaded detail, not title text already printed by the form:
			// the success flash can render before async selection has landed.
			heading := "# Runtime document"
			if kind == entityAudits {
				heading = "# Audit: Runtime document"
			}
			wait("created ", heading)
			tm.Send(press("q"))
			final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
			if final.activeCreation() != nil || final.selectedRef().empty() {
				t.Fatal("runtime form did not land")
			}
			if kind == entityAudits {
				r, err := final.svc.ShowAudit(final.selectedRef().key)
				if err != nil || r.Value.Audit.Area != "Runtime document 世界" || r.Value.Audit.Date != "2026-10-01" {
					t.Fatalf("runtime audit metadata: %+v, %v", r, err)
				}
			} else {
				r, err := final.svc.ShowResearch(final.selectedRef().key)
				if err != nil || r.Value.Research.Created != "2026-10-01" || r.Value.Research.Description != "A useful comparison" || !slices.Equal(r.Value.Research.Tags, []string{"tui", "runtime"}) {
					t.Fatalf("runtime research metadata: %+v, %v", r, err)
				}
			}
		})
	}
}
