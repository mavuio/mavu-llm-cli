TEST_PACKAGES := ./...

.PHONY: test

test:
	go test $(TEST_PACKAGES)
