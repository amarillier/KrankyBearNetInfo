.PHONY: build all clean macos windows linux test help

# Default target (skip windows if cross-compilation not available)
all: macos linux windows-optional

# Build for current platform
build:
	go build -o netinfo -ldflags="-w -s" -trimpath

# Build for macos
macos:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -trimpath -o bin/netinfo-macos-amd64
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -trimpath -o bin/netinfo-macos-arm64

# Build for Windows (requires Windows SDK/headers - only works on Windows or with cross-compilation tools)
windows:
	@echo "Building Windows binaries (requires Windows SDK or cross-compilation tools)..."
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -trimpath -o bin/netinfo-windows-amd64.exe || echo "Warning: Windows amd64 build failed (may require Windows SDK)"
	@CGO_ENABLED=1 GOOS=windows GOARCH=386 go build -ldflags="-w -s" -trimpath -o bin/netinfo-windows-386.exe || echo "Warning: Windows 386 build failed (may require Windows SDK)"

# Try to build Windows but don't fail if it doesn't work
windows-optional:
	@echo "Attempting Windows build (may skip if Windows SDK not available)..."
	@-$(MAKE) windows || echo "Windows build skipped (not available on this platform)"

# Build for Linux
linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -trimpath -o bin/netinfo-linux-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -trimpath -o bin/netinfo-linux-arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags="-w -s" -trimpath -o bin/netinfo-linux-386

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f netinfo netinfo.exe

# Run tests
test:
	go test ./...

# Help target
help:
	@echo "Available targets:"
	@echo "  make build      - Build for current platform"
	@echo "  make macos     - Build for MacOS (amd64 and arm64)"
	@echo "  make linux      - Build for Linux (amd64, arm64, and 386)"
	@echo "  make windows    - Build for Windows (amd64 and 386) - requires Windows SDK"
	@echo "  make all        - Build for all available platforms (Windows optional)"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make help       - Show this help message"
	@echo ""
	@echo "Note: Windows builds require Windows SDK/headers and typically must be"
	@echo "      built on Windows or with proper cross-compilation tools (MinGW, etc.)"

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
