# Taskflow GitHub Actions

Reusable, public GitHub Actions provided by **Taskflow** to automate planning tree validation and repository governance in CI.

## Available Actions

| Action | Path | Description |
| :--- | :--- | :--- |
| **Lint** | [`actions/lint`](./lint/) | Validates planning entities, frontmatter schemas, audit projections, and the task dependency DAG with `tskflwctl lint`. |

---

## `actions/lint`

Automate planning tree validation on pull requests and pushes to `main`. The action acquires the pre-built `tskflwctl` static binary for the runner architecture from GitHub Releases (~1s execution time) and runs `tskflwctl lint`.

### Usage

```yaml
name: Planning Lint

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    name: Lint Planning Tree
    runs-on: ubuntu-latest
    permissions:
      contents: read

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Validate planning tree
        uses: andy-esch/taskflow/actions/lint@v0.22.0
```

### Inputs

| Input | Description | Default |
| :--- | :--- | :--- |
| `version` | Target version of `tskflwctl` (`latest`, a tag like `v0.22.0`, `local` for checkout builds, or a git ref) | `latest` |
| `links` | Also check Markdown body cross-links for missing files (`--links`) | `false` |
| `working-directory` | Directory containing `.tskflwctl.toml` or `planning/` | `.` |
| `args` | Additional arguments passed directly to `tskflwctl lint` (e.g. `--no-color`) | `""` |
| `github-token` | GitHub token passed to `gh` CLI for release asset downloads (avoids API rate limits) | `${{ github.token }}` |

### Behavior & Gating

- **Fast binary install**: Downloads and extracts the pre-built Linux or macOS static binary matching the runner architecture.
- **Fail-safe fallback**: If a pre-built binary is unavailable for the requested ref (e.g. an unreleased branch or `@main`), it falls back to compiling via `go install`.
- **Exit codes**: Standard `tskflwctl lint` exit semantics apply; any validation error (exit code `11`) or unclassified failure (exit code `1`) causes the workflow step to fail.
