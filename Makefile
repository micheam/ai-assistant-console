.PHONY : clean test install
SOURCE = $(shell find . -name '*.go')
USER_BIN = $(shell echo $$HOME)/bin
TEST_OPTS = -tags e2e

chat : test ./cmd/chat/main.go $(SOURCE)
	@go build -o ./bin/chat ./cmd/chat

test :
	@go test $(TEST_OPTS) ./...

clean :
	@rm -f ./bin/chat

install : clean chat
	@cp ./bin/chat $(USER_BIN)/chat

protogen: ./proto
	@buf generate --clean
