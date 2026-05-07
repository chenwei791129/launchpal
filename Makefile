.PHONY: setup test lint build build-debug build-helper install-helper dev dmg clean cloc help

# Show available commands
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  setup        Install dependencies"
	@echo "  test         Run tests"
	@echo "  lint         Run linter"
	@echo "  build        Build production app (includes privhelper)"
	@echo "  build-debug  Build with devtools enabled"
	@echo "  build-helper Build the launchpal-privhelper binary"
	@echo "  dev          Build and run app"
	@echo "  dmg          Build and package as DMG"
	@echo "  clean        Clean build artifacts"
	@echo "  cloc         Count lines of code (respects .gitignore)"
	@echo "  help         Show this help message"

# Install dependencies
setup:
	go mod download
	cd frontend && pnpm install
	brew install create-dmg

# Run tests
test:
	go test -race ./...
	cd frontend && pnpm vitest run
	cd frontend && pnpm nuxi typecheck

# Run linter
lint:
	go tool golangci-lint run ./...
	cd frontend && pnpm eslint .

# Build the privhelper binary. Written to build/bin/launchpal-privhelper so
# the wails build step (which writes build/bin/launchpal.app) can pick it up
# alongside the main binary.
build-helper:
	mkdir -p build/bin
	CGO_ENABLED=1 go build -trimpath -o build/bin/launchpal-privhelper ./cmd/launchpal-privhelper

# Copy the privhelper into the .app bundle next to the main binary. This must
# run after `wails build` so the bundle directory exists. The install target
# is a no-op when the bundle is missing (used by `make build` below which
# runs this after wails succeeds).
install-helper: build-helper
	@if [ -d build/bin/launchpal.app/Contents/MacOS ]; then \
		cp build/bin/launchpal-privhelper build/bin/launchpal.app/Contents/MacOS/launchpal-privhelper; \
		chmod 0755 build/bin/launchpal.app/Contents/MacOS/launchpal-privhelper; \
		echo "Installed launchpal-privhelper into app bundle"; \
	else \
		echo "No app bundle found at build/bin/launchpal.app — skipping install-helper"; \
	fi

# Build production app
build:
	go tool wails build
	$(MAKE) install-helper

# Build with devtools enabled
build-debug:
	go tool wails build -debug
	$(MAKE) install-helper

# Build and run app
dev: build-debug
	open build/bin/launchpal.app

# Build and package as DMG
dmg: build
	@command -v create-dmg >/dev/null 2>&1 || { echo "Error: create-dmg is required. Install with: brew install create-dmg"; exit 1; }
	rm -f LaunchPal.dmg
	create-dmg \
		--volname "LaunchPal" \
		--background "build/darwin/dmg-background.png" \
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

# Count lines of code, respecting .gitignore via git
cloc:
	@command -v cloc >/dev/null 2>&1 || { echo "Error: cloc is required. Install with: brew install cloc"; exit 1; }
	cloc --vcs=git
