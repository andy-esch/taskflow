# taskflow Justfile — tskflwctl (Go CLI)
# The main package lives at ./cmd/tskflwctl (standard Go layout), so use
# `just build` / `just run ...` rather than `go build .` at the repo root.

# Default: build the CLI
default: build

# Build the binary to ./bin/tskflwctl
build:
	@echo "Building tskflwctl → bin/tskflwctl"
	go build -o bin/tskflwctl ./cmd/tskflwctl

# Run without installing: `just run task list --json`
run *ARGS:
	go run ./cmd/tskflwctl {{ARGS}}

# Install onto $GOBIN / $GOPATH/bin (so `tskflwctl` is on PATH)
install:
	go install ./cmd/tskflwctl

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

# Run tests
test:
	go test ./...

# Lint (golangci-lint)
lint:
	golangci-lint run ./...

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

# Clean build artifacts
clean:
	rm -rf bin
