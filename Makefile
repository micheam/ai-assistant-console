.PHONY : clean test install
TARGET = ./bin/aico
ENTRYPOINT = .
SOURCE = $(shell find . -name '*.go')
USER_BIN = $(shell echo $$HOME)/bin
TEST_OPTS = -v -tags e2e

$(TARGET) : test $(SOURCE)
	go build -o $(TARGET) $(ENTRYPOINT)

test :
	go test $(TEST_OPTS) ./...

clean :
	rm -f $(TARGET)

install : clean $(TARGET)
	cp $(TARGET) $(USER_BIN)
