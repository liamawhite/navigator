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

.PHONY: fmt test-unit test-integration generate dirty check dev dev-ui-only dev-backend build build-ui lint demo demo-clean clean

check: generate format lint test-unit dirty

format:
	licenser apply -r "Navigator Authors"
	gofmt -w .
	cd ui && npm ci && npm run format

lint:
	golangci-lint run --build-tags=lint
	cd ui && npm ci && npm run lint:fix

test-unit: 
	go test -tags=test -v ./cmd/... ./internal/... ./pkg/...

test-integration:
	go test -tags=integration -v -timeout 15m ./testing/integration/...

generate:
	cd api && buf generate
	cd ui && npm ci && npm run generate

dirty:
	git diff --exit-code

# Build targets
build:
	@echo "🔨 Building navigator binary with version info..."
	@mkdir -p bin
	@VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev"); \
	COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ"); \
	go build -ldflags "-X github.com/liamawhite/navigator/pkg/version.version=$$VERSION -X github.com/liamawhite/navigator/pkg/version.commit=$$COMMIT -X github.com/liamawhite/navigator/pkg/version.date=$$DATE" -o bin/navigator cmd/navigator/main.go
	@echo "✅ Binary built successfully: bin/navigator"

build-ui:
	@echo "🔨 Building UI for production..."
	cd ui && npm ci && npm run build
	@echo "✅ UI built successfully: ui/dist/"

# Development targets
dev:
	@echo "🚀 Starting full-stack development with Kind cluster and hot reloading..."
	@echo "📱 Frontend will be available at http://localhost:5173"
	@echo "🔧 Backend API will be available at http://localhost:8081"
	@echo ""
	@echo "🐳 Setting up demo environment if needed..."
	@if ! kind get clusters 2>/dev/null | grep -q "^navigator-demo$$"; then \
		echo "⚙️  Creating Kind cluster with microservices and Istio..."; \
		go run cmd/navigator/main.go demo --scenario istio-demo --cleanup-on-exit=false; \
	else \
		echo "✅ Kind cluster 'navigator-demo' already exists"; \
	fi
	@echo ""
	@echo "🎬 Starting development environment..."
	./tools/dev-full-stack.sh

clean:
	@echo "🧹 Cleaning up development environment..."
	@if kind get clusters 2>/dev/null | grep -q "^navigator-demo$$"; then \
		echo "🗑️  Deleting Kind cluster 'navigator-demo'..."; \
		kind delete cluster --name navigator-demo; \
		echo "✅ Kind cluster deleted"; \
	else \
		echo "ℹ️  No Kind cluster 'navigator-demo' found to delete"; \
	fi
	@if [ -f "/tmp/navigator-demo-kubeconfig" ]; then \
		echo "🗑️  Removing kubeconfig file..."; \
		rm -f /tmp/navigator-demo-kubeconfig; \
		echo "✅ Kubeconfig file removed"; \
	fi
	@if [ -f "bin/navigator" ]; then \
		echo "🗑️  Removing navigator binary..."; \
		rm -f bin/navigator; \
		echo "✅ Navigator binary removed"; \
	fi
	@if [ -d "./bin" ]; then \
		echo "🗑️  Removing bin directory..."; \
		rm -rf ./bin; \
		echo "✅ Bin directory removed"; \
	fi
	@echo "🎉 Development environment cleaned up!"


# Development targets for UI
dev-ui-only:
	cd ui && npm run dev

dev-backend:
	air

# Demo targets
demo:
	go run cmd/navigator/main.go demo

demo-clean:
	go run cmd/navigator/main.go demo --cleanup-on-exit=true

