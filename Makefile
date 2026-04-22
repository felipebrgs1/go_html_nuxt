.PHONY: ls build install dev playground clean fmt test-fmt test-lint test-fmt-lint

# Variables
BINARY_NAME=framework
CMD_PATH=./scripts/framework
BUILD_FLAGS=-ldflags="-s -w"

# List available commands
ls:
	@echo "Available commands:"
	@echo "  make build          - Build the framework binary"
	@echo "  make install        - Install framework in GOPATH/bin"
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
