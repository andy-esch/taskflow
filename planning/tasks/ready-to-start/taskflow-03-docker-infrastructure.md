---
status: ready-to-start
epic: taskflow-v1-core
effort: 4-6 hours
tier: 1
priority: high
project: taskflow-bootstrap
tags: [docker, postgres, infrastructure]
---

# TaskFlow 03: Docker Infrastructure

**Goal**: Setup the local containerized environment.

## Requirements
- Create `docker-compose.yml`.
- Service 1: PostgreSQL 16+ with `pgvector` extension installed.
- Service 2: Backend API (placeholder container for now).
- Volume management for persistent data (`.taskflow/data`).

## Deliverables
- `docker-compose up` starts a database ready for vector operations.
- Connection script to verify DB access.
