---
schema: 1
id: 6ga5dr4g73zq
bucket: open
area: go-1-27-modernization
date: "2026-09-14"
---

# Code Quality Audit: Go 1.27 Modernization — 2026-09-14

> Edit findings through `tskflwctl audit finding` so status and resolution
> metadata stay queryable. Never hand-edit a `**Status:**` line.

Routine: `code-quality-audit` · lens `go-1-27-modernization` · Go release comparison: Go 1.25.12 (current `go.mod`) vs Go 1.26 and Go 1.27.

Authoritative reference: <https://go.dev/doc/go1.27> (and <https://go.dev/doc/go1.26>, <https://go.dev/doc/go1.25>).

## Punch list

One line per finding for fast triage. Order: Critical → High → Medium → Low.

- H1. Reflection-based `errors.As` and verbose pointer variables can be modernized with Go 1.26 `errors.AsType` and `new(expr)`  (effort: S · urgency: acute)
- H2. CLI JSON streaming and ordered projection allocate intermediate buffers and escape HTML without Go 1.27 `encoding/json/v2` and `jsontext`  (effort: M · urgency: soon)
- M1. Envelope DTOs use 15+ heap pointer fields for optional structs instead of Go 1.24/1.27 `omitzero`  (effort: S · urgency: soon)
- M2. Store and dangler traversals lack kernel-level sandboxing available via `os.Root` (`os.OpenRoot`)  (effort: M · urgency: soon)
- M3. Manual string slice math and delimiter bounds checks can be simplified with Go 1.27 `strings.CutLast`  (effort: XS · urgency: soon)
- M4. Package-level retry, mutation, and list helpers can be converted to Go 1.27 Generic Methods on `*Service`, `*FS`, and `*App`  (effort: S · urgency: eventually)
- M5. Store concurrency tests burn ~1s of wall-clock time and risk race flakiness without Go 1.25/1.27 `testing/synctest`  (effort: S · urgency: soon)
- L1. Legacy benchmark loops, `WaitGroup` synchronization, and a TUI shutdown watcher leak can be modernized with `b.Loop()`, `wg.Go()`, and `goroutineleak` profiling  (effort: XS · urgency: eventually)
- L2. Transitive and dual dependencies in `go.mod` (dual JSON schema validators, manpage generation) can be pruned against modern stdlib and Cobra tooling  (effort: S · urgency: eventually)

## Files audited

- **Signal (CLI Machine Contract & Error Pipeline)**: `internal/cli/exit.go` (40 lines of error unwrapping), `internal/cli/fserror.go`, `internal/wire/wire.go:288-291` (`EncodeJSON`), and `internal/cli/render/columns.go:283-304` (`marshalOrderedObject`).
- **Signal (Wire DTOs & Schema Definitions)**: `internal/wire/envelopes.go` (1220 lines; 52 envelope structs, 15+ pointer fields for optional structs), `internal/wire/dto.go`.
- **Signal (Store Boundaries & Atomic Replacement)**: `internal/store/fsstore.go`, `internal/store/create.go`, `internal/store/resolve.go`, `internal/store/danglers.go`, and `internal/store/atomic.go`.
- **Adjacency (Concurrency, Locking & Debounce)**: `internal/store/threadmutation_test.go`, `internal/store/lifecyclemutation_test.go`, `internal/store/occ_test.go`, `internal/tui/watch.go`, `internal/tui/atlas.go`, and `internal/tui/read_retry.go`.
- **Adjacency (String & Path Operations)**: `internal/cli/task_dependency_repair.go:124-134`, `internal/domain/slug.go:68-71`, and `internal/cli/listmode.go:230-241`.
- **Adjacency (Generics & Method Ergonomics)**: `internal/core/retry.go:62` (`retryOnConflict`), `internal/store/edit.go:69` (`editFile`), `internal/store/body.go:24` (`writeBody`), and `internal/cli/moves.go:32` (`runMoves`).
- **Random**: `internal/configstore/fs.go:189-214` (configuration bool cloning and pointer creation).

## Commands run

```
go version                                      # go version go1.26.6 darwin/arm64
just build                                      # exit 0 -> bin/tskflwctl
just test                                       # exit 0 (race-enabled suite passes clean)
./bin/tskflwctl audit list --json               # 4 open audits; well below backpressure bar (10)
curl -s https://go.dev/doc/go1.27               # official Go 1.27 release documentation
curl -s https://go.dev/doc/go1.26               # official Go 1.26 release documentation
curl -s https://go.dev/doc/go1.25               # official Go 1.25 release documentation
```

An AST and grep sweep across `internal/` and `cmd/` checked:
- 18 production and 20 test call sites of `errors.As`.
- 15+ pointer fields in `internal/wire/` used exclusively for non-pointer struct omission.
- 8 timeout-based concurrency tests burning ~1s wall-clock latency in `internal/store/`.
- 7 benchmark functions in `internal/`.
- 4 WaitGroup completion groups and 2 entry barriers.

## Findings

### Critical

(none)

### High

#### H1. Reflection-based `errors.As` and verbose pointer variables can be modernized with Go 1.26 `errors.AsType` and `new(expr)`  · **Status:** open

**File:** `internal/cli/exit.go:84-123` | **Component:** cli/exit, cli/fserror, core, configstore
**Effort:** S · **Urgency:** acute
**Class:** language modernization
**Anchored to:** <https://go.dev/doc/go1.26#errors> and <https://go.dev/doc/go1.26#language>

In Go 1.26, the language added `new(expr)` to allocate and populate pointers from arbitrary expressions, and standard library `errors` introduced `errors.AsType[E error](err error) (target E, ok bool)` as a type-safe generic alternative to `errors.As(err, any)`.

Across taskflow's error synthesis and configuration layers, the codebase is forced to use verbose boilerplate:
1. In `internal/cli/exit.go:84-123`, 8 consecutive `var xyzErr *SomeCommandFailure` pointer variables are declared across 40 lines solely to satisfy `errors.As(err, &xyzErr)`. This pollutes outer function scope, incurs reflection indirection, and risks runtime errors if an unaddressed pointer is passed.
2. In `internal/cli/fserror.go:21-39`, `filesystemDetails` pre-declares `var (pathErr *fs.PathError; linkErr *os.LinkError)` before a switch with `errors.As`.
3. In `internal/core/configuration_test.go:29` (`func boolPtr(v bool) *bool { return &v }`) and `internal/cli/pager_test.go:15` (`func boolp(b bool) *bool { return &b }`), custom helper functions are declared to create pointers. In `internal/configstore/fs.go:208-214` (`cloneBool`), a temporary local variable `out := *v; return &out` is required.

**Failing scenario / Before vs After:**
In `internal/cli/exit.go`:
```go
// Current (40 lines of pre-declared uninitialized pointers):
var dependencyErr *dependencyCommandFailure
if errors.As(err, &dependencyErr) {
    details := wire.ToDependencyMutationJSON(dependencyErr.receipt, dependencyErr.workspace)
    payload.Error.DependencyMutation = &details
}
var repairErr *graphRepairCommandFailure
if errors.As(err, &repairErr) {
    details := wire.ToTaskGraphRepairJSON(repairErr.receipt, repairErr.workspace)
    payload.Error.GraphRepair = &details
}
...
```

Modernized with Go 1.26 `errors.AsType` and `new(expr)`:
```go
if dependencyErr, ok := errors.AsType[*dependencyCommandFailure](err); ok {
    payload.Error.DependencyMutation = new(wire.ToDependencyMutationJSON(dependencyErr.receipt, dependencyErr.workspace))
}
if repairErr, ok := errors.AsType[*graphRepairCommandFailure](err); ok {
    payload.Error.GraphRepair = new(wire.ToTaskGraphRepairJSON(repairErr.receipt, repairErr.workspace))
}
if lifecycleErr, ok := errors.AsType[*taskLifecycleCommandFailure](err); ok {
    payload.Error.TaskLifecycle = new(wire.ToTaskLifecycleRecoveryJSON(lifecycleErr.receipt, lifecycleErr.workspace))
}
if renameErr, ok := errors.AsType[*taskRenameCommandFailure](err); ok {
    payload.Error.TaskRename = new(wire.ToTaskRenameRecoveryJSON(renameErr.receipt, renameErr.workspace))
}
if threadErr, ok := errors.AsType[*threadCreationCommandFailure](err); ok {
    payload.Error.ThreadMutation = new(wire.ToThreadMutationJSON(threadErr.receipt, threadErr.path, threadErr.workspace))
}
if threadUpdateErr, ok := errors.AsType[*threadMutationCommandFailure](err); ok {
    payload.Error.ThreadUpdate = new(wire.ToThreadUpdateJSON(threadUpdateErr.receipt, threadUpdateErr.path, threadUpdateErr.workspace))
}
if threadPolicy, ok := errors.AsType[*core.ThreadMutationPolicyError](err); ok {
    payload.Error.ThreadFailure = new(wire.ToThreadMutationFailureJSON(threadPolicy))
}
if threadApplyErr, ok := errors.AsType[*threadApplyCommandFailure](err); ok {
    payload.Error.ThreadApply = new(wire.ToThreadApplyJSON(threadApplyErr.receipt, threadApplyErr.planPath, threadApplyErr.workspace))
}
```

**Recommendation:**
1. Refactor `internal/cli/exit.go:84-123`, `internal/cli/fserror.go:21-39`, `internal/cli/moves.go:41-58`, `internal/cli/task.go:711-714`, `internal/cli/thread.go:127-171`, and `internal/tui/entity.go:349-371` to use `errors.AsType`.
2. Delete `boolPtr` in `internal/core/configuration_test.go` and `boolp` in `internal/cli/pager_test.go`. Replace with `new(true)` / `new(false)`.
3. Simplify `cloneBool` in `internal/configstore/fs.go:208` to `if v == nil { return nil }; return new(*v)`.

---

#### H2. CLI JSON streaming and ordered projection allocate intermediate buffers and escape HTML without Go 1.27 `encoding/json/v2` and `jsontext`  · **Status:** open

**File:** `internal/wire/wire.go:288` · `internal/cli/render/columns.go:283` | **Component:** wire, cli/render
**Effort:** M · **Urgency:** soon
**Class:** language modernization
**Anchored to:** <https://go.dev/doc/go1.27#jsonv2>

In Go 1.27, `encoding/json/v2` and `encoding/json/jsontext` provide zero-allocation streaming, stricter UTF-8/duplicate key validation, and non-escaping HTML defaults.

Taskflow currently suffers from two serialization compromises:
1. **HTML Escaping Overhead & Mangling:** `wire.EncodeJSON` (`internal/wire/wire.go:288-291`) wraps `json.NewEncoder(w).Encode(payload)`. In Go v1, `json.NewEncoder` enables HTML escaping by default (`<`, `>`, `&` become `\u003c`, `\u003e`, `\u0026`). For instance, regexes or format hints in descriptions (e.g., `summary (<=200 chars)` in `internal/wire/dto.go:32` or repair selectors like `<task-or-path>:<field>=<raw-value>` in `task_dependency_repair.go:122`) are emitted as `\u003c...` across CLI JSON streams, inflating token count for AI agents and human JSON consumers.
2. **Manual Buffer Manipulation for Ordered Projections:** To preserve user-selected column order (`-c`) without `encoding/json` alphabetically re-sorting map keys, `internal/cli/render/columns.go:283-304` implements a hand-rolled `marshalOrderedObject` helper that manually manages `{`, `,`, `:`, marshals keys/values separately with `json.Marshal`, and concatenates them in a `bytes.Buffer`.

**Recommendation:**
1. Modernize `wire.EncodeJSON` to use `json/v2.MarshalWrite(w, payload)`:
   ```go
   func EncodeJSON(w io.Writer, payload any) error {
       if err := json.MarshalWrite(w, payload); err != nil {
           return err
       }
       _, err := w.Write([]byte{'\n'}) // json/v2 does not append trailing newline by default
       return err
   }
   ```
   This disables HTML escaping by default, cuts allocations via buffer pooling, and preserves taskflow's trailing newline invariant.
2. Refactor `marshalOrderedObject` in `internal/cli/render/columns.go:283-304` to use `jsontext.Encoder`:
   ```go
   func writeOrderedObject(w io.Writer, fields []orderedField) error {
       enc := jsontext.NewEncoder(w)
       if err := enc.WriteToken(jsontext.BeginObject); err != nil {
           return err
       }
       for _, f := range fields {
           if err := enc.WriteToken(jsontext.String(f.key)); err != nil {
               return err
           }
           if err := json.MarshalEncode(enc, f.value); err != nil {
               return err
           }
       }
       return enc.WriteToken(jsontext.EndObject)
   }
   ```
   This eliminates $O(N)$ intermediate slices per row in table/JSON projections.

---

### Medium

#### M1. Envelope DTOs use 15+ heap pointer fields for optional structs instead of Go 1.24/1.27 `omitzero`  · **Status:** open

**File:** `internal/wire/envelopes.go:54,210,258,1043` · `internal/wire/thread.go:234` | **Component:** wire
**Effort:** S · **Urgency:** soon
**Class:** language modernization
**Anchored to:** <https://go.dev/doc/go1.24#json> and <https://go.dev/doc/go1.27#jsonv2>

Prior to Go 1.24, `encoding/json` `omitempty` never omitted non-pointer struct values (rendering an empty `struct{}` as `{}`). To keep JSON envelopes compact and clean, taskflow defined 15+ optional struct fields as heap pointers:
- `internal/wire/envelopes.go`: `Graph *GraphHealthJSON` (lines 54, 264), `Findings *FindingsRollupJSON` (line 258), `Lifecycle *TaskLifecycleJSON` (line 210), `Summary *SummaryJSON` (line 324), and all failure detail structs on `ErrorItem` (lines 1043–1051: `DependencyMutation`, `GraphRepair`, `TaskLifecycle`, `TaskRename`, `ThreadMutation`, `ThreadUpdate`, `ThreadFailure`, `ThreadApply`, `Filesystem`).
- `internal/wire/dependency_repair.go:55,60`: `Target *TaskGraphRepairEditJSON`, `Problem *GraphProblemJSON`.
- `internal/wire/thread.go:234`: `Scope *ThreadGraphScopeJSON`.

Go 1.24 introduced `omitzero` (fully standardized in Go 1.27 `json/v2`), which omits any field holding its type's zero value or where `IsZero() bool` returns true.

**Recommendation:**
1. Replace pointer fields with direct struct values tagged `json:"...,omitzero"` on all optional single-struct fields where `nil` was used solely for omission.
2. **Caution / Invariant Preservation:** Do NOT change slice fields (such as `Unreadable []domain.FileProblem`) from `omitempty` to `omitzero`. In `omitzero`, an empty initialized slice (`[]T{}`) is NOT omitted (only `nil` is), which would alter the contract for `TasksEnvelope`, `EpicsEnvelope`, and `DoctorEnvelope`. Retain explicit non-omitted `DryRun bool` on all mutation envelopes per SchemaVersion 1.3 invariant (`wire.go:29`).

---

#### M2. Store and dangler traversals lack kernel-level sandboxing available via `os.Root` (`os.OpenRoot`)  · **Status:** open

**File:** `internal/store/fsstore.go:36` · `internal/store/create.go:194,439` · `internal/store/danglers.go:56` | **Component:** store, core
**Effort:** M · **Urgency:** soon
**Class:** security and hardening
**Anchored to:** <https://go.dev/doc/go1.24#os> and <https://go.dev/doc/go1.25#os>

Taskflow currently relies on manual lexical path checks to avoid directory traversal and symlink escapes:
1. `validQueryName` (`internal/store/resolve.go:187`) checks queries for `..` and `/`.
2. `markdownDoc` (`internal/store/resolve.go:23`) asserts `e.Type().IsRegular()` to reject symlinks in directory scans.
3. `DanglingLinks` (`internal/store/danglers.go:56-61`) performs lexical `filepath.Rel(s.root, resolved)` to skip links that escape `s.root`.

**Gaps Identified:**
- In `CreateTask` (`internal/store/create.go:194-195`) and `CreateEpic` (`internal/store/create.go:439-440`), while `t.ID` is validated, `t.Slug` is NOT validated with `validQueryName`. A malformed slug with `../` would be accepted by `filepath.Join(s.tasksDir, stem+".md")` and write outside `tasksDir`.
- In `DanglingLinks`, `filepath.Rel` is purely lexical; if an in-tree symlink points outside the planning root, subsequent `os.Stat(resolved)` follows the symlink outside the repository root.

Go 1.24 introduced `os.Root` (`os.OpenRoot(dir)`), expanded in Go 1.25 with `Root.ReadFile`, `Root.WriteFile`, `Root.Rename`, `Root.Stat`, and `Root.FS()`. `os.Root` enforces directory containment at the OS kernel/runtime level (`openat2` / `RESOLVE_BENEATH` on Linux), making it physically impossible for any open or stat call to escape the planning repository root.

**Recommendation:**
1. Add `rootHandle *os.Root` to `store.FS` opened during `store.NewFS(root)`.
2. Add `validQueryName` validation to `t.Slug` in `CreateTask` and `CreateEpic`.
3. In `DanglingLinks` and `RenameTask`, use `s.rootHandle.FS()` with `io/fs.WalkDir` and `s.rootHandle.Stat(rel)` to guarantee traversal-proof probing.
4. Maintain `root string` alongside `rootHandle *os.Root` to preserve absolute path requirements for external contracts (`$EDITOR` and CLI path commands).

---

#### M3. Manual string slice math and delimiter bounds checks can be simplified with Go 1.27 `strings.CutLast`  · **Status:** open

**File:** `internal/cli/task_dependency_repair.go:126` · `internal/domain/slug.go:68` · `internal/cli/listmode.go:231` | **Component:** cli, domain
**Effort:** XS · **Urgency:** soon
**Class:** simplification
**Anchored to:** <https://go.dev/doc/go1.27#strings>

Go 1.27 introduces `strings.CutLast` and `bytes.CutLast` (`CutLast(s, sep string) (before, after string, found bool)`).

In three distinct places, taskflow uses manual index tracking, slice bounds checking, or redundant split/join cycles:
1. In `internal/cli/task_dependency_repair.go:126-133`:
   ```go
   // Current:
   if hash := strings.LastIndex(value, "#"); hash >= 0 && hash < len(value)-1 {
       if parsed, err := strconv.Atoi(value[hash+1:]); err == nil {
           if parsed < 0 { return core.TaskGraphSourceEdit{}, ... }
           occurrence, value = parsed, value[:hash]
       }
   }
   ```
   Modernized with `strings.CutLast`:
   ```go
   if baseVal, occStr, found := strings.CutLast(value, "#"); found && occStr != "" {
       if parsed, err := strconv.Atoi(occStr); err == nil {
           if parsed < 0 { return core.TaskGraphSourceEdit{}, ... }
           occurrence, value = parsed, baseVal
       }
   }
   ```
   Eliminates manual index arithmetic and off-by-one bounds checks (`hash < len(value)-1`).
2. In `internal/domain/slug.go:68-70`:
   ```go
   // Current:
   if i := strings.LastIndex(text, "-"); i >= len(text)/2 {
       text = text[:i]
   }
   ```
   Modernized:
   ```go
   if before, _, found := strings.CutLast(text, "-"); found && len(before) >= len(text)/2 {
       text = before
   }
   ```
3. In `internal/cli/listmode.go:231-233`:
   `parts := strings.Split(toComplete, ","); prefix := strings.Join(parts[:len(parts)-1], ","); last := parts[len(parts)-1]`
   allocates and rejoins slices even when typing the first column with no comma. `strings.CutLast(toComplete, ",")` performs zero allocations when no comma is present.

**Recommendation:** Replace `LastIndex` and right-to-left slice operations with `strings.CutLast`.

---

#### M4. Package-level retry, mutation, and list helpers can be converted to Go 1.27 Generic Methods on `*Service`, `*FS`, and `*App`  · **Status:** open

**File:** `internal/core/retry.go:62` · `internal/store/edit.go:69` · `internal/store/body.go:24` · `internal/cli/moves.go:32` | **Component:** core, store, cli
**Effort:** S · **Urgency:** eventually
**Class:** language modernization
**Anchored to:** <https://go.dev/doc/go1.27#language>

Prior to Go 1.27, Go method declarations were forbidden from declaring their own type parameters (only receiver types could be generic). Consequently, generic helpers that logically belonged to a type had to be written as package-level functions with artificial receiver arguments and passed-in closures.

Go 1.27 supports generic methods on types: `func (s *Receiver) Method[T any](...)`.

**Key Opportunities:**
1. `retryOnConflict` (`internal/core/retry.go:62`):
   Currently: `func retryOnConflict[T any](s *Service, dryRun bool, fn func() (T, error)) (T, error)`
   Called 9 times across `internal/core/service_task.go`, `service_epic.go`, `service_audit.go`, `service_research.go`, and `finding.go`.
   Modernized: `func (s *Service) retryOnConflict[T any](dryRun bool, fn func() (T, error)) (T, error)`
   Call sites become natural receiver invocations: `s.retryOnConflict(dryRun, func() (...) { ... })`.
2. `editFile` and `writeBody` (`internal/store/edit.go:69`, `internal/store/body.go:24`):
   Currently accept `lock func() (func(), error)` as a parameter, forcing `EditTask`, `EditAudit`, `EditResearch`, `AppendTaskBody` to pass `s.checkedWriteLock` as a heap-allocated closure.
   Modernized as methods on `(s *FS) editFile[T any](...)` and `(s *FS) writeBody[T any](...)`, calling `s.checkedWriteLock()` directly on the receiver and removing the closure argument.
3. `runMoves` (`internal/cli/moves.go:32`):
   Currently: `func runMoves[T any](app *App, ...)` -> Modernized: `func (app *App) runMoves[T any](...)`.

**Recommendation:** Convert `retryOnConflict`, `editFile`, `writeBody`, and `runMoves` to receiver methods on their respective parent structs.

---

#### M5. Store concurrency tests burn ~1s of wall-clock time and risk race flakiness without Go 1.25/1.27 `testing/synctest`  · **Status:** open

**File:** `internal/store/threadmutation_test.go:302` · `internal/store/lifecyclemutation_test.go:293` · `internal/store/threadcreation_test.go:266` | **Component:** store/test
**Effort:** S · **Urgency:** soon
**Class:** test quality & performance
**Anchored to:** <https://go.dev/doc/go1.25#testing> and <https://go.dev/doc/go1.27#testingsynctest>

`testing/synctest` (graduated in Go 1.25, enhanced with `synctest.Sleep` in Go 1.27) provides virtualized time bubbles: time advances instantaneously when all goroutines in the bubble are durably blocked, and `synctest.Wait()` blocks until all goroutines have settled.

Across `internal/store/` tests, multiple test cases assert negative concurrency guarantees ("Goroutine B cannot proceed while Goroutine A holds the write lock") by sleeping real wall-clock time:
- `threadmutation_test.go:302, 351, 404`: `case <-time.After(100 * time.Millisecond)` (300ms total)
- `threadcreation_test.go:266`: `case <-time.After(100 * time.Millisecond)` (100ms)
- `lifecyclemutation_test.go:293, 356`: `case <-time.After(75 * time.Millisecond)` (150ms)
- `threadapply_test.go:493`: `case <-time.After(75 * time.Millisecond)` (75ms)
- `create_test.go:104`: `case <-time.After(50 * time.Millisecond)` (50ms)
- `occ_test.go:330`: `time.Sleep(time.Millisecond)` in tight retry loops across 8 goroutines.

**Impact:**
- Over 800ms–1.5s of real clock time is wasted on every test run.
- Under heavy CI container load, a 50–100ms timeout can fire before a goroutine is scheduled by the OS, causing intermittent false passes or race flakiness.

**Recommendation:**
Wrap negative concurrency assertions in `synctest.Test(t, func(t *testing.T) { ... })` and call `synctest.Wait()`. When Goroutine B blocks on `repositoryGuard.write.Lock()`, `synctest.Wait()` returns immediately, and a `select { case <-done: t.Fatal(...) default: }` confirms it is blocked without waiting real time.

---

### Low

#### L1. Legacy benchmark loops, `WaitGroup` synchronization, and a TUI shutdown watcher leak can be modernized with `b.Loop()`, `wg.Go()`, and `goroutineleak` profiling  · **Status:** open

**File:** `internal/store/graphrepair_test.go:309` · `internal/id/id_test.go:158` · `internal/tui/tui.go:56` | **Component:** store, tui, id
**Effort:** XS · **Urgency:** eventually
**Class:** hygiene & testing
**Anchored to:** <https://go.dev/doc/go1.24#testing>, <https://go.dev/doc/go1.25#sync>, <https://go.dev/doc/go1.26#goroutineleak-profiles>

1. **Benchmark Loops:** Go 1.24 introduced `(*testing.B).Loop()`, which acts as an optimization barrier preventing loop dead-code elimination. `internal/store/graphrepair_test.go:309-334` and `internal/store/threadapply_benchmark_test.go:52-76` still use `for iteration := 0; iteration < b.N; iteration++`. Convert to `for b.Loop()`. Remove redundant `b.ResetTimer()` preceding `for b.Loop()` in `internal/tui/thread_projection_test.go:3270, 3316`.
2. **`sync.WaitGroup.Go`:** Go 1.25 added `(*sync.WaitGroup).Go(fn func())`. Tests in `internal/id/id_test.go:158-169`, `internal/config/preference_test.go:60-76`, `internal/userconfig/spaces_test.go:314-331`, and `internal/store/occ_test.go:314-335` can replace `wg.Add(1); go func() { defer wg.Done(); ... }()` with `wg.Go(...)`. *(Caution: Do NOT replace readiness/rendezvous wait groups in `graphmutation_test.go:154` or `threadcreation_test.go:169` where `ready.Done()` is called before waiting on `<-start`, which would deadlock with `wg.Go`).*
3. **TUI Shutdown Watcher Leak Window:** In `internal/tui/atlas.go:175`, `openWorkspace` runs asynchronously in a `tea.Cmd`. If a user quickly presses `q` while a workspace is opening, `tui.go:56-62` closes the *current* watcher and exits. The in-flight `openWorkspace` command finishes after program termination, instantiating an orphaned `fsnotify.Watcher` and goroutine. Detected via Go 1.26/1.27 `goroutineleak` profile analysis.

**Recommendation:** Update benchmarks to `b.Loop()`, adopt `wg.Go()`, and track in-flight watchers during TUI shutdown.

---

#### L2. Transitive and dual dependencies in `go.mod` (dual JSON schema validators, manpage generation) can be pruned against modern stdlib and Cobra tooling  · **Status:** open

**File:** `go.mod:17-20` · `internal/tools/mangen/main.go:20` | **Component:** build, dependencies
**Effort:** S · **Urgency:** eventually
**Class:** dependency pruning
**Anchored to:** <https://go.dev/doc/go1.27#go-mod-tidy>

An inspection of `go.mod` reveals redundancy that can be streamlined:
1. **Dual JSON Schema Libraries:** `invopop/jsonschema v0.14.0` is used to *generate* Draft 2020-12 schemas from Go structs (`internal/wire/envelopes.go:7`), while `santhosh-tekuri/jsonschema/v6 v6.0.3` is used *only in tests* (`internal/wire/envelopes_test.go:9`) to validate golden outputs. This dual dependency pulls in heavy indirect libraries (`buger/jsonparser`, `pb33f/ordered-map/v2`, `bahlo/generic-list-go`).
2. **Manpage Dependencies:** `github.com/muesli/mango-cobra` and `github.com/muesli/roff` (`internal/tools/mangen/main.go:20-21`) can be replaced with `github.com/spf13/cobra/doc.GenManTree`, which is already part of the existing Cobra dependency.
3. **Go 1.27 Stdlib `uuid` vs Crockford Base32:** Go 1.27 adds a standard `uuid` package (`uuid.NewV7()`). **Verdict:** Keep taskflow's custom 12-char Crockford base32 IDs (`internal/id/`) for all file-based planning entities (`<id>-<slug>.md`); reserve standard library `uuid` strictly for CLI session tracing or external synchronization adapters.

**Recommendation:** Prune `muesli/mango-cobra` in favor of `cobra/doc`, evaluate consolidating JSON schema test validation, and run Go 1.27 `go mod tidy` to merge require blocks.

---

## What audited clean

- **Hexagonal Architecture Discipline**: No standard library `os` or `filepath` imports were found in `internal/core/`; Cobra remains strictly isolated in `internal/cli/`. Primary and secondary adapter boundaries remain fully respected.
- **Go 1.21+ Standard Library Adoption**: Slices and maps are clean. Taskflow has already adopted standard library `slices.Contains`, `slices.Equal`, `maps.Clone`, and `cmp.Or`, avoiding external utility dependencies.
- **Envelope Version Invariant**: All 52 JSON envelope structures in `internal/wire/` declare `SchemaVersion string` as their leading field.
- **Atomic File Writing Discipline**: All mutation and write verbs route through `store/atomic.go` (`writeFileAtomic` / `createFileAtomicWithMode`); no raw un-synced `os.WriteFile` calls exist in production write paths.

## External research

- **Go 1.27 Release Notes**: <https://go.dev/doc/go1.27> (Generic methods, struct literal selectors, `encoding/json/v2`, `strings.CutLast`, standard `uuid`, `testing/synctest.Sleep`, size-specialized malloc).
- **Go 1.26 Release Notes**: <https://go.dev/doc/go1.26> (`new(expr)`, `errors.AsType`, `reflect` iterators, Green Tea GC default, chunked `io.ReadAll`, `goroutineleak` profiling).
- **Go 1.25 Release Notes**: <https://go.dev/doc/go1.25> (`sync.WaitGroup.Go`, `testing/synctest` graduation, `FlightRecorder`).
- **Go 1.24 Release Notes**: <https://go.dev/doc/go1.24> (`os.Root` directory containment, `omitzero` struct tag, `testing.B.Loop`).

## Candidate tasks (human to triage)

- ⏳ H1: `tskflwctl task new "Adopt errors.AsType and new(expr) across CLI error mapping and config" --epic 21 --tags modernization,go1.26 --tier 2 --priority high --description "Replace reflection errors.As and temporary pointer variables with Go 1.26 errors.AsType and new(expr)"`
- ⏳ H2: `tskflwctl task new "Adopt encoding/json/v2 and jsontext for CLI streaming and ordered projections" --epic 21 --tags modernization,go1.27 --tier 2 --priority normal --description "Migrate wire.EncodeJSON to json/v2.MarshalWrite and columns.go to jsontext.Encoder"`
- ⏳ M1: `tskflwctl task new "Replace optional struct pointer fields with omitzero in wire envelopes" --epic 21 --tags modernization,wire --tier 2 --priority normal --description "Eliminate 15+ heap pointer fields in internal/wire DTOs using omitzero"`
- ⏳ M2: `tskflwctl task new "Sandbox planning workspace file operations using os.Root" --epic 21 --tags store,security --tier 2 --priority normal --description "Enforce kernel-level directory containment in store.FS using os.OpenRoot and validate create slugs"`
- ⏳ M3: `tskflwctl task new "Simplify suffix and delimiter parsing with strings.CutLast" --epic 21 --tags simplification,go1.27 --tier 3 --priority low --description "Replace LastIndex index arithmetic in repair selectors, slugify, and completer with strings.CutLast"`
- ⏳ M4: `tskflwctl task new "Convert core and store generic helpers to Go 1.27 generic methods" --epic 21 --tags refactor,go1.27 --tier 3 --priority low --description "Convert retryOnConflict, editFile, writeBody, and runMoves into idiomatic receiver methods"`
- ⏳ M5: `tskflwctl task new "Modernize store concurrency tests with testing/synctest" --epic 21 --tags testing,concurrency --tier 2 --priority normal --description "Eliminate 1s of wall-clock test latency and timeout flakiness using synctest.Test and synctest.Wait"`
- ⏳ L1: `tskflwctl task new "Modernize benchmark loops to b.Loop() and guard TUI shutdown watcher leak" --epic 21 --tags testing,tui --tier 3 --priority low --description "Convert benchmarks to b.Loop(), adopt WaitGroup.Go, and track in-flight watchers on TUI exit"`
- ⏳ L2: `tskflwctl task new "Prune redundant dependencies in go.mod" --epic 21 --tags build,dependencies --tier 3 --priority low --description "Replace mango-cobra with cobra/doc and evaluate consolidating JSON schema test validation"`
