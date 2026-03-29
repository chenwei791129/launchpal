.PHONY: setup test build build-debug dev clean help

# Show available commands
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  setup        Install dependencies"
	@echo "  test         Run tests"
	@echo "  build        Build production app"
	@echo "  build-debug  Build with devtools enabled"
	@echo "  dev          Build and run app"
	@echo "  clean        Clean build artifacts"
	@echo "  help         Show this help message"

# Install dependencies
setup:
	go mod download
	cd frontend && pnpm install

# Run tests
test:
	go test ./...

# Build production app
build:
	wails build

# Build with devtools enabled
build-debug:
	wails build -debug

# Build and run app
dev: build-debug
	open build/bin/launchpal.app

# Clean build artifacts
clean:
	rm -rf build/bin
	rm -rf frontend/.output
	rm -rf frontend/.nuxt
