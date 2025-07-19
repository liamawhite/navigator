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

.PHONY: build build-edge build-manager build-navctl build-ui check clean demo demo-clean dev dev-backend dev-ui-only dirty fmt generate generate-backend generate-frontend lint test-integration test-unit

check: generate format lint test-unit dirty

format:
	licenser apply -r "Navigator Authors"
	gofmt -w .
	cd ui && npm ci && npm run format

lint:
	golangci-lint run --build-tags=lint
	cd ui && npm ci && npm run lint:fix

test-unit: 
	go test -tags=test -v ./manager/... ./edge/... ./navctl/... ./pkg/...

generate: clean
	cd api && buf generate
	cd api && buf generate --template buf.gen.frontend.yaml
	cd api && buf generate --template buf.gen.frontend-docs.yaml
	cd api && buf generate --template buf.gen.backend-docs.yaml
	cd api && buf generate --template buf.gen.types-docs.yaml
	cd ui && npm ci && npm run generate

dirty:
	git diff --exit-code

clean:
	@rm -rf pkg/api/ ui/src/types/openapi/ ui/src/types/generated/ ui/dist/ bin/

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

build-ui:
	@echo "🔨 Building UI assets..."
	@cd ui && npm ci && npm run build
	@echo "✅ UI assets built successfully: ui/dist/"

build: build-manager build-edge build-navctl build-ui
	@echo "✅ All binaries and assets built successfully"

local:
	rm -rf bin/ && make build-navctl && bin/navctl local

