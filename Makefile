# Makefile for the fsmonitor Go project

.PHONY: build test clean  install-mockery

# Binary name
BINARY_NAME=shugosha
PROJECT_NAME=shugosha

# Build directory
BUILD_DIR=./build

# Go commands
GO_BUILD=go build
GO_CLEAN=go clean
GO_TEST=go test
GO_FMT=go fmt
GO_VET=go vet
GO_RUN=go run

# Define where to install binaries
BIN_DIR := $(CURDIR)/bin

# Define the version of mockery to install
MOCKERY_VERSION := v2.10.0

# Run
run:
	cd cmd/shugosha; $(GO_RUN) main.go wire_gen.go

# Build the project
build: 
	$(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(PROJECT_NAME)/main.go

generate: wire install-mockery
	$(BIN_DIR)/mockery --all --dir=./pkg/model --output=./mocks; go generate ./...

# Run tests
test:
	$(GO_TEST) ./...

# Clean build artifacts
clean:
	$(GO_CLEAN)
	rm -rf $(BUILD_DIR)

# Format the code
fmt:
	$(GO_FMT) ./...

# Vet the code
vet:
	$(GO_VET) ./...

wire:
	go install github.com/google/wire/cmd/wire@latest

# Target to install mockery locally
install-mockery:
	@echo "Installing mockery..."
	@mkdir -p $(BIN_DIR)
	@GOBIN=$(BIN_DIR) go install github.com/vektra/mockery/v2@$(MOCKERY_VERSION)

# Help command to display available commands
help:
	@echo "Available commands:"
	@echo "  build  - Build the project binary"
	@echo "  test   - Run tests"
	@echo "  clean  - Clean build artifacts"
	@echo "  fmt    - Format the code"
	@echo "  vet    - Vet the code"
