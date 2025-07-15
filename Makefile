# Use nix develop shell for all commands to ensure dependencies are available
# Disable dirty git tree warnings
export NIX_CONFIG := warn-dirty = false
SHELL := nix develop --command bash

.PHONY: fmt test-unit test-integration generate dirty check dev dev-ui-only dev-backend build build-ui lint demo demo-clean clean

check: generate format lint test-unit dirty

format:
	gofmt -w .
	cd ui && npm ci && npm run format

lint:
	golangci-lint run
	cd ui && npm ci && npm run lint:fix

test-unit: 
	go test -v $(shell go list ./... | grep -v /testing/integration)

test-integration:
	go test -v -timeout 15m ./testing/integration/...

generate:
	cd api && buf generate
	cd ui && npm ci && npm run generate

dirty:
	git diff --exit-code

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
	@if [ -f "./navigator" ]; then \
		echo "🗑️  Removing navigator binary..."; \
		rm -f ./navigator; \
		echo "✅ Navigator binary removed"; \
	fi
	@echo "🎉 Development environment cleaned up!"

# Build targets
build: build-ui
	go build -o navigator cmd/navigator/main.go

build-ui:
	cd ui && npm ci && npm run build

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

