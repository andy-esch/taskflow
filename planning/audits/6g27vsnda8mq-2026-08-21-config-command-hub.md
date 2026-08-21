---
schema: 1
id: 6g27vsnda8mq
bucket: closed
area: config-command-hub
date: "2026-08-21"
---
# Audit: config-command-hub — 2026-08-21

Scope: Comprehensive, adversarial review of the configuration-lifecycle consolidation introduced since commit `2a8dfe4`. Covers `internal/core/configuration.go`, `internal/configstore/fs.go`, `internal/configui/editor.go`, `internal/tomledit/table.go`, `internal/config/migrate.go`, `internal/config/preference.go`, `internal/config/lock_unix.go`, `internal/userconfig/preference.go`, `internal/userconfig/lock_unix.go`, `internal/cli/config.go`, `internal/cli/init.go`, `internal/cli/doctor.go`, `internal/cli/ui.go`, `internal/tui/dashboard.go`, `internal/tui/command_dispatch.go`, `internal/tui/overlay.go`, `internal/tui/model.go`, wire envelopes, and documentation.

## Findings

#### M1. Atomic configuration writes replaced symlinks and widened existing permissions  · **Status:** fixed 2026-08-21

**Severity:** Medium · **Files:** `internal/config/config.go`, `internal/userconfig/paths.go`

The initial review reported no actionable defects, but its non-destructive-write claim
did not exercise the actual rename destination. The new repository migration/preference
paths and user-preference path created a temporary file beside the configured pathname,
set it to `0644`, and renamed it over that pathname. When `.tskflwctl.toml` or
`config.toml` was a symlink into a dotfiles checkout, the rename replaced the symlink and
left its target unchanged. A regular existing file with a more restrictive mode such as
`0600` was silently broadened to `0644`.

*Resolution:* both atomic-write boundaries now resolve a symlink to its final target,
create the temporary file beside that target, and carry forward the target's existing
permission bits. A broken symlink errors instead of being overwritten. Regression tests
exercise both migration and repository-preference writes through a symlink, user
preference writes through a symlink, target-content changes, link survival, and `0600`
mode preservation. The registry writer now delegates the same policy to the shared
user-config atomic-write boundary.

## Adversarial Scenarios Examined

1. **Architecture & Boundary Integrity:**
   - Verified that `internal/core` owns `ConfigurationStore` and `ConfigurationService` without importing Cobra, Bubble Tea, filesystem packages (`os`, `filepath`), or design presentation packages.
   - Verified that theme validation vocabulary is injected via `WithConfigurationThemes(design.Names())` at the CLI composition root, keeping `core` pure.
   - Verified that `config` and `userconfig` remain strictly decoupled (enforced by `depguard` in `.golangci.yml`).

2. **Bootstrap vs. Maintenance (`init` vs. `config`):**
   - Verified bare `init` on existing direct repositories and pointer repositories: reports current topology, never silently alters files, and points to `config migrate` when migrations are available.
   - Verified conflicting flags (`--planning-repo` + `--taskflow-root`, `--planning-repo` + `--track`, `--no-link-back` without `--planning-repo`, escaping `taskflow-root`) are rejected with exit code 11 (`domain.ErrValidation`).
   - Verified `doctor` compatibility alias is hidden from root help but produces byte-identical human/JSON output and matching exit codes to canonical `config doctor`.

3. **Migration Correctness & Concurrency Durability:**
   - Tested direct repository missing `id` (backfills durable ID) and legacy pointer missing `planning_repo_id` (resolves target ID and backfills `planning_repo_id`).
   - Verified pointer migration against legacy target without ID fails with actionable error instructing user to migrate target first.
   - Verified idempotency: repeated migrations on current configs are byte-for-byte no-ops.
   - Verified non-destructive surgical edits: comments, custom unknown TOML tables/keys, and ordering survive migration and preference edits intact.
   - Verified dry-run and apply share `core.ConfigurationMigration` and `wire.ConfigMigrationEnvelope`.
   - Verified concurrency: repository writes take directory `flock` in `internal/config/lock_unix.go`, re-read under the lock, and write atomically via temp file + rename (`writeFileAtomic`). User preferences share directory `flock` with `spaces.toml`.
   - Follow-up verification found and fixed M1: atomic writes now preserve symlink-backed configs and existing permission bits.

4. **Precedence & Projection:**
   - Tested all precedence chains:
     - Theme: `--theme` > `TSKFLW_THEME` > repo `[theme].name` > user `[theme].name` > default (`neon`).
     - Pager enabled: `--no-pager`/`--paginate` > repo `[pager].enabled` > user `[pager].enabled` > default (`true`).
     - Pager command: `TSKFLW_PAGER` > repo `[pager].command` > user `[pager].command` > `PAGER` > default (`less -FRX`).
   - Verified field-by-field merging: unset/nil values defer down the chain without overriding sibling fields.
   - Verified provenance tracking (`flag`, `environment`, `repository`, `user`, `default`) in both human and JSON outputs.
   - Verified malformed user configuration degrades with a warning, while malformed repository configuration fails validation to prevent tree forking.

5. **Machine Contracts & Goldens:**
   - Verified JSON envelopes: `ConfigEnvelope` and `ConfigMigrationEnvelope` under schema version `1.40`.
   - Verified all golden snapshots and JSON Schema match the implementation.

6. **Interactive Editor & TUI Integration:**
   - Verified `config edit` Bubble Tea adapter (`internal/configui`): async load/save, live theme preview on cursor move, scope toggling (`s`), tri-state pager toggle (`u` to unset/inherit), and pager command validation (rejects blank).
   - Verified non-TTY, `--no-input`, and `--dry-run` rejections with clear validation errors.
   - Verified TUI integration in `internal/tui`: accessible via Overview `workspace` row, `:config`, and `ctrl+p` command palette.
   - Verified capability safety: when `ConfigurationService` is not injected (e.g. embedded/read-only), Overview omits the row and `:config` flashes an informative message.

## Candidate tasks

<!-- Mirror each finding: ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do -->

None. M1 was fixed inline and regression-tested; no deferred remediation remains.
