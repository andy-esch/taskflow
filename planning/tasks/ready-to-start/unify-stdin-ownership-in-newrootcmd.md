---
schema: 1
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: App.In, the prompt gate, the editor, and resolveBody read two different stdin handles; make one injectable source.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, architecture]
created: "2026-06-22"
---
# Unify stdin ownership in NewRootCmd

## Objective

root.go hard-wires app.In = os.Stdin and never SetIn; task.go reads cmd.InOrStdin()
for --body-file -, while the prompt gate and editor read app.In. One process has two
stdin handles, so an embedder/test injecting via cmd.SetIn feeds resolveBody but not
prompts/editor. Pick one owner: an io.Reader param to NewRootCmd (or App.SetIn that
also updates the cobra root) so all paths read the same source.

## Audit reference

planning/audits/open/2026-06-22-code-quality-architecture.md — **M12**. Mostly bites embedders + interactive-prompt tests today; tidies the
DI story's input side. Relates to epic 20 (CLI UX).

## Acceptance criteria

- [ ] One injectable stdin feeds App.In, the prompt gate, the editor, and resolveBody.
- [ ] An in-process test can drive a prompt by injecting stdin.
- [ ] just test + just lint green.

## Implementation plan

**Approach.** Add an explicit `in io.Reader` parameter to `NewRootCmd` and use it as
the single stdin owner: set it on both `app.In` AND the cobra root (`root.SetIn(in)`),
so the two stdin sources collapse to one. This is the audit's recommended fix and fits
the existing DI shape (`NewRootCmd(out, errOut io.Writer)` already injects the
writers; stdin is the missing third). Preferred over an `App.SetIn` method because the
constructor already takes the other two streams — keeping all three together is the
honest "DI via one *cli.App" story.

**Steps.**
1. **Signature (root.go).** Change `func NewRootCmd(out, errOut io.Writer)` →
   `func NewRootCmd(in io.Reader, out, errOut io.Writer)` and build the app with
   `&App{Out: out, ErrOut: errOut, In: in}` instead of `In: os.Stdin` (root.go:60–61).
   Add `root.SetIn(in)` next to the existing `root.SetOut(out)`/`root.SetErr(errOut)`
   (root.go:85–86) so `cmd.InOrStdin()` (used by `resolveBody` in task.go:25) reads the
   same reader as `app.In`.
2. **Call sites (~50).** Update the three production/tool callers —
   `cmd/tskflwctl/main.go:18`, `internal/tools/mangen/main.go:30`,
   `internal/tools/docgen/main.go:26` (all `NewRootCmd(os.Stdout, os.Stderr)`) — to pass
   stdin (`os.Stdin` for main; an empty reader for the doc/man generators). Then sweep
   every in-process test that calls `NewRootCmd(&out, &out)` (~45 sites across
   task_test.go, create_test.go, body_test.go, pipeline_test.go, fill_test.go, etc.) and
   thread a reader — default `strings.NewReader("")` where stdin is unused. This is the
   bulk of the work; it's mechanical and the compiler enforces completeness.
3. **Verify the four consumers now agree.** `app.In` feeds: the prompt gate
   (`isTerminalReader(a.In)` in `setStyle`, root.go:54), the prompter
   (`prompt.NewTTY(a.In, a.ErrOut)`, root.go:55), the editor
   (`cmd.Stdin = app.In` in edit.go:117), and `resolveBody`'s `cmd.InOrStdin()` (now
   also `app.In` via `SetIn`). Confirm no other reader of `os.Stdin` remains
   (`grep os.Stdin internal/cli`).

**Tests.** Add an in-process CLI test (in `internal/cli`) that injects stdin via the
new param and drives a prompt: e.g. `task new "X" --epic e1` on a seeded repo with the
gate forced open, feeding the tags/description prompt from the injected reader, and
assert the created task picks up the piped values. This is currently impossible
(prompts read `app.In`, which tests couldn't set distinctly from cobra's stdin) — its
existence is the M12 acceptance criterion. Keep `body_test.go`'s `--body-file -` stdin
test green (it now reads the same injected reader). Note: the prompt gate also requires
a TTY; for the test, exercise the resolution via the `prompt.Fake` (fake.go) or a
gate-open seam rather than a real TTY, and assert the *plumbing* (one reader reaches
the prompter) — the gate's TTY check is out of scope.

**Risks / gotchas.** (a) Every `NewRootCmd` call site must be updated or it won't
compile — that's the safety net; sweep `internal/cli/*_test.go` and
`cmd/tskflwctl/main.go`. (b) Don't break the agent/pipeline contract: off a TTY the
gate stays closed regardless of the injected reader, so behavior for non-interactive
callers is unchanged — only embedders/tests gain a working injection point. (c)
`isTerminalReader(a.In)` must still detect a real `*os.File` TTY when production passes
`os.Stdin` (a `strings.Reader` is correctly non-TTY). (d) Keep the param order
`(in, out, errOut)` consistent with the io convention; update the doc comment on
`NewRootCmd`.

**Done when.** `NewRootCmd(in, out, errOut)` is the single stdin owner feeding
`App.In`, the gate, the prompter, the editor, and `resolveBody`; an in-process test
drives a prompt by injecting stdin; no stray `os.Stdin` reader remains in
`internal/cli`; and `go build ./...`, `go test ./...`, `golangci-lint run ./...` are
green.
