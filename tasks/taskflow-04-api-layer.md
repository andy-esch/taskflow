---
status: ready-to-start
epic: taskflow-v1-core
effort: 4-6 hours
tier: 1
priority: high
project: taskflow-bootstrap
tags: [python, fastapi, docker, api]
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
