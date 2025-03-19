# Makefile for Pararius Property Notification Bot

# Variables
BINARY_NAME=property-bot
GO=go
OUTPUT_DIR=./properties

# Default target
.PHONY: all
all: build

# Get dependencies
.PHONY: deps
deps:
	$(GO) mod download
	$(GO) mod tidy

# Build the binary
.PHONY: build
build: deps
	$(GO) build -o $(BINARY_NAME) *.go

# Run the bot (requires TELEGRAM_TOKEN environment variable)
.PHONY: run
run: build
	./$(BINARY_NAME) -output $(OUTPUT_DIR) -token $(TELEGRAM_TOKEN)

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
	$(GO) clean

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make deps       - Download dependencies"
	@echo "  make build      - Build the binary"
	@echo "  make run        - Run the bot (requires TELEGRAM_TOKEN env var)"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make help       - Show this help message"
	@echo ""
	@echo "Example usage:"
	@echo "  TELEGRAM_TOKEN=your_token make run"
	@echo "  make build && ./$(BINARY_NAME) -output ./properties -token your_token -url custom_url"
