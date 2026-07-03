// Package wire is the machine JSON contract for tskflwctl — the --json
// envelopes, the per-entity DTOs + their mappers, the SchemaVersion, and the
// reflected JSON Schema. It is a neutral leaf: it depends only on core + domain
// and on no presentation (no cobra, no lipgloss), so EVERY adapter that needs the
// same wire format imports it — the CLI's render package wraps the value
// constructors here in io.Writer emit funcs, and a future web adapter
// (`tskflwctl serve`) can obtain the same envelope value to embed in an HTTP
// response. A machine wire contract is an API, not presentation, so it lives
// here rather than inside a primary adapter.
//
// The split: every `ToXEnvelope(...)` constructor returns a VALUE (so a web
// handler can wrap it), and render's `XJSON(w, …)` emit funcs build that value
// then encode it. The human renderers (`*Human`, Style, lipgloss) stay in render.
package wire

import (
	"encoding/json"
	"io"
)

// SchemaVersion is the semver of the --json payloads — ONE version for the
// whole CLI output schema, not per envelope (decided 2026-06-12). Adding a
// field bumps the minor; renaming/removing bumps the major. Key naming rule:
// JSON keys match the frontmatter keys exactly (`created`, `updated_at`).
// 1.1: every CLI-settable field round-trips (effort, autonomy_level), and the
// misfiled signal (previously human-output-only ⚠) is machine-readable.
// 1.2: mutation envelopes carry dry_run:true under --dry-run previews.
// 1.3: dry_run is always present on mutation envelopes (was omitted when false);
// the fix report carries `unreadable` (files it couldn't repair).
// 1.4: `schema` envelopes (the tool's self-description contract + per-kind
// authoring guidance) added.
// 1.5: the create envelope carries `status` (task status / epic status / audit
// bucket); its `path` is now relative to the planning root in both human and
// JSON modes (was absolute in JSON).
// 1.6: the `findings` envelope (audit finding-level query) added.
// 1.7: the `task_mutation` envelope (task set/append/set --body) added — it
// carries dry_run and the resulting body, which `task_show` (a read) does not.
// 1.8: the `init` envelope carries `mode` (scaffold|pointer) and `planning_repo`
// (set in pointer mode), for `init --planning-repo`.
// 1.9: the `doctor` envelope (planning_repo <-> tracked_repos linkback audit) added.
// 1.10: the `init` envelope carries `linked_back` (pointer-mode auto-link-back
// path) and `tracked` (scaffold-mode --track entries).
// 1.11: epic rollups exclude deprecated (withdrawn) tasks from total/done; the
// epic payload carries a separate `deprecated` count.
// 1.12: the `status` summary envelope carries `open_audits` — open-bucket audits
// (the actionable subset) with the same finding rollup `audit list` reports;
// omitted when there are none.
// 1.13: every audit payload carries the finding-disposition tally the segmented
// progress bar bands by — `in_progress_findings`, `done_findings` (fixed/landed),
// `dropped_findings` (deferred/superseded/wontfix) — alongside `open_findings`.
// 1.14: the `epic_mutation` envelope (`epic set`) added — the epic counterpart to
// `task_mutation`; it carries dry_run + the reloaded epic (field-only, no body).
// 1.15: the schema contract carries `epic_fields` — the epic frontmatter registry
// (sorted known epic field names), the epic counterpart to `task_fields`, so an
// agent can discover the epic field set without parsing prose. The `fix` envelope
// carries `remaining` — the lint findings `--fix` could NOT repair (report-only
// epics, unfixable task issues), so a --json consumer learns the residual breakage
// without re-running plain lint.
// 1.16: task payloads carry `revisit_at` — the optional snooze-until date set by
// `task defer`; the `status` summary envelope carries `revisit_due` (the
// count of deferred tasks whose revisit_at has arrived) alongside `misfiled`; and
// the move report (`task defer --json`) carries `revisit_at` per item so a preview
// and the real run both confirm the snooze.
// 1.17: the `status` summary envelope carries `findings` — the actionable audit
// findings (open/in-progress) aggregated `by_urgency` and `by_component` with the
// `acute` ones listed — and each open audit carries `ready_to_close` (true when it
// has no open/in-progress findings left).
// 1.18: epic payloads carry `open` (not-yet-done tasks = total − done) and
// `liveness` — the derived activity band (working | fresh | dormant) computed from
// the rollup, not stored — so a consumer can foreground live domain buckets and
// recede drained ones without re-deriving the rule.
// 1.19: the `status` summary envelope carries `bad_epic_status` — the count of
// epics whose status is outside the canonical vocabulary (a fixable data problem;
// these epics are flagged, not dropped), mirroring `misfiled` for tasks.
// 1.20: the `audit_mutation` envelope (`audit append`) added — the audit counterpart
// to `task_mutation`; it carries dry_run + the reloaded audit + the resulting body.
// 1.21: the schema contract carries `finding_statuses` — the legal audit
// finding-status vocabulary (open · in-progress · fixed · landed · deferred ·
// superseded · wontfix), so an agent writing a finding discovers the status set
// without parsing prose, the audit counterpart to `statuses`/`audit_buckets`.
// 1.22: epic and audit payloads carry `updated_at` — the entity's own last-edited
// date, stamped by the tool on every content write (set/edit/append, and epic
// status moves) the way tasks already are. For epics it is distinct from the
// derived task-activity date; for audits it advances on edits while `date` stays
// the immutable slug. A pure relocation (audit bucket move) does not change it.
// 1.23: the `board` envelope (`board --json`) added — the active-work view, tasks
// grouped by their active status (next-up · ready-to-start · in-progress), each
// row the same TaskJSON as `task list`.
// 1.24: task and audit payloads carry `id` — the stable immutable key minted on
// create (survives slug/status changes), the task/audit counterpart to an epic's
// `id`; omitted on entities created before id assignment (pre-migration).
// 1.25: task `status` is now AUTHORITATIVE from frontmatter, not the directory
// (ADR-0003 Phase A). `misfiled`/`declared_status` inverted meaning: `misfiled` = the
// file's directory disagrees with its (authoritative) frontmatter status, and
// `declared_status` now carries the stale mirror DIRECTORY the file sits in (was: the
// frontmatter's claimed status). A file whose directory lags is repaired by `lint
// --fix` MOVING it to match, not by rewriting the status.
const SchemaVersion = "1.25"

// EncodeJSON writes the payload as compact (un-indented) JSON with a single
// trailing newline. Machine output: pretty-printing is pure token cost for a
// consumer that parses it. Off-tree consumers pipe through `jq .` to read it.
// Exported so render's emit funcs (and any other adapter) encode envelopes
// identically.
func EncodeJSON(w io.Writer, payload any) error {
	enc := json.NewEncoder(w)
	return enc.Encode(payload)
}
