# TaskFlow Justfile

# Default recipe: build the CLI
default: build-cli

# --- Protobuf ---

# Generate code from protobuf definitions
proto:
	buf generate

# --- Development ---

# Start the dev stack
dev-up:
	docker-compose -f dev/docker-compose.yml up -d

# Stop the dev stack
dev-down:
	docker-compose -f dev/docker-compose.yml down

# --- Build ---

# Build the Go CLI
build-cli:
	@echo "Building Go CLI..."
	go build -o bin/taskflow ./cmd/taskflow

# Build the Python API Docker image
build-api:
	@echo "Building Python API..."
	cd services/semantic-engine && docker build -t taskflow-api .

# --- Testing ---

# Run all tests
test: test-go test-python

# Run Go tests
test-go:
	go test ./...

# Run Python tests
test-python:
	cd services/semantic-engine && pytest
