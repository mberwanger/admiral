SHELL:=/usr/bin/env bash
.DEFAULT_GOAL:=all

MAKEFLAGS += --no-print-directory

PROJECT_ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: help # Print this help message.
help:
	@grep -E '^\.PHONY: [a-zA-Z_-]+ .*?# .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = "(: |#)"}; {printf "%-30s %s\n", $$2, $$3}'

.PHONY: proto # Generate proto assets.
proto:
	./tools/buf.sh generate --clean

.PHONY: proto-lint # Lint the generated proto assets.
proto-lint:
	./tools/buf.sh lint

.PHONY: proto-verify # Verify proto changes.
proto-verify:
	@$(MAKE) proto
	tools/ensure-no-diff.sh server/api web/src/api

.PHONY: server # Build the standalone server.
server: preflight-checks-server
	cd server && go build -o ../build/server

.PHONY: server-with-assets # Build the server with web assets.
server-with-assets: preflight-checks-server
	cd server && go run cmd/assets/generate.go ../web/build && go build -tags withAssets -o ../build/server

.PHONY: server-dev # Start the server in development mode.
server-dev: preflight-checks-server
	tools/air.sh

.PHONY: server-lint # Lint the server code.
server-lint: preflight-checks-server
	tools/golangci-lint.sh run --timeout 2m30s

.PHONY: server-lint-fix # Lint and fix the server code.
server-lint-fix:
	tools/golangci-lint.sh run --fix
	cd server && go mod tidy

.PHONY: server-test # Run unit tests for the server code.
server-test: preflight-checks-server
	cd server && go test -race -covermode=atomic ./...

.PHONY: server-verify # Verify go modules' requirements files are clean.
server-verify:
	cd server && go mod tidy
	tools/ensure-no-diff.sh server

.PHONY: preflight-checks-server
preflight-checks-server:
	@tools/preflight-checks.sh server

.PHONY: web # Build production web assets.
web: preflight-checks-web bun-install
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

.PHONY: preflight-checks-agent
preflight-checks-agent:
	@tools/preflight-checks.sh agent

.PHONY: agent # Build the agent.
agent: preflight-checks-agent
	cd agent && go build -o ../build/agent

.PHONY: agent-lint # Lint the agent code.
agent-lint: preflight-checks-agent
	tools/golangci-lint.sh run --timeout 2m30s

.PHONY: agent-lint-fix # Lint and fix the agent code.
agent-lint-fix:
	tools/golangci-lint.sh run --fix
	cd agent && go mod tidy

.PHONY: agent-test # Run unit tests for the agent code.
agent-test: preflight-checks-agent
	cd agent && go test -race -covermode=atomic ./...

.PHONY: server-verify # Verify go modules' requirements files are clean.
agent-verify:
	cd agent && go mod tidy
	tools/ensure-no-diff.sh agent

.PHONY: bun-install # Install web dependencies.
bun-install:
	bun install --cwd web --frozen-lockfile

.PHONY: preflight-checks-web
preflight-checks-web:
	@tools/preflight-checks.sh web

.PHONY: dev # Run the application in development mode.
dev:
	$(MAKE) -j2 server-dev web-dev

.PHONY: lint # Lint all of the code.
lint: server-lint web-lint

.PHONY: lint-fix # Lint and fix all of the code.
lint-fix: server-lint-fix web-lint-fix

.PHONY: test # Unit test all of the code.
test: server-test web-test

.PHONY: verify # Verify all of the code.
verify: server-verify web-verify

.PHONY: clean # Remove build and cache artifacts.
clean:
	rm -rf build
	cd server && rm -rf .air && rm cmd/assets/generated_assets.go
	cd web && rm -rf build node_modules

.PHONY: preflight-checks
preflight-checks:
	@tools/preflight-checks.sh