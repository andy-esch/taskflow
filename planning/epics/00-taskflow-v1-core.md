---
status: deprecated
description: SUPERSEDED by 17-pm-go-cli — original over-ambitious taskflow vision (docker/pgvector/semantic/AI). Kept for history.
created: 2026-02-04
tags: [pm-tooling, archived, superseded]
---

# Epic: TaskFlow v1 Core — ⛔ SUPERSEDED

> **Superseded 2026-06-07 by [17-pm-go-cli](17-pm-go-cli.md)** (`planning/epics/17-pm-go-cli.md`).
> This was the original, over-ambitious vision (Dockerized semantic engine,
> pgvector, MCP, AI generation). The scoped design — CLI-first, single static
> binary, **no intelligence layer** — lives in **17-pm-go-cli** + the
> `planning/research/2026-06-06-*` docs. The 9 active tasks here were
> deprecated 2026-06-07; the 4 completed ones remain as history below.

**Status**: Archived (superseded)
**Goal (original)**: Build the MVP of TaskFlow, a local-first, AI-native project management tool.

## Scope
Transform the current `pm` script concept into a robust, dockerized application with semantic search capabilities and AI automation.

## Architecture
- **CLI (Go)**: Fast, single-binary interface using **Cobra** and **Bubble Tea**.
    - **Hybrid Search**: Reads local JSON index directly for speed; queries API for intelligence.
    - **TUI**: Split-pane interaction, Vim keybindings (`j/k`), `y/n` selection.
- **Backend (Python)**: Dockerized "Semantic Engine" service (FastAPI).
    - **AI Abstraction**: Uses **LiteLLM** to support local (Ollama) and cloud (Claude/GPT) models.
    - **MCP Server**: Exposes tools/resources for external AI Agents (IDEs).
- **Database**: PostgreSQL with `pgvector` for task metadata and semantic embeddings.
- **Watcher**: Service to sync Markdown files to DB and auto-update the JSON index.

## Success Criteria
- [ ] `taskflow` CLI provides instant list/filter via local JSON.
- [ ] Semantic Search finds related tasks even if keywords don't match.
- [ ] "Project Start" wizard allows interactive bundling + AI prompt expansion.
- [ ] Documentation (Roadmap/Changelog) is auto-generated from task state.
- [ ] System handles split-repo setups (Plan vs Code) via config.

## Related Tasks

### Phase 0: Foundation (Monorepo & Standards)
- `taskflow-00-monorepo-structure.md` - Pattern C (Hybrid Standard), single Go/Python modules.
- `taskflow-00-schema-contracts.md` - ✅ Implemented (`./contracts/README.md` — dead link, disused). Protobuf as source of truth.
- `taskflow-00-documentation-standards.md` - Definitions (Project vs Epic), Style Guide.
- `taskflow-00-naming-and-terminology.md` - "Semantic Engine", "Fast Path".
- `taskflow-00-cross-repo-context.md` - Config schema for split repositories.

### Phase 1: Core CLI & File System
- `taskflow-01-cli-skeleton-go.md` - ✅ [Implemented](../../cmd/taskflow). Cobra/Viper setup.
- `taskflow-02-json-schema-design.md` - ✅ [Implemented](../../internal/index). Scalable JSON index for 10k+ tasks.

### Phase 2: Intelligence Layer (Docker)
- `taskflow-03-docker-infrastructure.md` - Postgres + pgvector setup.
- `taskflow-04-api-layer.md` - Python FastAPI service ("Semantic Engine").

### Phase 3: Indexing & Watcher
- `taskflow-05-file-watcher.md` - Real-time sync service.
- `taskflow-06-embedding-pipeline.md` - Vector embedding generation.

### Phase 4: Interaction & AI
- `taskflow-07-semantic-search-cmd.md` - `taskflow related` command.
- `taskflow-08-doc-generation-engine.md` - Auto-generate Roadmap/Changelog.
- `taskflow-09-interactive-project-start.md` - TUI Wizard (Bubble Tea, Vim keys).
- `taskflow-09a-ai-task-generation-flow.md` - Prompt-to-Task generation pipeline.
- `taskflow-10-testing-strategy.md` - Validation logic for Task Templates & Content.