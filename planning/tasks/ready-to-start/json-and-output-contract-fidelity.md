---
status: ready-to-start
epic: 17-pm-go-cli
description: JSON drops effort/autonomy/misfiled and epic fields, init ignores --json, cobra help bypasses injected writers, double error prints, stream drift
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [go, cli, json, agents]
created: "2026-06-12"
---
# `--json` and output contract fidelity

> ⚠️ **Externally proposed — filed from the 2026-06-12 review**
> ([[2026-06-12-critical-code-review-multi-lens]], findings M7/M8/M10/M11 +
> output lows). Theme: agents are the stated `--json` audience, and the
> machine output is currently *less* informative than the human output.

## Objective

1. **M7 — JSON drops fields the CLI itself writes.** `taskJSON`
   (`internal/cli/render/render.go:21-31`) omits `effort` and
   `autonomy_level` (both settable via flags — write-only fields) and has no
   `misfiled`/`declared_status` even though human output renders `⚠`.
   `epicJSON`/`epicMetaJSON` similarly drop `priority`, `created`, `tags`.
   Round-trip all frontmatter fields + a misfiled flag (minor schema bump
   per the stated policy). While in there: resolve the `"created"` vs
   `"updated_at"` key asymmetry, and decide whether one global
   `SchemaVersion` for ~10 payload shapes is intentional — before 1.x
   consumers exist.
2. **M10 — `init` is the one command that ignores `--json`.**
   (`internal/cli/init.go:26-45`). Add an `InitJSON` envelope (created
   paths + root).
3. **M8 — Cobra's own output bypasses the injected writers.**
   `NewRootCmd` (`internal/cli/root.go:45-81`) never calls
   `root.SetOut/SetErr`, so help text, usage errors, and completion scripts
   go to `os.Stdout/Stderr` — contradicting the package's "all output flows
   through the injected writers" header; tests already patch around it.
4. **Output lows:** failed transitions print the error twice
   (`moves.go:13-23` + `main.go`); load-problem diagnostics go to stderr in
   list commands but stdout in `lint` — pick one stream.
5. **M11 — Completion gaps:** `task move` offers task slugs for the
   `<status>` position (`task.go:193`); the existing flag completers are
   registered only on `task new --epic`, not `task set --epic` /
   `task list --epic/--status/--tag`.

## Acceptance criteria

- [ ] Every field settable via the CLI is readable back via `--json`;
      misfiled state visible to JSON consumers.
- [ ] `init --json` emits a versioned envelope.
- [ ] Help/usage output flows through the injected writers (de-dupe the
      per-test patches).
- [ ] One transition failure → one error print; diagnostics stream is
      consistent and documented.
- [ ] `task move <slug> <tab>` completes statuses.

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/cli/render/render.go`, `internal/cli/root.go`,
  `internal/cli/init.go`, `internal/cli/moves.go`,
  `internal/cli/completion.go`, `internal/cli/task.go`.