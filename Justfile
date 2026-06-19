# taskflow Justfile — tskflwctl (Go CLI)
# The main package lives at ./cmd/tskflwctl (standard Go layout), so use
# `just build` / `just run ...` rather than `go build .` at the repo root.

# Default: build the CLI
default: build

# Version stamped into the binary (git tag/sha, or "dev").
version := `git describe --tags --always --dirty 2>/dev/null || echo dev`
ldflags := "-X github.com/andy-esch/taskflow/internal/cli.version=" + version

# Build the binary to ./bin/tskflwctl
build:
	@echo "Building tskflwctl {{version}} → bin/tskflwctl"
	go build -ldflags "{{ldflags}}" -o bin/tskflwctl ./cmd/tskflwctl

# Run without installing: `just run task list --json`
run *ARGS:
	go run ./cmd/tskflwctl {{ARGS}}

# Install onto $GOBIN / $GOPATH/bin (so `tskflwctl` is on PATH)
install:
	go install -ldflags "{{ldflags}}" ./cmd/tskflwctl

# Print the completion script for a shell (bash|zsh|fish|powershell) to stdout.
completion SHELL="zsh":
	go run ./cmd/tskflwctl completion {{SHELL}}

# Install zsh tab-completion → ~/.zsh/completions/_tskflwctl (one-time).
# Run `just install` first so `tskflwctl` is on PATH when completion fires.
completion-zsh:
	mkdir -p ~/.zsh/completions
	go run ./cmd/tskflwctl completion zsh > ~/.zsh/completions/_tskflwctl
	@echo 'Installed → ~/.zsh/completions/_tskflwctl'
	@echo 'First time? add to ~/.zshrc:  fpath=(~/.zsh/completions $fpath)  then:  autoload -Uz compinit && compinit'

# Run tests (with the race detector — the fsnotify/debounce code is exactly
# where races would live, and they should surface locally before CI)
test:
	go test -race ./...

# Lint (golangci-lint — needs a v2.x binary BUILT WITH Go ≥ go.mod's target;
# .golangci.yml is v2-schema. `go install github.com/golangci/golangci-lint/v2/
# cmd/golangci-lint@latest` compiles with your local toolchain, which sidesteps
# the prebuilt-binary Go-version skew that brew/CI downloads can hit.)
lint:
	golangci-lint run ./...

# Scan dependencies + stdlib usage for known vulnerabilities
vulncheck:
	govulncheck ./...

# Format Go sources + tidy lint formatting
fmt:
	gofmt -w cmd internal
	golangci-lint fmt ./... || true

# Tidy modules
tidy:
	go mod tidy

# Check Go module tidiness (fails if go.mod or go.sum would change)
tidy-check:
	go mod tidy -diff

# Regenerate the CLI reference under docs/cli/ from the cobra command tree.
docs:
	go run ./internal/tools/docgen -out docs/cli

# Fail if the committed CLI reference is stale (CI drift guard). Regenerates,
# then errors if anything changed — so editing a command/flag without running
# `just docs` breaks the build.
docs-check: docs
	git diff --exit-code docs/cli

# Dry-run a release locally: build all darwin/linux binaries + checksums into
# ./dist via goreleaser, publishing nothing. Mirrors what the manual
# workflow_dispatch path produces. Needs goreleaser installed
# (go install github.com/goreleaser/goreleaser/v2@latest).
release-snapshot:
	goreleaser release --snapshot --clean

# Validate .goreleaser.yml without building.
release-check:
	goreleaser check

# Clean build artifacts
clean:
	rm -rf bin dist
