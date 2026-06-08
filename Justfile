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

# Clean build artifacts
clean:
	rm -rf bin
