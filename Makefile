# Makefile for Unisphere Application

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOGENERATE=$(GOCMD) generate # Keep generate for potential future use, but swag uses init
GORUN=$(GOCMD) run
GOINSTALL=$(GOCMD) install

# Main package path and directory
MAIN_PATH=cmd/api/main.go
MAIN_DIR=cmd/api

# Binary name
BINARY_NAME=unisphere-api

# Swag command
# Use the default path where 'go install' places binaries ($HOME/go/bin).
# If your Go binaries are in a different location (e.g., custom GOPATH/bin or GOBIN),
# adjust this path accordingly. You might try: SWAG_CMD=$(shell go env GOPATH)/bin/swag
SWAG_CMD=$(HOME)/go/bin/swag
SWAG_PKG=github.com/swaggo/swag/cmd/swag@latest

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building the application..."
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_NAME)"

# Run the application
.PHONY: run
run:
	@echo "Running the application..."
	$(GORUN) $(MAIN_PATH)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Clean up binary
.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	# Optionally remove generated docs
	# rm -rf $(MAIN_DIR)/docs

# Install swag CLI tool if not already installed
.PHONY: install-swag
install-swag:
	@echo "Ensuring swag CLI is installed..."
	$(GOINSTALL) $(SWAG_PKG)
	@echo "Swag CLI should be available now (check $(HOME)/go/bin or $$GOPATH/bin)."

# Generate swagger docs using swag init
# Depends on install-swag to make sure the command exists
.PHONY: swagger
swagger: install-swag
	@echo "Generating Swagger documentation using $(SWAG_CMD)..."
	$(SWAG_CMD) init -g cmd/api/main.go -d . -o docs --parseDependency
	@echo "Swagger files generated in docs directory"

# All-in-one command to regenerate swagger and run the application
.PHONY: dev
dev: swagger run

# Get dependencies
.PHONY: deps
deps:
	@echo "Getting dependencies..."
	$(GOGET) -v ./...

# Help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make build        Build the application binary ($(BINARY_NAME))"
	@echo "  make run          Run the application (using go run)"
	@echo "  make test         Run tests"
	@echo "  make clean        Remove the built binary and potentially generated files"
	@echo "  make install-swag Install/update the swag CLI tool"
	@echo "  make swagger      Generate swagger documentation using swag init (installs swag if needed)"
	@echo "  make dev          Generate swagger docs and then run the application"
	@echo "  make deps         Download Go module dependencies"
	@echo "  make help         Show this help message"

