TARGET = bin/chatgpt
SOURCE = $(shell find . -name '*.go')
TEST_OPTS = -v -tags e2e
.PHONY : clean test install

$(TARGET) : test $(SOURCE)
	go build -o $(TARGET) .

test :
	go test $(TEST_OPTS) ./...

clean :
	rm -f $(TARGET)

install : clean $(TARGET)
	cp $(TARGET) /usr/local/bin
