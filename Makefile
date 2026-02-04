# Universal Makefile for TaskFlow

.PHONY: all dev build-cli build-api test

all: build-cli build-api

# --- Development ---

dev:
	@echo "Starting dev environment..."
	docker-compose -f dev/docker-compose.yml up -d
	go run ./cmd/taskflow

# --- Build ---

build-cli:
	@echo "Building Go CLI..."
	go build -o bin/taskflow ./cmd/taskflow

build-api:
	@echo "Building Python API..."
	cd services/semantic-engine && docker build -t taskflow-api .

# --- Testing ---

test: test-go test-python

test-go:
	go test ./...

test-python:
	cd services/semantic-engine && pytest
