BIN_NAME = aico
ENTRY_POINT = ./cmd/aico
TARGET = bin/$(BIN_NAME)
SOURCE = $(shell find . -name '*.go')
INSTALL_PATH ?= /usr/local/bin
BIN_DIR = ./bin
USER_BIN = $(shell echo $$HOME)/bin
TEST_OPTS = -tags e2e
.DEFAULT_GOAL := help

.PHONY: help clean install test build

help: ## Describe make targets
	@grep -E '^[a-zA-Z0-9/_-]+\s?:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "%-30s %s\n", $$1, $$2}'

build: $(TARGET) ## Build the aico binary with version info

$(TARGET): $(SOURCE)
	go build -ldflags \
		"-X main.version=$(shell git describe --tags --always --dirty)" \
		-o $(BIN_DIR)/aico \
		$(ENTRY_POINT)

clean: ## Clean up build artifacts
	rm -f $(BIN_DIR)/*

install: $(TARGET) ## Install the binary to INSTALL_PATH (default: /usr/local/bin)
	install -m 755 $(TARGET) $(INSTALL_PATH)/$(BIN_NAME)

test: ## Run tests with specified options
	go vet ./...
	go test $(TEST_OPTS) ./...

