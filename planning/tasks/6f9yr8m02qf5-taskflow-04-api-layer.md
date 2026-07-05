---
status: deprecated
epic: 00-taskflow-v1-core
effort: 4-6 hours
tier: 1
priority: high
project: taskflow-bootstrap
tags: [python, fastapi, docker, api]
deprecated_at: 2026-06-07
deprecated_reason: Superseded by the scoped tskflwctl design (epic 17-pm-go-cli); old over-ambitious taskflow vision (docker/api/semantic/AI-gen) retired 2026-06-07. Research mined + kept.
updated_at: 2026-06-07
id: 6f9yr8m02qf5
---

# TaskFlow 04: API Layer (Python)

**Goal**: Build the backend "Brain" service.

## Requirements
- Setup **FastAPI** project structure.
- Create endpoints for:
    - `GET /tasks`: List/Filter tasks.
    - `POST /search`: Semantic search (placeholder for now).
- Dockerize the service.
- Ensure it can connect to the Postgres container.

## Deliverables
- Docker container running FastAPI.
- `curl localhost:8000/health` returns 200.
