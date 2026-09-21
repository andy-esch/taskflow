---
schema: 1
id: 6gbn4v2jpf2m
status: in-progress
description: Turn the v0.22 machine-contract foundation into complete discovery and bounded planning reads for agents.
goal: Agents discover safe commands and exit behavior, retain durable Thread handles, and page large planning corpora without prose scraping.
created: "2026-09-19"
tags: [cli, agents, json, contract]
tasks: [6g63hhk3eddf, 6gbn4g1bb7wr, 6gbn4g1eypzs, 6gbn4g1j40pj]
updated_at: "2026-09-20"
started_at: "2026-09-20"
---

# Thread: Make CLI contracts self-describing and bounded

**Goal.** Agents discover safe commands and exit behavior, retain durable Thread handles, and page large planning corpora without prose scraping.

## Context

v0.22.0 made the existing machine surface explicit and enforceable: projected registries reject
drift, task and audit projections expose durable IDs, and revision 1.68 publishes a monotonic
compatibility policy. The remaining audit findings are no longer policy ambiguities; they are four
bounded product gaps in discovery and token-efficient reads.

The members are deliberately independent at creation. Exit taxonomy, command discovery, Thread
projection, and query bounds share a release story but none currently requires another's code to be
correct. Do not invent dependency edges merely to serialize implementation. Start this Thread only
after the smaller Tool-owned actionable sub-entities Thread closes, then use its frontier to choose
work by value and review capacity.

Compact body-mutation receipts remain outside this Thread because changing their default envelope
is a separate compatibility decision. Likewise, this effort publishes CLI capabilities; it does not
create a second execution API, an MCP server, or the future web adapter.
