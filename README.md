# TaskFlow: AI-Native Project Management

**TaskFlow** is a local-first, AI-native project management tool designed to bridge the gap between human intuition and AI automation. It uses a flat-file Markdown structure as the source of truth, augmented by a local intelligence layer (JSON Index + Vector Database).

## 🚀 Mission
To offload the cognitive load of project management (searching, correlating, prioritizing) from humans and LLMs onto a specialized, low-latency local tool.

## 🏗️ Architecture

TaskFlow operates as a hybrid system:

### 1. The Storage Layer (Source of Truth)
- **Markdown Files**: Git-native, human-readable files.
    - `epics/`: Strategic themes.
    - `tasks/`: Actionable units of work.
    - `research/`: ADRs and technical spikes.
- **Benefits**: Works with any editor, version controlled via Git, perfect "context" for LLMs.

### 2. The Intelligence Layer (Dockerized Service)
A local `docker-compose` stack providing the "Brain":
- **PostgreSQL + pgvector**: Stores task metadata and semantic embeddings.
- **FastAPI / Go API**: Exposes endpoints for the CLI.
- **Watcher Service**: Monitors file system events (`inotify`) to auto-update the index and vector store when Markdown files change.
    - *Note*: No historical vector storage needed; current state is what matters.

### 3. The Interaction Layer (CLI)
- **`taskflow` (Python CLI)**: The user interface.
    - `taskflow list`: Fast queries via API.
    - `taskflow related <task>`: Semantic search via vector store.
    - `taskflow project start`: Interactive wizard for bundling tasks.
    - `taskflow recommend`: Dynamic, context-aware suggestions.

## 🌟 Core Features

### Hybrid Indexing
- **JSON Index (Fast)**: Immediate lookups for "Find all Tier 1 tasks".
- **Vector Store (Smart)**: Semantic understanding for "Find tasks related to auth refactoring".

### AI Collaboration
- **Prompt-to-Task**: "Create a project for Q1 cleanup" -> Tool generates file stubs.
- **Context Injection**: Tool can dump a compressed, highly relevant context summary for an LLM session to save tokens.

### Workflow Automation
- **Projects/Sprints**: Cross-cutting bundles of tasks defined by tags, not folders.
- **Dynamic Recommendations**: "You just finished Task A; Task B is unblocked and high priority."

## 🗺️ Implementation Plan

We will bootstrap TaskFlow using TaskFlow (meta!).

1.  **Phase 1: Core CLI & File System** (Porting `pm` script logic)
2.  **Phase 2: Indexing Engine** (JSON State + File Watcher)
3.  **Phase 3: Semantic Layer** (Docker + pgvector + Embeddings)
4.  **Phase 4: API & Interaction** (The "App" experience)

---
*This project is being incubated within `desirelines-planning` but is designed to be spun out as a standalone open-source tool.*
