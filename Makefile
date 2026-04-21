.PHONY: dev build install clean

# Variáveis
BINARY_NAME=framework
CMD_PATH=./cmd/framework
BUILD_FLAGS=-ldflags="-s -w"

# Desenvolvimento: compila e executa o CLI em modo dev
dev:
	go build -o $(BINARY_NAME) $(CMD_PATH)
	./$(BINARY_NAME) dev

# Build de produção: compila o CLI otimizado
build:
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) $(CMD_PATH)
	@echo "✅ Binário gerado: ./$(BINARY_NAME)"

# Instala o CLI no GOPATH/bin
install:
	go install $(BUILD_FLAGS) $(CMD_PATH)
	@echo "✅ Instalado em $$(go env GOPATH)/bin/$(BINARY_NAME)"

# Limpa artefatos gerados
clean:
	rm -f $(BINARY_NAME)
	rm -rf .framework/
	find . -name "*_templ.go" -delete
	find . -name "styles.css" -path "*/public/*" -delete
	@echo "🧹 Limpo"
