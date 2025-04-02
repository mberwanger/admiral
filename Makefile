# Use bash as the shell, with environment lookup
SHELL := /usr/bin/env bash

.DEFAULT_GOAL := all

MAKEFLAGS += --no-print-directory --silent

VERSION ?= 0.0.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILT_BY ?= $(shell whoami)
PROJECT_ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: help # Print this help message.
help:
	@grep -E '^\.PHONY: [a-zA-Z_-]+ .*?# .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = "(: |#)"}; {printf "%-30s %s\n", $$2, $$3}'

.PHONY: dev # Run the application in development mode.
dev:
	$(MAKE) -j2 server-dev web-dev

.PHONY: lint # Lint all of the code.
lint: server-lint web-lint proto-lint

.PHONY: lint-fix # Lint and fix all of the code.
lint-fix: server-lint-fix web-lint-fix

.PHONY: test # Unit test all of the code.
test: server-test web-test

.PHONY: verify # Verify all of the code.
verify: server-verify web-verify proto-verify

.PHONY: clean # Remove build and cache artifacts.
clean:
	rm -rf build .air cmd/assets/generated_assets.go web/build web/node_modules

.PHONY: proto # Generate proto assets.
proto:
	rm -rf api web/src/api && ./tools/buf.sh generate --clean

.PHONY: proto-lint # Lint the generated proto assets.
proto-lint:
	./tools/buf.sh lint

.PHONY: proto-verify # Verify proto changes.
proto-verify:
	@$(MAKE) proto
	tools/ensure-no-diff.sh server/api web/src/api

.PHONY: server # Build the standalone server.
server: preflight-checks-go
	go build -o ./build/admiral-server \
		-ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -X main.builtBy=$(BUILT_BY)"

.PHONY: server-with-assets # Build the server with web assets.
server-with-assets: preflight-checks-go
	go run cmd/assets/generate.go ./web/build && go build -tags withAssets -o ./build/admiral-server \
		-ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -X main.builtBy=$(BUILT_BY)"

.PHONY: server-dev # Start the server in development mode.
server-dev: preflight-checks-go
	tools/air.sh

.PHONY: server-lint # Lint the server code.
server-lint: preflight-checks-go
	tools/golangci-lint.sh run --timeout 2m30s

.PHONY: server-lint-fix # Lint and fix the server code.
server-lint-fix: preflight-checks-go
	tools/golangci-lint.sh run --fix
	cd server && go mod tidy

.PHONY: server-test # Run unit tests for the server code.
server-test: preflight-checks-go
	cd server && go test -race -covermode=atomic ./...

.PHONY: server-verify # Verify go modules' requirements files are clean.
server-verify: preflight-checks-go
	cd server && go mod tidy
	tools/ensure-no-diff.sh server

.PHONY: web # Build production web assets.
web: bun-install
	bun run --cwd web build

.PHONY: web-dev-build # Build development web assets.
web-dev-build: bun-install
	bun run --cwd web preview

.PHONY: web-dev # Start the web in development mode.
web-dev: bun-install
	bun run --cwd web dev

.PHONY: web-lint # Lint the web code.
web-lint: bun-install
	bun run --cwd web lint

.PHONY: web-lint-fix # Lint and fix the web code.
web-lint-fix: bun-install
	bun run --cwd web lint:fix

.PHONY: web-test # Run unit tests for the web code.
web-test: bun-install
	bun test --cwd web

.PHONY: web-verify # Verify web packages are sorted.
web-verify: bun-install
	bun run --cwd web lint:packages

.PHONY: bun-install # Install web dependencies.
bun-install: preflight-checks-bun
	bun install --cwd web --frozen-lockfile

.PHONY: preflight-checks-bun
preflight-checks-bun:
	@tools/preflight-checks.sh bun

.PHONY: preflight-checks-go
preflight-checks-go:
	@tools/preflight-checks.sh go

.PHONY: preflight-checks
preflight-checks:
	@tools/preflight-checks.sh
