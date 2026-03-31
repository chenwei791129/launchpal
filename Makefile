.PHONY: setup test build build-debug dev dmg clean help

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
	@echo "  dmg          Build and package as DMG"
	@echo "  clean        Clean build artifacts"
	@echo "  help         Show this help message"

# Install dependencies
setup:
	go mod download
	cd frontend && pnpm install
	brew install create-dmg

# Run tests
test:
	go test ./...

# Build production app
build:
	go tool wails build

# Build with devtools enabled
build-debug:
	go tool wails build -debug

# Build and run app
dev: build-debug
	open build/bin/launchpal.app

# Build and package as DMG
dmg: build
	@command -v create-dmg >/dev/null 2>&1 || { echo "Error: create-dmg is required. Install with: brew install create-dmg"; exit 1; }
	rm -f LaunchPal.dmg
	create-dmg \
		--volname "LaunchPal" \
		--window-size 600 400 \
		--icon-size 128 \
		--icon "launchpal.app" 150 185 \
		--app-drop-link 450 185 \
		"LaunchPal.dmg" \
		"build/bin/" \
	|| test $$? -eq 2
	mv LaunchPal.dmg build/bin/LaunchPal.dmg

# Clean build artifacts
clean:
	rm -rf build/bin
	rm -rf frontend/.output
	rm -rf frontend/.nuxt
