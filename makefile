# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
MAIN_PATH=./cmd/main.go
BINARY_NAME=finance
BINARY_UNIX=$(BINARY_NAME)_unix

# Make parameters
.PHONY: all build run clean test deps help

all: build

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	$(GORUN) $(MAIN_PATH)

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Run tests
test:
	$(GOTEST) ./...

# Install dependencies
deps:
	$(GOMOD) download
	$(GOGET) -u github.com/gin-gonic/gin

# Build for Unix/Linux
build-unix:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) $(MAIN_PATH)

# Initialize the project database (if needed)
init-db:
	$(GORUN) ./scripts/init_db.go

# Help command
help:
	@echo "Available commands:"
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make test        - Run tests"
	@echo "  make deps        - Install dependencies"
	@echo "  make build-unix  - Build for Unix/Linux"
	@echo "  make init-db     - Initialize database (if implemented)"
	@echo "  make help        - Show this help message"

# Default to help if no command given
default: help
