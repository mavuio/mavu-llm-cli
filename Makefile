BINARY := llm
TEST_PACKAGES := ./...

.PHONY: build test

build:
	go build -o $(BINARY) .

test:
	go test $(TEST_PACKAGES)
