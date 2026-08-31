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
// progress bar bands by — `in_progress_findings`, `done_findings` (fixed/tracked),
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
// finding-status vocabulary (open · in-progress · fixed · tracked · deferred ·
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
// (ADR-0003 Phase A). `misfiled`/`declared_status` inverted meaning to surface
// directory drift.
// 1.26: retired the task `misfiled`/`declared_status` fields and the `status` summary
// `misfiled` count — the flat, id-led layout (ADR-0003 §4) removes the directory mirror
// entirely, so a task/audit can never be misfiled (status/bucket live only in frontmatter).
// 1.27: the `task_info` envelope (`task info` — token-cheap metadata read: path +
// triage fields + acceptance-criteria tally `ac:{checked,total}`, no body) and the
// `path` envelope (`task path --json` — the resolved absolute file path) added.
// 1.28: the `audit_info` envelope (`audit info` — token-cheap audit metadata: path +
// bucket + finding disposition tally `findings:{total,open,in_progress,done,dropped}`,
// no body; the audit counterpart to `task_info`) added. The `path` envelope now also
// backs `epic path` / `audit path` (unchanged shape).
// 1.29: the `acceptance` envelope (`task ac --list` — a task's acceptance criteria,
// each `{index, checked, text}`, the list an agent flips by index) added. A flip
// (`task ac --check/--uncheck`) returns the existing `task_mutation` envelope.
// 1.30: `research` added as a document kind (epic 28) — it appears in the schema
// contract's `kinds` and `template list`, and brings the `research_list` /
// `research_show` envelopes. The shape is deliberately thin, and the ABSENCES are
// contract: a research doc has no `status`/`bucket` (it has no lifecycle — a later doc
// supersedes an earlier one; a decision needing a lifecycle is an ADR) and no
// `epic`/`tasks` (provenance stays body links, so there is no rollup to consume). `id`
// is minted from `created`, so ordering by id is ordering by authorship date.
// 1.31: every MUTATION envelope (task_mutation, epic_mutation, moves, created,
// audit_mutation, fix) gained a `workspace` object — {planning_root, config_path,
// source: pointer|config|discovered} — and a new `workspace` envelope backs the
// `workspace` command. Additive: no existing field changed. It exists because
// external-planning routing (epic 23) means the directory you run in is not
// necessarily the tree you write to, so a receipt has to say which one it was
// (audit 2026-07-24-ai-agent-cli-ergonomics, H1).
// 1.32: `created.id` now carries the STABLE id and `created.slug` is added alongside.
// Previously `created.id` carried the mutable SLUG — a defect, not a deliberate
// contract: every other DTO already distinguished the two, so `new --json` was the one
// command where an agent captured the wrong handle, on exactly the command where
// capturing the durable one matters most (audit 2026-07-24-ai-agent-cli-ergonomics, H2).
// CONSUMERS: anything that stored `created.id` as a reference held a slug and must
// re-read. Epic ids are unchanged — an epic's identity has always been its NN-slug.
// 1.33: the `research_mutation` envelope (`research set` / `research append`) added —
// the research counterpart to `task_mutation`/`audit_mutation`, carrying dry_run, the
// reloaded doc, and the resulting body for a body write. Research gained the two faces
// of mutation the other entities have: field-level `set` (agent) and whole-file `edit`
// (human), plus `append`. `created` is deliberately NOT settable — the stable id is
// minted from it, so changing one would desync the pair. Like every other mutation
// envelope since 1.31, it carries a `workspace` object naming the planning tree the
// receipt describes.
// 1.34: the schema contract carries `research_fields` — the research frontmatter keys
// and their YAML types, mirroring `task_fields`. Added for the same reason `epic_fields`
// was added at 1.15 when `epic set` landed: `research set` gates unknown keys on a field
// registry, so without this an agent had to trigger an error and parse prose to learn the
// set. Note RECOGNIZED is wider than WRITABLE — only description and tags can be set; the
// per-kind conventions in `schema research` name the protected ones.
// 1.35: `workspace.repo_id` added — the planning repo's durable identity, minted into its
// committed config by `init` and verified by a pointer that records `planning_repo_id`.
// Additive, and omitted for a repo that predates the mint. It names the REPO; planning_root
// names the CHECKOUT, and every worktree of a repo shares one id — so the two are paired,
// never substituted.
// 1.36: `workspace.branch` and `workspace.checkout` ("base"|"worktree") added. RepoID
// names the repo and every worktree of it shares one, so these are what tell two working
// trees apart — their directory names are often near-identical. Additive, and both omitted
// when the planning tree is not in a git repo at all.
// 1.37: the `spaces` and `space_mutation` envelopes added — the home-scoped registry of
// known planning repos (`space list|add|forget`). An entry carries BOTH identities: `id`
// is the local label that addresses a checkout, `verify_id` the target repo's durable id,
// which every worktree of that repo shares and so can never be the address. The registry
// is advisory: it changes nothing about how a cwd resolves.
// 1.38: registered-space health became one typed projection shared by `space list` and
// `doctor`: empty/mismatch states plus remedies on SpaceEntry, and a registry section on
// DoctorEnvelope.
// 1.39: registered paths expose their derived direct/pointer role and logical planning
// identity, so one planning space can have multiple machine-readable entry points.
// 1.40: the `config` and `config_migration` envelopes added — resolved repository/user
// scopes, per-value provenance, pending migrations, and deterministic migration receipts.
// 1.41: `workspace.space` added — the machine-local registry label that explicitly
// selected an invocation through `--space` / `TSKFLW_SPACE`. Additive and omitted for
// ordinary -C/cwd discovery. `space` names an entry point; `repo_id` remains the durable
// planning identity shared by all of that repo's checkouts.
// 1.42: the `status_all` envelope (`status --all --json`) added — one summary per
// logical planning identity, its registered entry-point diagnoses, the healthy entry
// point selected for reading, and a combined space-badged in-progress working set. The
// envelope owns one top-level schema_version; nested summaries reuse the versionless
// SummaryJSON payload rather than pretending to be independent envelopes.
// 1.56: guarded existing-Thread membership and lifecycle mutations add atomic
// member outcome receipts, before/after projections, cancel/complete/reopen
// semantics, typed policy failures, and committed recovery. Task lifecycle
// receipts now name every Thread projection changed by the transition.
// 1.55: eligibility now admits both queued (`next-up`) and candidate
// (`ready-to-start`) work when the authoritative graph is healthy and the gate
// is clear. Thread frontier and task list --unblocked share that derivation;
// lifecycle refusal payloads allow dependency override only for blocked gates.
// 1.54: Thread documents add list/show/frontier and committed creation envelopes,
// with persisted membership, nominal/sound rollups, external gates, graph health,
// and stable member/external role vocabulary. The schema contract publishes the
// Thread lifecycle vocabulary. Init receipts expose safe Projects-scaffold
// removals, preserved legacy content, or an available scaffold-repair command.
// Thread views separate repository graph health from projection health, explain
// completed inconsistencies with stable codes, and hoist list-level graph
// diagnostics. Post-commit Thread creation failures carry the same mutation
// receipt in the error envelope.
// 1.53: task lifecycle receipts expose committed durability, failed start rows
// retain typed eligibility state/blockers/remedy, and post-commit cleanup errors
// carry a structured task/workspace recovery receipt.
// 1.52: task transition rows carry guarded lifecycle detail: exact prior status,
// before/after derived state, typed override, retained blockers, downstream impact
// count/details, and an explanatory remedy. Dependency mutation receipts add the
// same before/after impact shape for directly affected dependents.
// 1.51: dependency blocker/downstream query envelopes carry the queried task's
// derived state, so eligibility is explicit and never inferred from an empty list.
// 1.50: guarded dependency operations add `dependency_mutation` receipts, structured
// partial-failure details on the error envelope, and the `task_blockers` /
// `task_unblocks` diagnostic envelopes. Graph queries carry health, taskflow-owned
// problems, legacy diagnostics, stable reason/path data, and derived task state.
//
// 1.49: task payloads carry `depends_on`, the sorted stable IDs of repository-global
// prerequisites declared by that task. Additive and omitted for tasks without edges.
// The task field/schema contract also recognizes the persisted list while generic
// mutation remains forbidden until the guarded dependency commands land. Lint issues
// may carry `severity: "advisory"`; omitted severity retains the established blocking
// error behavior.
//
// 1.48: the `schema` contract carries `criterion_states` — the non-binary acceptance
// criterion states, published for the same reason `finding_statuses` is: `state` has been a
// criterion wire field since 1.46, and without the set an agent had to trigger an error and
// parse prose to learn it. A finding may carry `note` — the `**Resolution:**` paragraph saying HOW it was
// resolved, written by `audit finding --note` — and `status_decoration`, everything after
// the status word (the date on `fixed 2026-08-24`, the destination on `tracked by <id>`),
// which the wire previously dropped because `status` carries only the vocabulary token.
// Both additive and omitted when absent, so a finding recorded before they existed
// serialises exactly as before.
//
// 1.47: the finding-status vocabulary drops `landed` and gains `tracked` — a finding
// handed to a task, which counts toward `done_findings` because the AUDIT's interest in it
// has concluded. NOT additive: `landed` is no longer accepted, though no audit in the
// corpus ever used it, and a consumer switching on the status set must learn the new word.
//
// 1.46: `task ac --list --json` criteria may carry `state` and `reason` — the disposition
// beyond the checkbox (deferred / wontfix / tracked / n/a) and why. Absent for a plain met/not-met
// criterion, so a body written before the vocabulary existed serialises unchanged.
//
// 1.45: a task's acceptance tally carries `explained` — how many UNMET criteria state why
// (deferred / wontfix / tracked / n/a). Additive and zero for every task written before the criterion
// vocabulary existed, so a consumer that ignores it sees the previous shape.
//
// 1.44: `lint --fix` results may carry `skipped`, marking a file the pass deliberately did
// NOT repair with the reason in `changes` — an invalid id that is still referenced
// elsewhere, or one containing `u`, which Crockford gives no canonical decode. Additive:
// absent on every repaired file, so a consumer that ignores it sees the previous shape.
//
// 1.43: fresh `init --json` receipts may include `registration`, describing the
// best-effort machine-local space registration (including preview vs applied and whether
// the physical checkout was already registered).
// 1.57: `thread compose` and `thread apply` add materialized-plan and resumable
// operation-receipt envelopes; apply failures may carry `thread_apply` recovery
// detail in the standard error envelope.
//
// 1.58: `thread graph` and `thread plan` add one renderer-neutral node, edge,
// wave, role, health, and topology-completeness projection for CLI and future
// primary adapters.
const SchemaVersion = "1.58"

// EncodeJSON writes the payload as compact (un-indented) JSON with a single
// trailing newline. Machine output: pretty-printing is pure token cost for a
// consumer that parses it. Off-tree consumers pipe through `jq .` to read it.
// Exported so render's emit funcs (and any other adapter) encode envelopes
// identically.
func EncodeJSON(w io.Writer, payload any) error {
	enc := json.NewEncoder(w)
	return enc.Encode(payload)
}
