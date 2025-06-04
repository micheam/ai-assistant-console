SOURCE = $(shell find . -name '*.go')
USER_BIN = $(shell echo $$HOME)/bin
TEST_OPTS = -tags e2e

.PHONY: test clean install protogen
.DEFAULT_GOAL := bin/chat

bin/chat: ./cmd/chat/main.go $(SOURCE)
	@go build -ldflags "-X main.version=$(shell git describe --tags --always --dirty)" -o ./bin/chat ./cmd/chat

test:
	@go test $(TEST_OPTS) ./...

clean:
	@rm -f ./bin/chat

install: bin/chat
	@cp ./bin/chat $(USER_BIN)/chat
	@echo "Installed chat to $(USER_BIN)/chat"

protogen: ./proto
	@buf generate --clean
