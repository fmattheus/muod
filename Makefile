.PHONY: all build clean

# Get version from git tag
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Supported platforms
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Default target
all: build

# Build for all platforms
build:
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform#*/}; \
		echo "Building for $$OS/$$ARCH..."; \
		if [ "$$OS" = "windows" ]; then \
			GOOS=$$OS GOARCH=$$ARCH go build -o dist/muod-$(VERSION)-$$OS-$$ARCH.exe -ldflags="-s -w" ./cmd/muod; \
		else \
			GOOS=$$OS GOARCH=$$ARCH go build -o dist/muod-$(VERSION)-$$OS-$$ARCH -ldflags="-s -w" ./cmd/muod; \
		fi; \
	done
	@echo "Build complete! Binaries are in the dist directory:"
	@ls -lh dist/

# Clean build artifacts
clean:
	rm -rf dist/ 