# --- Configuration ---
BINARY_NAME    := openoutbox-relay
COMPOSE_FILE   := deployments/docker-compose.yaml
MAIN_PACKAGE   := ./cmd/relay/main.go
PRODUCER_PKG   := ./cmd/producer/main.go

.PHONY: all build run producer test clean fmt lint up down setup docs help

all: setup fmt lint build

# ==========================================
# Development & Execution
# ==========================================

# Run the Relay (Uses your local .env file automatically via your Go code)
run:
	go run $(MAIN_PACKAGE)

# Run the Producer to generate traffic
producer:
	go run $(PRODUCER_PKG)

# Build the local binary
build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PACKAGE)

# Clean build artifacts
clean:
	rm -rf bin/
	go clean -testcache

# ==========================================
# Quality & Linting
# ==========================================

# Format code, fix imports, and shorten long lines (100 chars)
fmt:
	goimports -w .
	golines . -w --max-len=100
	go mod tidy

# Run all linters (Runs fmt first to ensure clean diffs)
lint: fmt
	golangci-lint run ./...

# Run all tests with race detection
test:
	go test -v -race ./...

# Open documentation in your browser
docs:
	@echo "Opening pkgsite..."
	pkgsite -open .

# ==========================================
# Infrastructure (Docker)
# ==========================================

# Start all infrastructure
up:
	docker-compose -f $(COMPOSE_FILE) up -d

# Start specific service: make up-kafka
up-%:
	docker-compose -f $(COMPOSE_FILE) up -d $*

# Stop and remove all containers/networks
down:
	docker-compose -f $(COMPOSE_FILE) down

# Stop specific service: make down-postgres
down-%:
	docker-compose -f $(COMPOSE_FILE) stop $*

# Logs for all service: make logs
logs:
	docker-compose -f $(COMPOSE_FILE) logs -f

# Logs for specific service: make logs-relay
logs-%:
	docker-compose -f $(COMPOSE_FILE) logs -f $*

# Check infrastructure status
ps:
	docker-compose -f $(COMPOSE_FILE) ps

# ==========================================
# Tooling Setup
# ==========================================

# Install all required development tools
setup:
	@echo "Installing Go tools..."
	brew install pre-commit || pip install pre-commit
	pre-commit install
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/segmentio/golines@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Done. Make sure \$$GOPATH/bin is in your \$$PATH."
