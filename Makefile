# Copyright 2025 Navigator Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Use nix develop shell for all commands to ensure dependencies are available
# Disable dirty git tree warnings
export NIX_CONFIG := warn-dirty = false
SHELL := nix develop --command bash

# Version variables
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/liamawhite/navigator/pkg/version.version=$(VERSION) -X github.com/liamawhite/navigator/pkg/version.commit=$(COMMIT) -X github.com/liamawhite/navigator/pkg/version.date=$(DATE)

.PHONY: build build-edge build-manager build-navctl build-navctl-dev build-ui build-ui-dev
.PHONY: check clean dirty format generate generate-cli-docs lint local test-unit test-ui test-e2e test-e2e-ci test-e2e-setup

check: generate format lint test-unit test-ui dirty

format:
	licenser apply -r "Navigator Authors"
	gofmt -w .
	cd ui && npm ci && npm run format

lint:
	golangci-lint run --build-tags=lint
	cd ui && npm ci && npm run lint:fix

test-unit: 
	go test -race -tags=test -v ./manager/... ./edge/... ./navctl/... ./pkg/...

test-ui:
	cd ui && npm ci && npm run test

generate: clean
	cd api && buf generate
	cd api && buf generate --template buf.gen.frontend.yaml
	cd api && buf generate --template buf.gen.frontend-docs.yaml
	cd api && buf generate --template buf.gen.backend-docs.yaml
	cd api && buf generate --template buf.gen.types-docs.yaml
	cd ui && npm ci && npm run generate
	go run -tags=docs ./navctl/main.go docs

generate-cli-docs:
	go run ./navctl/main.go docs

dirty:
	git diff --exit-code

clean:
	@rm -rf pkg/api/ ui/src/types/openapi/ ui/src/types/generated/ ui/dist/ bin/ docs/reference/

# Build targets
build-manager:
	@echo "🔨 Building manager binary with version info..."
	@mkdir -p bin
	@go build -ldflags "$(LDFLAGS)" -o bin/manager manager/main.go
	@echo "✅ Manager binary built successfully: bin/manager"

build-edge:
	@echo "🔨 Building edge binary with version info..."
	@mkdir -p bin
	@go build -ldflags "$(LDFLAGS)" -o bin/edge edge/main.go
	@echo "✅ Edge binary built successfully: bin/edge"

build-navctl:
	@echo "🔨 Building navctl binary with version info..."
	@mkdir -p bin
	@go generate ./ui/...
	@go build -ldflags "$(LDFLAGS)" -o bin/navctl navctl/main.go
	@echo "✅ Navctl binary built successfully: bin/navctl"

build-navctl-dev:
	@echo "🔨 Building navctl binary with development UI (dev mode for error details)..."
	@mkdir -p bin
	@cd ui && npm ci && npm run build:dev
	@go build -ldflags "$(LDFLAGS)" -o bin/navctl navctl/main.go
	@echo "✅ Navctl binary built successfully: bin/navctl (with dev UI for debugging)"

build-ui:
	@echo "🔨 Building UI assets..."
	@cd ui && npm ci && npm run build
	@echo "✅ UI assets built successfully: ui/dist/"

build: build-manager build-edge build-navctl build-ui
	@echo "✅ All binaries and assets built successfully"

# Local development target - builds all components and starts them locally
local: build-navctl-dev
	@echo "🚀 Starting Navigator services locally..."
	@./bin/navctl local

# E2E testing targets
test-e2e: build-navctl-dev
	@echo "🧪 Running E2E tests with fresh navctl build..."
	@cd ui && npm ci
	@cd ui && npx playwright install chromium
	@cd ui && npm run e2e

test-e2e-ci: build-navctl-dev
	@echo "🧪 Running E2E tests in CI mode (with demo setup)..."
	@cd ui && npm ci
	@cd ui && sudo npx playwright install-deps chromium
	@cd ui && npx playwright install chromium
	@cd ui && npm run e2e:ci

test-e2e-setup: build-navctl-dev
	@echo "🧪 Running E2E tests with full demo setup and teardown..."
	@echo "🎭 Setting up demo cluster..."
	@./bin/navctl demo start --name navigator-e2e --istio-version 1.25.4
	@echo "🧪 Running E2E tests..."
	@cd ui && npm ci && npx playwright install chromium && E2E_DEMO_NAME=navigator-e2e npm run e2e
	@echo "🧹 Cleaning up demo cluster..."
	@./bin/navctl demo stop --name navigator-e2e

test-e2e-debug: build-navctl-dev
	@echo "🐛 Running E2E tests in debug mode..."
	@cd ui && npm ci
	@cd ui && npx playwright install chromium
	@cd ui && npm run e2e:debug

test-e2e-ui: build-navctl-dev
	@echo "🎭 Running E2E tests with Playwright UI..."
	@cd ui && npm ci
	@cd ui && npx playwright install chromium
	@cd ui && npm run e2e:ui

