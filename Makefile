# Use nix develop shell for all commands to ensure dependencies are available
SHELL := nix develop --command bash

.PHONY: fmt test-unit test-integration generate dirty check dev dev-ui dev-backend build build-ui lint

check: generate format lint test-unit dirty

format:
	gofmt -w .
	cd ui && npm run format

lint:
	go vet ./...
	gosec -exclude-dir=testing ./...
	golangci-lint run
	cd ui && npm run lint:fix

test-unit: 
	go test -v $(shell go list ./... | grep -v /testing/integration)

test-integration:
	go test -v -timeout 15m ./testing/integration/...

generate:
	cd api && buf generate

dirty:
	git diff --exit-code

# Development targets
dev:
	@echo "🚀 Starting full-stack development with hot reloading..."
	@echo "📱 Frontend will open at http://localhost:5173"
	@echo "🔧 Backend API available at http://localhost:8081"
	cd ui && npm install && npm run dev:air

dev-ui:
	@echo "🎨 Starting frontend development server..."
	cd ui && npm install && npm run dev

dev-backend:
	@echo "⚙️  Starting backend with hot reloading..."
	air


