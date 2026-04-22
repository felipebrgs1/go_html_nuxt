.PHONY: dev build install clean generate

# Variables
BINARY_NAME=framework
CMD_PATH=./scripts/framework
BUILD_FLAGS=-ldflags="-s -w"

# Development: clean, build and run CLI in dev mode
dev: clean
	go build -o $(BINARY_NAME) $(CMD_PATH)
	./$(BINARY_NAME) dev

# Production build: generate gonx/routes and compile optimized CLI
build: generate
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Binary generated: ./$(BINARY_NAME)"

# Generate compiled files (.gonx -> gonx/) and routes (gonx/framework_gen/routes.gen.go)
generate:
	@go run scripts/generate.go

# Install CLI in GOPATH/bin
install:
	go install $(BUILD_FLAGS) $(CMD_PATH)
	@echo "Installed in $$(go env GOPATH)/bin/$(BINARY_NAME)"

# Clean generated artifacts
clean:
	@rm -f $(BINARY_NAME)
	@rm -rf gonx/
	@find . -name "*_templ.go" -delete
	@find . -name "styles.css" -path "*/public/*" -delete
	@echo "Cleaned"
