# taskflow

Home of **`tskflwctl`** — a local-first planning CLI over markdown+frontmatter
task/epic/audit files. It's the Go port of the Python `pm` prototype, and it
dogfoods on its own planning under [`planning/`](./planning/).

## Map

| Path | Purpose |
| :--- | :--- |
| **[`cmd/tskflwctl/`](./cmd/tskflwctl/)** | The CLI entrypoint (thin composition root). |
| **[`internal/`](./internal/)** | `domain` (pure) · `core` (use cases) · `store` (markdown adapter) · `cli` (cobra) · `config`. |
| **[`planning/`](./planning/)** | This repo's own epics, tasks, and research (self-hosted). |
| **[`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md)** | One-screen orientation: the primary/secondary-adapter design. |

## Quick start

```bash
just build              # → bin/tskflwctl
just run task list      # run without installing
just install            # put tskflwctl on $PATH
```

## Development

`just` wraps the common tasks:

- `just build` — build `bin/tskflwctl`
- `just run *ARGS` — `go run ./cmd/tskflwctl …`
- `just test` — `go test ./...`
- `just lint` — `golangci-lint run ./...`
- `just fmt` — gofmt + lint formatting
- `just tidy` — `go mod tidy`

Design rationale lives in [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) and
[`planning/epics/17-pm-go-cli.md`](./planning/epics/17-pm-go-cli.md).

> **Note:** `services/` and `internal/tui/` are leftovers from an earlier Go
> spike (a Python "brain" and a Bubble Tea TUI sketch) and are not part of the
> current CLI. The TUI will be rebuilt over the same `core` in a later phase.
