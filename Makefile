.PHONY : clean test install
TARGET = ./bin/chat
ENTRYPOINT = ./cmd/chat/main.go
SOURCE = $(shell find . -name '*.go')
TEST_OPTS = -v -tags e2e

$(TARGET) : test $(SOURCE)
	go build -o $(TARGET) $(ENTRYPOINT)

test :
	go test $(TEST_OPTS) ./...

clean :
	rm -f $(TARGET)

install : clean $(TARGET)
	cp $(TARGET) /usr/local/bin
