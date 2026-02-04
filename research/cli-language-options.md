# Research: CLI Language & Distribution for TaskFlow

**Status**: Proposal
**Created**: 2026-01-03
**Context**: We need a robust, distributable CLI tool named `taskflow`. The heavy lifting (embeddings, database) is moving to a Dockerized API layer. The CLI is primarily a client for this API.

## The Problem with Python CLIs
- **Distribution**: Requires users to have Python installed, manage venvs, or use tools like `pipx`.
- **Startup Time**: Python has a noticeable startup cost (imports), which hurts the "snappy" feel we want.
- **Dependencies**: Bundling libraries (like `requests` or UI libs) complicates distribution.

## The Architecture Shift
**Original Idea**: CLI does everything (Embeddings + Logic).
**New Model**:
1.  **Server (Docker)**: Python (FastAPI) + Postgres (pgvector) + Sentence-Transformers. This is where the heavy dependencies live.
2.  **Client (CLI)**: A thin, compiled binary that talks to the local API (e.g., `http://localhost:8000`).

## Option 1: Go (Golang) ⭐ **Recommended**
- **Pros**:
    - **Single Binary**: Compiles to a static binary (`taskflow`) that runs anywhere with no dependencies.
    - **Speed**: Instant startup (sub-10ms).
    - **Libraries**: Excellent CLI libraries like `Cobra` (similar structure to Python's Click/Typer) or `Bubble Tea` (for rich TUI experiences).
    - **Networking**: Standard library HTTP client is rock solid.
- **Cons**: Learning curve if you only know Python.

## Option 2: Rust
- **Pros**: Even faster, memory safe, great ecosystem (`clap`, `ratatui`).
- **Cons**: Slower compilation, steeper learning curve.

## Option 3: Python (Compiled)
- **Tools**: `PyInstaller` or `Nuitka`.
- **Pros**: Stick with one language.
- **Cons**: Resulting binaries are large (bundle the interpreter) and slower to start.

## Recommendation: Go + Bubble Tea
If we want a "really nice terminal experience," **Go with the [Bubble Tea](https://github.com/charmbracelet/bubbletea)** framework is the industry standard for modern, beautiful CLIs (used by GitHub CLI, etc.).

### Workflow
1.  **User**: `taskflow list`
2.  **CLI (Go)**:
    - Checks if Docker stack is running.
    - Sends `GET /tasks` to localhost API.
    - Renders interactive TUI list.
3.  **Server (Python)**:
    - Handles database query.
    - Returns JSON.

This separates concerns perfectly: Python for the "Brain" (AI/Embeddings), Go for the "Face" (UI/Speed).
