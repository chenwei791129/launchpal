.PHONY: setup test build build-debug dev clean

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

# Run development mode
dev:
	wails dev

# Clean build artifacts
clean:
	rm -rf build/bin
	rm -rf frontend/.output
	rm -rf frontend/.nuxt
