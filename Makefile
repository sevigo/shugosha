# Makefile for the fsmonitor Go project

.PHONY: build test clean

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

# Run
run:
	$(GO_RUN) cmd/shugosha/main.go

# Build the project
build: 
	$(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(PROJECT_NAME)/main.go

generate:
	$(GO_RUN) generate ./...

# Run tests
test: generate
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

# Help command to display available commands
help:
	@echo "Available commands:"
	@echo "  build  - Build the project binary"
	@echo "  test   - Run tests"
	@echo "  clean  - Clean build artifacts"
	@echo "  fmt    - Format the code"
	@echo "  vet    - Vet the code"
