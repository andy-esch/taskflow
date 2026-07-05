---
status: completed
epic: 00-taskflow-v1-core
effort: 4-6 hours
tier: 1
priority: high
project: taskflow-bootstrap
tags: [architecture, protobuf, json-schema, contracts]
completed_at: 2026-02-04
id: 6e2bwq0037wm
---

# TaskFlow 00: Schema & Contracts Strategy

**Goal**: Establish the "Source of Truth" for data structures shared between the Go CLI, Python API, and Markdown Frontmatter.

## The Problem
We have three consumers of our data model:
1.  **Go CLI**: Needs structs for TUI rendering and JSON parsing.
2.  **Python API**: Needs Pydantic models for FastAPI and Vector logic.
3.  **Markdown Files**: Needs validation rules (JSON Schema) for frontmatter.

We do NOT want to manually keep three definitions in sync.

## Options to Evaluate

### Option A: Protobuf (gRPC Style)
- **Source**: `.proto` files.
- **Generate**: Go structs (`protoc-gen-go`), Python classes (`protoc-gen-python`), JSON Schema (`protoc-gen-jsonschema`).
- **Pros**: Strongly typed, industry standard, handles breaking changes well.
- **Cons**: Protobuf types in Python/Go can be verbose/unidiomatic compared to native structs.

### Option B: JSON Schema First
- **Source**: `.schema.json` files.
- **Generate**: Go structs (`quicktype` or similar), Python Pydantic (`datamodel-code-generator`).
- **Pros**: Native to the frontmatter problem.
- **Cons**: Tooling for code generation varies in quality.

### Option C: Pydantic First (Python Centric)
- **Source**: Python Pydantic models.
- **Generate**: JSON Schema (built-in). Go structs (via `datamodel-code-generator` or similar).
- **Pros**: Best for the "Brain" (Python).
- **Cons**: Go generation is second-class.

## Recommendation to Validate
**Protobuf** seems strongest given the likely HTTP/gRPC communication between CLI and API.

## Deliverables
- [ ] Create a "Proof of Concept" repo/folder.
- [ ] Define a `Task` message in Proto.
- [ ] Generate Go code, Python code, and JSON Schema from it.
- [ ] Validate that the JSON Schema successfully validates a sample Markdown frontmatter.
