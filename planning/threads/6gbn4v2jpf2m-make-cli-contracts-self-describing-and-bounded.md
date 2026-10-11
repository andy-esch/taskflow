---
schema: 1
id: 6gbn4v2jpf2m
status: completed
description: Turn the v0.22 machine-contract foundation into complete discovery and bounded planning reads for agents.
goal: Agents discover safe commands and exit behavior, retain durable Thread handles, and page large planning corpora without prose scraping.
created: "2026-09-19"
tags: [cli, agents, json, contract]
tasks: [6g63hhk3eddf, 6gbn4g1bb7wr, 6gbn4g1eypzs, 6gbn4g1j40pj, 6ggfd81jg0qg, 6ggjmtmdd54w, 6gjdgpq17sr6]
updated_at: "2026-10-10"
started_at: "2026-09-20"
ended_at: "2026-10-10"
---

# Thread: Make CLI contracts self-describing and bounded

**Goal.** Agents discover safe commands and exit behavior, retain durable Thread handles, and page large planning corpora without prose scraping.

## Context

v0.22.0 made the existing machine surface explicit and enforceable: projected registries reject
drift, task and audit projections expose durable IDs, and revision 1.68 publishes a monotonic
compatibility policy. The remaining audit findings at that point identified four bounded
product gaps in discovery and token-efficient reads rather than policy ambiguities.

The original members were deliberately independent at creation. Exit taxonomy, command discovery,
Thread projection, and query bounds shared a release story but did not require one another's code
to be correct. The Thread followed the smaller Tool-owned actionable sub-entities Thread; its
frontier guided work by value and review capacity rather than artificial serial dependencies.

Compact body-mutation receipts remain outside this Thread because changing their default envelope
is a separate compatibility decision. Likewise, this effort publishes CLI capabilities; it does not
create a second execution API, an MCP server, or the future web adapter.

## Identity-recovery follow-up (2026-10-04)

[Executable ID-less Thread recovery instructions](../tasks/6ggfd81jg0qg-make-id-less-thread-planning-recovery-instructions-executable.md)
delivered a bounded diagnostic follow-up from the workspace-parity review: an agent should not be sent
to `config migrate` when that command cannot run without a marker. Execute the suggested repairs,
keep stale-plan and pointer identity refusals intact, and avoid a general remediation framework.
It followed the shipped configuration lifecycle foundation and proceeded independently of
list-query bounds, without blocking adapter-neutral refactor closeout.

## Bootstrap targeting follow-up (2026-10-04)

[Prevent silently ignored init selectors](../tasks/6ggjmtmdd54w-prevent-silently-ignored-target-selectors-during-init.md)
resolved a dogfood finding: `init` accepted global `-C` but scaffolded/repaired the
`--path` target, which defaults to the caller's working directory. Bootstrap selectors now
have explicit behavior and no-write refusals for unsupported/conflicting targets; fresh
initialization remains independent of existing-repo discovery. This followed the completed
configuration lifecycle foundation independently of the other members, without blocking
adapter-neutral closeout.

## Delivery and closeout (2026-10-10)

The original four product gaps and both recovery/bootstrap follow-ups are implemented.
Bounded queries now share a portable application contract, opt-in continuation, and metadata
search; the ID-less repair commands preserve existing tree layout and pointer bindings.
Independent reviews confirmed recovery behavior and prompted two bounded-query corrections:
compact Thread output retains repository graph context, and malformed cursor digests receive
validation rather than changed-view advice. Neither changes legacy unbounded output.

Self-review found one additional pre-existing identity mismatch, tracked in
[data-directory pointer parity](../tasks/6gjdgpq17sr6-preserve-planning-identity-for-pointers-addressing-data-subdirectories.md).
That final member now restores address/identity parity for a single-governing-marker planning
tree, including custom subdirectories, symlinks and canonical-worktree pointers. Actual init and
migration bindings survive discovery and identity-bound Thread plan replay; missing-ID repair
migrates the governing marker without silently rebinding. Build, full race tests, code/planning
lint and generated-documentation checks pass. All seven delivery members are complete; the
Thread's product scope is drained. Independent pointer-parity reviews found no new in-scope
defects; concrete replay/recovery probes and mutation checks support closeout. A permanent
data-alias replay regression pins governing-marker revalidation through ordinary and workspace
opening. Antigravity's broad race wording and pending transfer attestation are explicitly
qualified in its owner disposition; no global configuration transaction is claimed. Review
evidence is linked in the final member, and this branch's closeout is not a claim of merge or release.

### What remains outside this Thread

[Overlapping-marker identity authority](../tasks/6gjgccnzxj0a-decide-and-enforce-identity-authority-for-overlapping-planning-markers.md)
tracks a separately reproduced configuration-topology issue: initializing both a repo marker and
a marker inside its configured data tree can assign two IDs to one physical tree. It follows the
alias fix but needs its own authority/recovery policy; do not silently pick an ID or delete a
marker. Compact body-mutation receipts, new execution interfaces and the separately scoped
TUI/documentation Threads also remain outside this completed delivery effort.

Use `tskflwctl thread show make-cli-contracts-self-describing-and-bounded` for the closeout
projection and `tskflwctl thread graph make-cli-contracts-self-describing-and-bounded` for its
preserved execution graph. A future discovered CLI-contract defect should be scoped explicitly,
not folded into this history without a new task or a deliberate reopen.
