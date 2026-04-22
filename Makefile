.PHONY: ls build install dev playground clean fmt test-fmt test-lint test-fmt-lint compress dist

# Variables
BINARY_NAME=framework
CMD_PATH=./scripts/framework
BUILD_FLAGS=-trimpath -ldflags="-s -w"

# Environment for static, lightweight binaries
export CGO_ENABLED=0

# List available commands
ls:
	@echo "Available commands:"
	@echo "  make build          - Build the framework binary"
	@echo "  make install        - Install framework in GOPATH/bin"
	@echo "  make compress       - Compress binary with UPX (if available)"
	@echo "  make dist           - Create compressed distribution archive"
	@echo "  make dev            - Run playground in dev mode"
	@echo "  make fmt            - Format all .gonx files in playground"
	@echo "  make clean          - Remove built artifacts"
	@echo "  make test-fmt       - Run fmt on the test bank"
	@echo "  make test-lint      - Run lint on the test bank"
	@echo "  make test-fmt-lint  - Run both fmt and lint on the test bank"

# Build the framework binary
build:
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Framework built: ./$(BINARY_NAME)"

# Install framework in GOPATH/bin
install:
	go install $(BUILD_FLAGS) $(CMD_PATH)
	@echo "Framework installed in $$(go env GOPATH)/bin/$(BINARY_NAME)"

# Helper to run the playground project using the root framework
dev: build
	@cd playground && ../$(BINARY_NAME) dev .

# Clean framework artifacts
clean:
	rm -f $(BINARY_NAME)
	@echo "Cleaned framework"

# Format all .gonx files in the playground project
fmt: build
	@cd playground && ../$(BINARY_NAME) fmt .

# Run fmt on the test bank
test-fmt: build
	@cd tests/fmt-lint && ../../$(BINARY_NAME) fmt .

# Run lint on the test bank
test-lint: build
	@cd tests/fmt-lint && ../../$(BINARY_NAME) lint

# Run both fmt and lint on the test bank
test-fmt-lint: test-fmt test-lint

# Compress binary with UPX if available
compress: build
	@if command -v upx >/dev/null 2>&1; then \
		echo "Compressing with UPX..."; \
		upx --best --lzma -o $(BINARY_NAME)-compressed $(BINARY_NAME); \
		mv $(BINARY_NAME)-compressed $(BINARY_NAME); \
		echo "Binary compressed: ./$(BINARY_NAME)"; \
	else \
		echo "UPX not installed. Install it to compress the binary."; \
		echo "  Debian/Ubuntu: sudo apt install upx"; \
		echo "  macOS: brew install upx"; \
		exit 1; \
	fi

# Create compressed distribution archive
dist: build
	@echo "Creating distribution archives..."
	@cp $(BINARY_NAME) $(BINARY_NAME)-linux-amd64
	@tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	@rm $(BINARY_NAME)-linux-amd64
	@echo "Distribution archive: $(BINARY_NAME)-linux-amd64.tar.gz"
