---
status: ready-to-start
epic: taskflow-v1-core
effort: 1-2 hours
tier: 2
priority: medium
project: taskflow-bootstrap
tags: [branding, documentation, architecture]
---

# TaskFlow 00: Naming & Terminology

**Goal**: Establish professional, consistent terminology for the system components, specifically replacing the placeholder "Brain".

## Context
"Brain" is too sci-fi/informal. We need a functional name for the Python/Docker service that handles Embeddings and Vector Search.

## Candidates for "The API Layer"
1.  **Intelligence Service** (Clear, but verbose)
2.  **Semantic Engine** (Accurate)
3.  **Vector Store** (Too specific to the DB?)
4.  **Core** (Vague)
5.  **TaskFlow Server** (Standard)

## Candidates for "The Search Types"
1.  **Fast Path**: "Local Search" / "Index Search"
2.  **Smart Path**: "Semantic Search" / "Deep Search"

## The Fallback Strategy
If the **Semantic Engine** is offline (Docker down):
- `taskflow related` should gracefully degrade.
- **Fallback**: Go-based fuzzy matching (e.g., using `fzf` logic or simple token overlap) against the local JSON index.
- **UX**: Show a warning: "⚠️ Semantic Engine unavailable. Using keyword matching."

## Deliverables
- [ ] Select final names.
- [ ] Update `taskflow/README.md` and Architecture diagrams.
- [ ] Document the Fallback Strategy in the Architecture Epic.
