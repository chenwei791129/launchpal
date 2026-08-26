.PHONY: setup test lint build build-debug build-helper install-helper dev dmg clean cloc help
.DEFAULT_GOAL := help

# ANSI color codes for help output
CYAN  := \033[36m
BOLD  := \033[1m
RESET := \033[0m

help: ## Show this help message
	@printf "$(BOLD)Usage:$(RESET) make $(CYAN)<target>$(RESET)\n\n$(BOLD)Targets:$(RESET)\n"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(CYAN)%-13s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Install dependencies
	go mod download
	cd frontend && pnpm install
	brew install create-dmg

test: ## Run tests
	go test -race ./...
	cd frontend && pnpm vitest run
	cd frontend && pnpm nuxi typecheck

lint: ## Run linter
	go tool golangci-lint run ./...
	cd frontend && pnpm eslint .

# Build the privhelper binary. Written to build/bin/launchpal-privhelper so
# the wails build step (which writes build/bin/launchpal.app) can pick it up
# alongside the main binary.
build-helper: ## Build the launchpal-privhelper binary
	mkdir -p build/bin
	CGO_ENABLED=1 go build -trimpath -o build/bin/launchpal-privhelper ./cmd/launchpal-privhelper

# Copy the prebuilt privhelper into the .app bundle next to the main binary.
# Expects build-helper to have run already (so the pinned binary is copied
# verbatim); must run after `wails build` so the bundle directory exists. A
# no-op when the bundle is missing.
install-helper: ## Copy the prebuilt privhelper into the app bundle
	@if [ -d build/bin/launchpal.app/Contents/MacOS ]; then \
		cp build/bin/launchpal-privhelper build/bin/launchpal.app/Contents/MacOS/launchpal-privhelper; \
		chmod 0755 build/bin/launchpal.app/Contents/MacOS/launchpal-privhelper; \
		echo "Installed launchpal-privhelper into app bundle"; \
	else \
		echo "No app bundle found at build/bin/launchpal.app — skipping install-helper"; \
	fi

# The macOS deployment target for the Wails build, and the single source of
# truth for it. It must stay equal to LSMinimumSystemVersion in
# build/darwin/Info.plist and Info.dev.plist — build_metadata_test.go pins the
# plist side, and `otool -l ... | grep -A 4 LC_BUILD_VERSION` is the manual
# check for this side (see .claude/CLAUDE.md "Minimum macOS Version").
#
# Why this has to be set at all: the `desktop,production` tags pull in the Wails
# darwin Objective-C backend, which forces external linking, and under external
# linking clang — not Go — writes LC_BUILD_VERSION. Go's own default (13.0 as of
# Go 1.27) only applies to internally linked binaries, so it does not reach the
# main binary. Wails hardcodes -mmacosx-version-min=10.13 into these two
# variables (pkg/commands/build/base.go), which arm64 clamps to 11.0; its
# UpsertEnv skips its own value when ours already carries the flag, and still
# appends -framework UniformTypeIdentifiers. Leaving the flag out entirely is
# not an option either — clang would then default to the build machine's SDK
# version, making minos differ per machine.
MACOS_MIN_VERSION := 13.0

WAILS_CGO_ENV := CGO_CFLAGS="-mmacosx-version-min=$(MACOS_MIN_VERSION)" \
	CGO_LDFLAGS="-mmacosx-version-min=$(MACOS_MIN_VERSION)"

# VERSION, when set (release builds pass it, e.g. `make build VERSION=v1.6.0`),
# is injected as main.version; empty locally → the binary keeps its "dev"
# default. Kept as a make-time prefix to the ldflags string below.
VERSION_LDFLAG := $(if $(VERSION),-X main.version=$(VERSION) ,)

# The build order is load-bearing: build the helper first, hash it, then inject
# that hash as main.helperPin BEFORE the main binary is linked (via wails build
# ldflags), and finally copy the same helper binary into the bundle so its
# SHA-256 matches the embedded pin. This target is the single source of truth
# for that order — CI calls `make build VERSION=...` rather than re-inlining it.
build: build-helper ## Build production app (includes privhelper)
	HELPER_PIN=$$(shasum -a 256 build/bin/launchpal-privhelper | awk '{print $$1}'); \
	$(WAILS_CGO_ENV) go tool wails build -ldflags "$(VERSION_LDFLAG)-X main.helperPin=$$HELPER_PIN"
	$(MAKE) install-helper

build-debug: build-helper ## Build with devtools enabled
	HELPER_PIN=$$(shasum -a 256 build/bin/launchpal-privhelper | awk '{print $$1}'); \
	$(WAILS_CGO_ENV) go tool wails build -debug -ldflags "$(VERSION_LDFLAG)-X main.helperPin=$$HELPER_PIN"
	$(MAKE) install-helper

dev: build-debug ## Build and run app
	open build/bin/launchpal.app

dmg: build ## Build and package as DMG
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

clean: ## Clean build artifacts
	rm -rf build/bin
	rm -rf frontend/.output
	rm -rf frontend/.nuxt

cloc: ## Count lines of code (respects .gitignore)
	@command -v cloc >/dev/null 2>&1 || { echo "Error: cloc is required. Install with: brew install cloc"; exit 1; }
	cloc --vcs=git
