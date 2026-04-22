.PHONY: build install dev playground clean fmt

# Variables
BINARY_NAME=framework
CMD_PATH=./scripts/framework
BUILD_FLAGS=-ldflags="-s -w"

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
