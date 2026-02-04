# TaskFlow

**AI-Native Project Management**

TaskFlow is a local-first tool that bridges the gap between human intuition and AI automation. It uses a flat-file Markdown structure as the source of truth, augmented by a local intelligence layer.

## 🗺 Map

| Directory | Purpose |
| :--- | :--- |
| **[planning/](./planning/)** | **The Plan**. Epics, Tasks, and Research. |
| **[contracts/](./contracts/)** | **Data Model**. The Single Source of Truth (Protobuf) for Tasks. |
| `cmd/` | **The CLI**. Go-based TUI application. |
| `services/` | **The Brain**. Python/FastAPI Semantic Engine. |

## 🚀 Quick Start

1.  **Install Tools**: Ensure you have `go`, `docker`, and `just` installed.
2.  **Generate Contracts**:
    ```bash
    just proto
    ```
3.  **Build CLI**:
    ```bash
    just build-cli
    ```
4.  **Run**:
    ```bash
    ./bin/taskflow --help
    ```

## 🛠 Development

Use `just` to run common tasks:

- `just proto`: Regenerate code from Protobuf definitions.
- `just build`: Build the CLI and API.
- `just dev-up`: Start the local development stack (Database + API).
- `just test`: Run the test suite.