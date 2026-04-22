.PHONY: dev build install clean

# Variables
BINARY_NAME=framework
CMD_PATH=../scripts/framework
# No Makefile level cleanup here, let the CLI handle it or the user do it

# Run the dev server on the current directory
dev:
	@go build -o $(BINARY_NAME) $(CMD_PATH)
	./$(BINARY_NAME) dev .

# Clean generated artifacts in the current project
clean:
	@rm -f $(BINARY_NAME)
	@rm -rf gonx/
	@find . -name "*_templ.go" -delete
	@find . -name "styles.css" -path "*/public/*" -delete 2>/dev/null || true
	@echo "Cleaned"
