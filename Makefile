# Use nix develop shell for all commands to ensure dependencies are available
SHELL := nix develop --command bash

.PHONY: fmt test-unit test-integration generate dirty check

check: generate format test-unit dirty

format:
	gofmt -w .

test-unit: 
	go test -v $(shell go list ./... | grep -v /testing/integration)

test-integration:
	go test -v -timeout 15m ./testing/integration/...

generate:
	cd api && buf generate

dirty:
	git diff --exit-code

